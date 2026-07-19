package thunderskill

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type PlayerProfile struct {
	Nickname  string       `json:"nickname"`
	Clan      string       `json:"clan"`
	VKLevel   string       `json:"vk_level"`
	KDOverall float64      `json:"kd_overall"`
	WROverall float64      `json:"wr_overall"`
	Modes     []PlayerMode `json:"modes"`
}

type PlayerMode struct {
	Mode   string  `json:"mode"`
	Games  int     `json:"games"`
	WR     float64 `json:"wr"`
	KD     float64 `json:"kd"`
	KPB    float64 `json:"kpb"`
	Skill  float64 `json:"skill"`
	EFF    float64 `json:"eff"`
	Level  string  `json:"level"`
}

type PlayerVehicle struct {
	VehicleName string  `json:"vehicle_name"`
	VehicleSlug string  `json:"vehicle_slug"`
	Country     string  `json:"country"`
	BR          float64 `json:"br"`
	Battles     int     `json:"battles"`
	WR          float64 `json:"wr"`
	KD          float64 `json:"kd"`
	Efficiency  float64 `json:"efficiency"`
}

const playerCacheTTL = 5 * time.Minute

type playerCacheEntry struct {
	profile   *PlayerProfile
	vehicles  []PlayerVehicle
	expiresAt time.Time
}

var (
	playerCache   = make(map[string]*playerCacheEntry)
	playerCacheMu = make(chan struct{}, 1)
)

func init() {
	playerCacheMu <- struct{}{}
}

func (c *Client) FetchPlayerProfile(nickname string) (*PlayerProfile, error) {
	return c.fetchPlayerData(nickname, false)
}

func (c *Client) FetchPlayerVehicles(nickname string) ([]PlayerVehicle, error) {
	_, err := c.fetchPlayerData(nickname, true)
	if err != nil {
		return nil, err
	}
	playerCacheMu <- struct{}{}
	entry := playerCache[nickname]
	<-playerCacheMu
	if entry == nil {
		return nil, fmt.Errorf("cache miss after fetch")
	}
	return entry.vehicles, nil
}

