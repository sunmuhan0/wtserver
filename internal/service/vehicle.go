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
	IsHidden       bool     `json:"isHidden"`
	Crew           int      `json:"crew"`
	Mass           float64  `json:"mass"`
	EnginePower    int      `json:"enginePower"`
	MaxSpeed       float64  `json:"maxSpeed"`
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
		IsEvent:   v.IsEvent,
		IsGift:    v.IsGift,
		IsNormal:  v.IsNormal,
		IsHidden:  !v.IsNormal,
		Crew:      v.Crew,
		Mass:      v.Mass,
		EnginePower: v.EnginePower,
		MaxSpeed:  v.MaxSpeed,
	}
}

func ListVehicles(country, vtype, search string, offset, limit int) ([]model.Vehicle, int) {
	cache, err := loadSSVehicleCache()
	if err != nil {
		return nil, 0
	}
	lcountry := strings.ToLower(country)
	lvtype := strings.ToLower(vtype)
	lsearch := strings.ToLower(search)

	var all []model.Vehicle
	for _, v := range cache {
		if lcountry != "" && !strings.Contains(strings.ToLower(v.Country), lcountry) {
			continue
		}
		if lvtype != "" && !strings.Contains(strings.ToLower(v.UnitClass), lvtype) {
			continue
		}
		if lsearch != "" && !strings.Contains(strings.ToLower(v.Name), lsearch) {
			continue
		}
		all = append(all, *ssVehicleToModel(v))
	}
	total := len(all)
	if offset >= total {
		return nil, total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return all[offset:end], total
}

func GetCountries() []string {
	cache, err := loadSSVehicleCache()
	if err != nil {
		return nil
	}
	seen := make(map[string]bool)
	var countries []string
	for _, v := range cache {
		c := strings.TrimPrefix(v.Country, "country_")
		if c != "" && !seen[c] {
			seen[c] = true
			countries = append(countries, c)
		}
	}
	return countries
}

func GetTypes() []string {
	cache, err := loadSSVehicleCache()
	if err != nil {
		return nil
	}
	seen := make(map[string]bool)
	var types []string
	for _, v := range cache {
		t := v.UnitClass
		if t != "" && !seen[t] {
			seen[t] = true
			types = append(types, t)
		}
	}
	return types
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

var articleRe = regexp.MustCompile(`(?s)<section class="section section--narrow article">(.*?)</section>`)
var tagRe = regexp.MustCompile(`<[^>]+>`)

func GetNewsDetail(url string) (*model.NewsDetail, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	titleMatch := regexp.MustCompile(`<div class="content__title">\s*([^<]+)`).FindStringSubmatch(html)
	title := ""
	if titleMatch != nil {
		title = strings.TrimSpace(titleMatch[1])
	}

	articleMatch := articleRe.FindStringSubmatch(html)
	if articleMatch == nil {
		return nil, fmt.Errorf("article content not found")
	}
	articleHTML := articleMatch[1]

	headingRe2 := regexp.MustCompile(`(?s)<h2[^>]*>(.*?)</h2>`)
	headingRe3 := regexp.MustCompile(`(?s)<h3[^>]*>(.*?)</h3>`)
	pRe := regexp.MustCompile(`(?s)<p>(.*?)</p>`)
	imgRe := regexp.MustCompile(`<img[^>]+src="([^"]+)"`)

	var blocks []model.ContentBlock

	blocksRe := regexp.MustCompile(`(?s)(<h[23][^>]*>.*?</h[23]>|<p>.*?</p>|<img[^>]+>)`)
	matches := blocksRe.FindAllString(articleHTML, -1)

	for _, m := range matches {
		if h := headingRe2.FindStringSubmatch(m); h != nil {
			text := stripTags(h[1])
			text = strings.TrimSpace(text)
			if text != "" {
				blocks = append(blocks, model.ContentBlock{Type: "heading", Level: 2, Text: text})
			}
		} else if h := headingRe3.FindStringSubmatch(m); h != nil {
			text := stripTags(h[1])
			text = strings.TrimSpace(text)
			if text != "" {
				blocks = append(blocks, model.ContentBlock{Type: "heading", Level: 3, Text: text})
			}
		} else if p := pRe.FindStringSubmatch(m); p != nil {
			text := stripTags(p[1])
			text = strings.TrimSpace(text)
			if text != "" {
				blocks = append(blocks, model.ContentBlock{Type: "text", Text: text})
			}
		} else if img := imgRe.FindStringSubmatch(m); img != nil {
			u := img[1]
			if strings.HasPrefix(u, "//") {
				u = "https:" + u
			}
			blocks = append(blocks, model.ContentBlock{Type: "image", URL: u})
		}
	}

	if len(blocks) == 0 {
		return nil, fmt.Errorf("no content blocks found")
	}

	return &model.NewsDetail{Title: title, Content: blocks}, nil
}

func stripTags(s string) string {
	return strings.TrimSpace(tagRe.ReplaceAllString(s, ""))
}
