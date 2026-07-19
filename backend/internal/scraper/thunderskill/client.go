package thunderskill

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	BaseURL   = "https://thunderskill.com"
	UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

var modeMap = map[string]string{
	"a": "arcade",
	"r": "realistic",
	"s": "simulator",
}

type Client struct {
	HTTPClient *http.Client
}

func NewClient() *Client {
	return &Client{
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) newRequest(rawURL string) (*http.Request, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "application/json, text/html, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	return req, nil
}

type loadMoreResponse struct {
	Entries []loadMoreEntry `json:"entries"`
}

type loadMoreEntry struct {
	VehicleID   any    `json:"vehicleId"`
	VehicleName string `json:"vehicleName"`
	ObjectCode  string `json:"objectCode"`
	VehicleURL  string `json:"vehicleUrl"`
	TypeLabel   string `json:"typeLabel"`
	ObjectType  any    `json:"objectType"`
	RoleName    string `json:"roleName"`
	RoleCode    any    `json:"roleCode"`
	Country     string `json:"country"`
	Mode        string `json:"mode"`
	RankValue   any    `json:"rankValue"`
	BattleCount any    `json:"battleCount"`
	WinRate     any    `json:"winrate"`
	Efficiency  any    `json:"efficiency"`
	Search      string `json:"search"`
	Pic         string `json:"pic"`
}

type VehicleIndexEntry struct {
	VehicleID    int     `json:"vehicle_id"`
	VehicleSlug  string  `json:"vehicle_slug"`
	VehicleName  string  `json:"vehicle_name"`
	VehicleURL   string  `json:"vehicle_url"`
	Country      string  `json:"country"`
	Rank         int     `json:"rank"`
	BattleCount  int     `json:"battle_count"`
	WinRate      float64 `json:"win_rate"`
	Efficiency   float64 `json:"efficiency"`
}

type VehicleDetail struct {
	VehicleSlug  string           `json:"vehicle_slug"`
	Country      string           `json:"country"`
	VehicleType  string           `json:"vehicle_type"`
	Rank         int              `json:"rank"`
	ArcadeBR     float64          `json:"arcade_br"`
	RealisticBR  float64          `json:"realistic_br"`
	SimulatorBR  float64          `json:"simulator_br"`
	IsPremium    bool             `json:"is_premium"`
	IsSquadron   bool             `json:"is_squadron"`
	DailyStats   []VehicleDailyStat `json:"daily_stats"`
}

type VehicleDailyStat struct {
	Date          string  `json:"date"`
	Mode          string  `json:"mode"`
	KillsPerDeath float64 `json:"kills_per_death"`
	KillsPerBattle float64 `json:"kills_per_battle"`
	WinRate       float64 `json:"win_rate"`
	Efficiency    float64 `json:"efficiency"`
}

func (c *Client) FetchVehicleIndex(mode string, vehicleType, limit, maxPages int) ([]VehicleIndexEntry, error) {
	if mode == "" {
		mode = "R"
	}
	if limit <= 0 {
		limit = 100
	}
	if maxPages <= 0 {
		maxPages = 5
	}

	seen := make(map[string]bool)
	var entries []VehicleIndexEntry

	for page := 0; page < maxPages; page++ {
		offset := len(entries)
		params := url.Values{
			"mode":   {mode},
			"offset": {strconv.Itoa(offset)},
			"limit":  {strconv.Itoa(limit)},
			"layout": {"table"},
			"type":   {strconv.Itoa(vehicleType)},
		}

		req, err := c.newRequest(BaseURL + "/en/vehicles/load-more?" + params.Encode())
		if err != nil {
			return entries, fmt.Errorf("page %d request: %w", page+1, err)
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return entries, fmt.Errorf("page %d: %w", page+1, err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return entries, fmt.Errorf("page %d read: %w", page+1, err)
		}

		var lr loadMoreResponse
		if err := json.Unmarshal(body, &lr); err != nil {
			return entries, fmt.Errorf("page %d decode: %w", page+1, err)
		}

		if len(lr.Entries) == 0 {
			break
		}

		for _, e := range lr.Entries {
			slug := e.ObjectCode
			if slug == "" && e.VehicleURL != "" {
				slug = strings.TrimSuffix(strings.TrimRight(e.VehicleURL, "/"), "/")
				if idx := strings.LastIndex(slug, "/"); idx >= 0 {
					slug = slug[idx+1:]
				}
			}
			if slug == "" || seen[slug] {
				continue
			}
			seen[slug] = true

			country := strings.TrimPrefix(e.Country, "country_")
			country = strings.ToUpper(country)

			entries = append(entries, VehicleIndexEntry{
				VehicleID:   toInt(e.VehicleID),
				VehicleSlug: slug,
				VehicleName: e.VehicleName,
				VehicleURL:  toAbsURL(e.VehicleURL),
				Country:     country,
				Rank:        toInt(e.RankValue),
				BattleCount: toInt(e.BattleCount),
				WinRate:     toFloat(e.WinRate),
				Efficiency:  toFloat(e.Efficiency),
			})
		}

		if len(lr.Entries) < limit {
			break
		}

		time.Sleep(time.Duration(500+page*50) * time.Millisecond)
	}

	return entries, nil
}

func (c *Client) FetchVehicleDetail(slug string) (*VehicleDetail, error) {
	vehicleURL := fmt.Sprintf("%s/en/stat/%s/", BaseURL, slug)
	req, err := c.newRequest(vehicleURL)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status %d for %s", resp.StatusCode, slug)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}

	return parseDetailPage(doc, slug), nil
}

func parseDetailPage(doc *goquery.Document, slug string) *VehicleDetail {
	pageText := doc.Find("body").Text()

	detail := &VehicleDetail{
		VehicleSlug: slug,
		Country:     ExtractBetween(pageText, "Country", "Vehicle type"),
		VehicleType: ExtractBetween(pageText, "Vehicle type", "Rank"),
		Rank:        toInt(ExtractBetween(pageText, "Rank", "Battle rating")),
		ArcadeBR:    toFloat(ExtractBetween(pageText, "Arcade mode", "Realistic mode")),
		RealisticBR: toFloat(ExtractBetween(pageText, "Realistic mode", "Simulator mode")),
		SimulatorBR: toFloat(ExtractBetween(pageText, "Simulator mode", "Premium Vehicle:")),
		IsPremium:   strings.Contains(strings.ToLower(ExtractBetween(pageText, "Premium Vehicle:", "Squadron Vehicle:")), "yes"),
		IsSquadron:  strings.Contains(strings.ToLower(ExtractBetween(pageText, "Squadron Vehicle:", "Pack Vehicle:")), "yes"),
	}

	detail.DailyStats = parseChartCanvases(doc)
	return detail
}

func parseChartCanvases(doc *goquery.Document) []VehicleDailyStat {
	var stats []VehicleDailyStat
	seen := make(map[string]bool)

	doc.Find("canvas[data-symfony--ux-chartjs--chart-view-value]").Each(func(_ int, sel *goquery.Selection) {
		raw, exists := sel.Attr("data-symfony--ux-chartjs--chart-view-value")
		if !exists || raw == "" {
			return
		}

		raw = strings.NewReplacer("&#x2F;", "/", "&quot;", "\"", "&#x27;", "'").Replace(raw)

		var chartData struct {
			Data struct {
				Labels   []string `json:"labels"`
				Datasets []struct {
					Label string    `json:"label"`
					Data  []float64 `json:"data"`
				} `json:"datasets"`
			} `json:"data"`
		}

		if err := json.Unmarshal([]byte(raw), &chartData); err != nil || len(chartData.Data.Datasets) == 0 {
			return
		}

		modeCode := "r"
		parent := sel.ParentsFiltered(".tab-pane")
		parentID, _ := parent.Attr("id")
		re := regexp.MustCompile(`vehicle-mode-metric-([a-z]+)-`)
		if m := re.FindStringSubmatch(parentID); len(m) > 1 {
			modeCode = m[1]
		}

		mode := modeMap[modeCode]
		if mode == "" {
			mode = modeCode
		}

		for _, ds := range chartData.Data.Datasets {
			metricLabel := strings.TrimSpace(strings.ToLower(ds.Label))
			metricLabel = strings.NewReplacer(" ", "_", "/", "_per_", "-", "_").Replace(metricLabel)

			for i, label := range chartData.Data.Labels {
				if i >= len(ds.Data) {
					break
				}
				key := fmt.Sprintf("%s|%s|%d", mode, metricLabel, i)
				if seen[key] {
					continue
				}
				seen[key] = true

				s := VehicleDailyStat{
					Date: label,
					Mode: mode,
				}
				switch metricLabel {
				case "kills_per_death", "k/d":
					s.KillsPerDeath = ds.Data[i]
				case "kills_per_battle", "frags_per_battle":
					s.KillsPerBattle = ds.Data[i]
				case "win_rate", "winrate":
					s.WinRate = ds.Data[i]
				case "efficiency", "effic":
					s.Efficiency = ds.Data[i]
				}
				stats = append(stats, s)
			}
		}
	})

	return mergeDailyStats(stats)
}

func mergeDailyStats(stats []VehicleDailyStat) []VehicleDailyStat {
	merged := make(map[string]VehicleDailyStat)
	var keys []string
	for _, s := range stats {
		key := s.Date + "|" + s.Mode
		if existing, ok := merged[key]; ok {
			if s.KillsPerDeath > 0 {
				existing.KillsPerDeath = s.KillsPerDeath
			}
			if s.KillsPerBattle > 0 {
				existing.KillsPerBattle = s.KillsPerBattle
			}
			if s.WinRate > 0 {
				existing.WinRate = s.WinRate
			}
			if s.Efficiency > 0 {
				existing.Efficiency = s.Efficiency
			}
			merged[key] = existing
		} else {
			merged[key] = s
			keys = append(keys, key)
		}
	}
	var result []VehicleDailyStat
	for _, k := range keys {
		result = append(result, merged[k])
	}
	return result
}

func ExtractBetween(text, start, end string) string {
	pattern := regexp.QuoteMeta(start) + `\s*\|\s*(.*?)\s*\|\s*` + regexp.QuoteMeta(end)
	re := regexp.MustCompile("(?i)" + pattern)
	m := re.FindStringSubmatch(text)
	if len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

func KeepLatestNDates(stats []VehicleDailyStat, n int) []VehicleDailyStat {
	if n <= 0 {
		n = 30
	}
	dateSet := make(map[string]bool)
	var uniqueDates []string
	for _, s := range stats {
		if !dateSet[s.Date] {
			dateSet[s.Date] = true
			uniqueDates = append(uniqueDates, s.Date)
		}
	}

	sortDatesDesc(uniqueDates)
	if len(uniqueDates) > n {
		uniqueDates = uniqueDates[:n]
	}

	keep := make(map[string]bool)
	for _, d := range uniqueDates {
		keep[d] = true
	}

	var filtered []VehicleDailyStat
	for _, s := range stats {
		if keep[s.Date] {
			filtered = append(filtered, s)
		}
	}

	for i := 0; i < len(filtered); i++ {
		for j := i + 1; j < len(filtered); j++ {
			if filtered[j].Date > filtered[i].Date {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			}
		}
	}
	return filtered
}

func sortDatesDesc(dates []string) {
	for i := 0; i < len(dates); i++ {
		for j := i + 1; j < len(dates); j++ {
			if dates[j] > dates[i] {
				dates[i], dates[j] = dates[j], dates[i]
			}
		}
	}
}

func toAbsURL(rawURL string) string {
	if strings.HasPrefix(rawURL, "http") {
		return rawURL
	}
	return BaseURL + rawURL
}

func toInt(v any) int {
	switch val := v.(type) {
	case float64:
		return int(val)
	case string:
		n, _ := strconv.Atoi(val)
		return n
	case int:
		return val
	}
	return 0
}

func toFloat(v any) float64 {
	switch val := v.(type) {
	case float64:
		return math.Round(val*100) / 100
	case string:
		n, _ := strconv.ParseFloat(val, 64)
		return math.Round(n*100) / 100
	case int:
		return float64(val)
	}
	return 0
}