func (c *Client) fetchPlayerData(nickname string, needVehicles bool) (*PlayerProfile, error) {
	playerCacheMu <- struct{}{}
	if entry, ok := playerCache[nickname]; ok && time.Now().Before(entry.expiresAt) {
		profile := entry.profile
		playerCacheMu <- struct{}{}
		return profile, nil
	}
	playerCacheMu <- struct{}{}

	req, err := c.newRequest(fmt.Sprintf("%s/en/stat/%s/", BaseURL, url.PathEscape(nickname)))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("player %s not found on thunderskill", nickname)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("thunderskill returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	profile := parsePlayerProfile(doc)
	if profile == nil {
		return nil, fmt.Errorf("failed to parse profile from thunderskill page")
	}
	profile.Nickname = nickname

	var vehicles []PlayerVehicle
	if needVehicles {
		vehicles = parsePlayerVehicles(doc)
	}

	playerCacheMu <- struct{}{}
	playerCache[nickname] = &playerCacheEntry{
		profile:   profile,
		vehicles:  vehicles,
		expiresAt: time.Now().Add(playerCacheTTL),
	}
	playerCacheMu <- struct{}{}

	return profile, nil
}

func parsePlayerProfile(doc *goquery.Document) *PlayerProfile {
	profile := &PlayerProfile{}

	profile.Clan = strings.TrimSpace(doc.Find(".clan a").First().Text())
	doc.Find(".profile .info-row").Each(func(_ int, row *goquery.Selection) {
		label := strings.TrimSpace(row.Find(".label").Text())
		value := strings.TrimSpace(row.Find(".value").Text())
		switch strings.ToLower(label) {
		case "vk level":
			profile.VKLevel = value
		}
	})

	profile.KDOverall = parseStatValue(doc, "total.kd")
	profile.WROverall = parseStatValue(doc, "total.wr")

	doc.Find(".mode-stat-box").Each(func(_ int, box *goquery.Selection) {
		mode := PlayerMode{}
		mode.Mode = strings.TrimSpace(box.Find(".mode-name").Text())
		mode.Games = parseIntFromText(box.Find(".games-count").Text())
		mode.WR = parseStatBoxValue(box, "wr")
		mode.KD = parseStatBoxValue(box, "kd")
		mode.KPB = parseStatBoxValue(box, "kpb")
		mode.Skill = parseStatBoxValue(box, "skill")
		mode.EFF = parseStatBoxValue(box, "eff")
		mode.Level = strings.TrimSpace(box.Find(".level-value").Text())
		profile.Modes = append(profile.Modes, mode)
	})

	return profile
}

func parsePlayerVehicles(doc *goquery.Document) []PlayerVehicle {
	var vehicles []PlayerVehicle
	doc.Find(".vehicle-card, .vehicle-row, [data-vehicle-slug]").Each(func(_ int, sel *goquery.Selection) {
		v := PlayerVehicle{}
		v.VehicleSlug, _ = sel.Attr("data-vehicle-slug")
		v.VehicleName = strings.TrimSpace(sel.Find(".vehicle-name, .name").First().Text())
		v.Country = strings.TrimSpace(sel.Find(".country").First().Text())
		v.BR = parseStatValue(sel, "br")
		v.Battles = int(parseStatValue(sel, "battles"))
		v.WR = parseStatValue(sel, "wr")
		v.KD = parseStatValue(sel, "kd")
		v.Efficiency = parseStatValue(sel, "eff")
		if v.VehicleSlug != "" {
			vehicles = append(vehicles, v)
		}
	})
	return vehicles
}

func parseStatValue(doc interface{}, key string) float64 {
	var selector string
	switch key {
	case "total.kd":
		selector = ".stat-kd .value, [data-stat='kd']"
	case "total.wr":
		selector = ".stat-wr .value, [data-stat='wr']"
	default:
		selector = fmt.Sprintf("[data-stat='%s'] .value, .stat-%s .value", key, key)
	}

	switch d := doc.(type) {
	case *goquery.Document:
		text := strings.TrimSpace(d.Find(selector).First().Text())
		re := regexp.MustCompile(`[0-9.]+`)
		m := re.FindString(text)
		if m == "" {
			return 0
		}
		var v float64
		fmt.Sscanf(m, "%f", &v)
		return v
	case *goquery.Selection:
		text := strings.TrimSpace(d.Find(selector).First().Text())
		re := regexp.MustCompile(`[0-9.]+`)
		m := re.FindString(text)
		if m == "" {
			return 0
		}
		var v float64
		fmt.Sscanf(m, "%f", &v)
		return v
	}
	return 0
}

func parseStatBoxValue(box *goquery.Selection, key string) float64 {
	sel := box.Find(fmt.Sprintf(".stat-%s .value", key))
	if sel.Length() == 0 {
		sel = box.Find(fmt.Sprintf("[data-stat='%s']", key))
	}
	text := strings.TrimSpace(sel.First().Text())
	re := regexp.MustCompile(`[0-9.]+`)
	m := re.FindString(text)
	if m == "" {
		return 0
	}
	var v float64
	fmt.Sscanf(m, "%f", &v)
	return v
}

func parseIntFromText(text string) int {
	re := regexp.MustCompile(`[0-9]+`)
	m := re.FindString(text)
	if m == "" {
		return 0
	}
	var v int
	fmt.Sscanf(m, "%d", &v)
	return v
}

func (c *Client) FetchPlayerExportJSON(nickname string) (map[string]any, error) {
	exportURL := fmt.Sprintf("%s/en/stat/%s/export/json/", BaseURL, url.PathEscape(nickname))
	req, err := c.newRequest(exportURL)
	if err != nil {
		return nil, fmt.Errorf("create export request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("export http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("export for %s not found", nickname)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("export returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read export body: %w", err)
	}

	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("parse export json: %w", err)
	}
	return data, nil
}
