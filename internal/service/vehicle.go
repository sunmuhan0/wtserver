package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/warthunder/assistant/internal/model"
)

type ssVehicleInfo struct {
	Name           string   `json:"name"`
	UnitClass      string   `json:"unitClass"`
	UnitMoveType   string   `json:"unitMoveType"`
	Rank           int      `json:"rank"`
	BR             float64  `json:"battleRatingArcade"`
	BRHist         *float64 `json:"battleRatingHistorical"`
	BRSim          *float64 `json:"battleRatingSimulation"`
	Country        string   `json:"country"`
	IsPrem         bool     `json:"isPrem"`
	IsEvent        bool     `json:"isEvent"`
	IsGift         bool     `json:"isGift"`
	IsNormal       bool     `json:"isNormal"`
}

var (
	ssVehicleCache     map[string]ssVehicleInfo
	ssVehicleCacheMu   sync.RWMutex
	ssVehicleCacheTime time.Time
	ssVehicleCacheTTL  = 60 * time.Minute
)

func loadSSVehicleCache() (map[string]ssVehicleInfo, error) {
	ssVehicleCacheMu.RLock()
	if ssVehicleCache != nil && time.Since(ssVehicleCacheTime) < ssVehicleCacheTTL {
		defer ssVehicleCacheMu.RUnlock()
		return ssVehicleCache, nil
	}
	ssVehicleCacheMu.RUnlock()

	ssVehicleCacheMu.Lock()
	defer ssVehicleCacheMu.Unlock()

	if ssVehicleCache != nil && time.Since(ssVehicleCacheTime) < ssVehicleCacheTTL {
		return ssVehicleCache, nil
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post("https://statshark.net/api/misc/getVehicleinfo",
		"application/json",
		bytes.NewReader([]byte("{}")))
	if err != nil {
		return nil, fmt.Errorf("vehicle api unreachable: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var vehicles map[string]ssVehicleInfo
	if err := json.Unmarshal(body, &vehicles); err != nil {
		return nil, fmt.Errorf("parse vehicle data: %w", err)
	}

	ssVehicleCache = vehicles
	ssVehicleCacheTime = time.Now()
	log.Printf("[vehicle] cached %d vehicles from statshark", len(vehicles))
	return vehicles, nil
}

func GetVehicle(name string) (*model.Vehicle, error) {
	cache, err := loadSSVehicleCache()
	if err != nil {
		return nil, err
	}

	lname := strings.ToLower(name)

	if v, ok := cache[lname]; ok {
		return ssVehicleToModel(v), nil
	}

	var bestID string
	var bestV ssVehicleInfo
	bestSuffix := -1

	for id, v := range cache {
		if id == lname {
			return ssVehicleToModel(v), nil
		}
		idx := strings.Index(id, lname)
		if idx < 0 {
			continue
		}
		suffixLen := len(id) - idx - len(lname)
		if bestSuffix < 0 || suffixLen < bestSuffix || (suffixLen == bestSuffix && len(id) < len(bestID)) {
			bestID = id
			bestV = v
			bestSuffix = suffixLen
		}
	}
	if bestSuffix >= 0 {
		_ = bestID
		return ssVehicleToModel(bestV), nil
	}

	for _, v := range cache {
		if strings.Contains(strings.ToLower(v.Name), lname) {
			return ssVehicleToModel(v), nil
		}
	}

	return nil, fmt.Errorf("vehicle not found: %s", name)
}

func ssVehicleToModel(v ssVehicleInfo) *model.Vehicle {
	country := strings.TrimPrefix(v.Country, "country_")
	brHist := 0.0
	if v.BRHist != nil {
		brHist = *v.BRHist
	}
	brSim := 0.0
	if v.BRSim != nil {
		brSim = *v.BRSim
	}
	return &model.Vehicle{
		Name:      v.Name,
		Country:   country,
		Type:      v.UnitClass,
		Rank:      v.Rank,
		BR:        fmt.Sprintf("%.1f/%.1f/%.1f", v.BR, brHist, brSim),
		IsPremium: v.IsPrem,
	}
}

func GetSquadron(name string) (*model.Squadron, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get("https://warthunder.com/en/clan/search/?name=" + name)
	if err != nil {
		return nil, fmt.Errorf("squadron search unavailable: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	if strings.Contains(html, `data-tag="`) {
		tagRe := regexp.MustCompile(`(?i)data-tag="([^"]+)"`)
		nameRe := regexp.MustCompile(`(?i)class="clan-search__name"[^>]*>([^<]+)`)
		memberRe := regexp.MustCompile(`(?i)class="clan-search__members"[^>]*>([^<]+)`)

		tags := tagRe.FindStringSubmatch(html)
		names := nameRe.FindStringSubmatch(html)
		members := memberRe.FindStringSubmatch(html)

		s := &model.Squadron{}
		if len(tags) >= 2 {
			s.Tag = strings.TrimSpace(tags[1])
		}
		if len(names) >= 2 {
			s.Name = strings.TrimSpace(names[1])
		}
		if len(members) >= 2 {
			fmt.Sscanf(strings.TrimSpace(members[1]), "%d", &s.Members)
		}
		if s.Name != "" || s.Tag != "" {
			return s, nil
		}
	}

	re := regexp.MustCompile(`/en/clan/(\d+)-([^/"]+)`)
	matches := re.FindStringSubmatch(html)
	if matches != nil {
		return &model.Squadron{
			Tag:  matches[2],
			Name: strings.ReplaceAll(matches[2], "-", " "),
		}, nil
	}

	return nil, fmt.Errorf("squadron search unavailable: site requires JavaScript rendering")
}

var newsLinkRe = regexp.MustCompile(`/(en|zh)/news/(\d+)-`)
var newsWidgetRe = regexp.MustCompile(`(?s)class="widget__link"\s*href="(/(?:en|zh)/news/[^"]+)".*?data-src="([^"]+)".*?class="widget__title">\s*([^<]+)`)

func GetNews(lang string) ([]model.NewsItem, error) {
	if lang != "en" && lang != "zh" {
		lang = "zh"
	}
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", "https://warthunder.com/"+lang+"/news", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("news fetch failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	matches := newsWidgetRe.FindAllStringSubmatch(html, 10)
	if len(matches) == 0 {
		return nil, fmt.Errorf("news unavailable: could not parse news page")
	}

	var items []model.NewsItem
	seen := make(map[string]bool)
	for _, m := range matches {
		path := m[1]
		if seen[path] {
			continue
		}
		seen[path] = true

		title := strings.TrimSpace(m[3])
		if title == "" {
			continue
		}

		image := m[2]
		if strings.HasPrefix(image, "//") {
			image = "https:" + image
		}

		slug := newsLinkRe.FindStringSubmatch(path)
		date := ""
		if slug != nil {
			date = slug[2]
		}

		items = append(items, model.NewsItem{
			Title: title,
			URL:   "https://warthunder.com" + path,
			Image: image,
			Date:  date,
		})
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("news unavailable: no items found")
	}
	return items, nil
}
