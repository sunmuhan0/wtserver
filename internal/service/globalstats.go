package service

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/warthunder/assistant/internal/model"
)

var (
	cache     *model.GlobalStats
	cacheMu   sync.RWMutex
	cacheTime time.Time
	cacheTTL  = 30 * time.Minute
)

type vehicleStatRaw struct {
	Identifier string  `json:"identifier"`
	Country    string  `json:"country"`
	Type       string  `json:"vehicle_type"`
	WinRate    float64 `json:"win_rate"`
	Games      int     `json:"games_played"`
	Players    int     `json:"players"`
	BR         float64 `json:"br"`
}

func GetGlobalStats() (*model.GlobalStats, error) {
	cacheMu.RLock()
	if cache != nil && time.Since(cacheTime) < cacheTTL {
		c := cache
		cacheMu.RUnlock()
		return c, nil
	}
	cacheMu.RUnlock()

	stats, err := fetchWTVehicleStatsDirect()
	if err == nil {
		result := buildHeatmap(stats)
		cacheMu.Lock()
		cache = result
		cacheTime = time.Now()
		cacheMu.Unlock()
		return result, nil
	}
	log.Printf("[globalstats] direct fetch failed: %v", err)

	return nil, fmt.Errorf("global stats unavailable: %v", err)
}

func fetchWTVehicleStatsDirect() ([]vehicleStatRaw, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get("https://wtvehiclesapi.duckdns.org/api/vehicles")
	if err != nil {
		return nil, fmt.Errorf("vehicles api: %w", err)
	}
	defer resp.Body.Close()

	var vehicles []struct {
		Identifier  string  `json:"identifier"`
		Country     string  `json:"country"`
		Type        string  `json:"vehicle_type"`
		WinRate     float64 `json:"win_rate"`
		Games       int     `json:"games_played"`
		Players     int     `json:"players"`
		ArcadeBr    float64 `json:"arcade_br"`
		RealisticBr float64 `json:"realistic_br"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&vehicles); err != nil {
		return nil, fmt.Errorf("parse vehicles: %w", err)
	}

	var stats []vehicleStatRaw
	for _, v := range vehicles {
		br := v.ArcadeBr
		if br == 0 {
			br = v.RealisticBr
		}
		stats = append(stats, vehicleStatRaw{
			Identifier: v.Identifier,
			Country:    v.Country,
			Type:       v.Type,
			WinRate:    v.WinRate,
			Games:      v.Games,
			Players:    v.Players,
			BR:         br,
		})
	}
	return stats, nil
}

func buildHeatmap(stats []vehicleStatRaw) *model.GlobalStats {
	nationSet := make(map[string]bool)
	typeSet := make(map[string]bool)
	cellMap := make(map[string]*model.GlobalStatCell)

	for _, s := range stats {
		nation := normalizeNation(s.Country)
		vtype := normalizeType(s.Type)
		if nation == "" || vtype == "" {
			continue
		}
		nationSet[nation] = true
		typeSet[vtype] = true
		key := nation + "|" + vtype

		if _, ok := cellMap[key]; !ok {
			cellMap[key] = &model.GlobalStatCell{
				Nation:      nation,
				Type:        vtype,
				WinRate:     s.WinRate,
				Count:       1,
				AvgBR:       s.BR,
				GamesPlayed: s.Games,
				PlayerCount: s.Players,
			}
		} else {
			c := cellMap[key]
			c.Count++
			c.GamesPlayed += s.Games
			c.PlayerCount += s.Players
			total := c.WinRate*float64(c.Count-1) + s.WinRate
			c.WinRate = math.Round(total/float64(c.Count)*10) / 10
			brTotal := c.AvgBR*float64(c.Count-1) + s.BR
			c.AvgBR = math.Round(brTotal/float64(c.Count)*10) / 10
		}
	}

	nations := sortedKeys(nationSet)
	vtypes := sortedKeys(typeSet)

	var cells []model.GlobalStatCell
	for _, n := range nations {
		for _, t := range vtypes {
			key := n + "|" + t
			if c, ok := cellMap[key]; ok {
				cells = append(cells, *c)
			} else {
				cells = append(cells, model.GlobalStatCell{
					Nation: n, Type: t,
				})
			}
		}
	}

	return &model.GlobalStats{
		Nations: nations,
		Types:   vtypes,
		Cells:   cells,
	}
}

func normalizeNation(country string) string {
	m := map[string]string{
		"usa": "usa", "us": "usa",
		"germany": "germany", "de": "germany",
		"ussr": "ussr", "ru": "ussr", "russia": "ussr",
		"britain": "britain", "uk": "britain", "gb": "britain",
		"japan": "japan", "jp": "japan",
		"china": "china", "cn": "china",
		"italy": "italy", "it": "italy",
		"france": "france", "fr": "france",
		"sweden": "sweden", "se": "sweden",
		"israel": "israel", "il": "israel",
	}
	if v, ok := m[country]; ok {
		return v
	}
	return country
}

func normalizeType(vtype string) string {
	m := map[string]string{
		"aircraft": "aircraft", "fighter": "aircraft", "assault": "aircraft",
		"bomber": "aircraft", "strike_aircraft": "aircraft",
		"tank": "tanks", "tanks": "tanks", "heavy_tank": "tanks",
		"medium_tank": "tanks", "light_tank": "tanks",
		"tank_destroyer": "tanks", "spaa": "tanks",
		"helicopter": "helicopters", "helicopters": "helicopters",
		"ship": "ships", "ships": "ships", "destroyer": "ships",
		"cruiser": "ships", "battleship": "ships",
		"coastal": "coastal", "boat": "coastal",
		"motor_gun_boat": "coastal", "motor_torpedo_boat": "coastal",
	}
	if v, ok := m[vtype]; ok {
		return v
	}
	return vtype
}

func sortedKeys(m map[string]bool) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
