package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/warthunder/assistant/internal/model"
)

func GetPlayerTS(nickname string) (*model.TSSkillStats, error) {
	stats, err := fetchThunderSkill(nickname)
	if err == nil {
		return stats, nil
	}

	log.Printf("[player-ts] thunderskill not found for %q: %v", nickname, err)
	return nil, err
}

func fetchThunderSkill(nickname string) (*model.TSSkillStats, error) {
	encoded := url.PathEscape(nickname)
	apiURL := fmt.Sprintf("https://thunderskill.com/en/stat/%s/export/json", encoded)

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("player %s not found on thunderskill", nickname)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("thunderskill status %d: %s", resp.StatusCode, string(body))
	}

	var raw struct {
		Stats struct {
			Nick      string `json:"nick"`
			Rank      string `json:"rank"`
			LastStat  string `json:"last_stat"`
			Arcade    tsMode `json:"a"`
			Realistic tsMode `json:"r"`
			Simulator tsMode `json:"s"`
		} `json:"stats"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}

	convert := func(m tsMode) model.TSModeStats {
		return model.TSModeStats{
			Battles:     m.Mission,
			Wins:        m.Win,
			WinRate:     float64(ptrOrZero(m.Winrate)),
			Kills:       int(ptrOrZero(m.KB) * float64(m.Mission)),
			Deaths:      m.Death,
			KD:          float64(ptrOrZero(m.KD)),
			KPB:         float64(ptrOrZero(m.KB)),
			AirKills:    int(float64(m.Mission) * float64(ptrOrZero(m.KBair))),
			GroundKills: int(float64(m.Mission) * float64(ptrOrZero(m.KBground))),
			KPS:         float64(ptrOrZero(m.KPS)),
			Respawns:    float64(ptrOrZero(m.Respawns)),
			Lifetime:    ptrOrZeroInt(m.Lifetime),
		}
	}

	return &model.TSSkillStats{
		Nick:      raw.Stats.Nick,
		Rank:      raw.Stats.Rank,
		LastStat:  raw.Stats.LastStat,
		Arcade:    convert(raw.Stats.Arcade),
		Realistic: convert(raw.Stats.Realistic),
		Simulator: convert(raw.Stats.Simulator),
	}, nil
}

func SearchPlayer(nickname string) (map[string]string, error) {
	apiURL := fmt.Sprintf("https://companion-app.warthunder.com/call/?classname=eaw_Contacts&method=jzx_findUsersByNickPrefix&nick=%s&count=100&v=9", url.QueryEscape(nickname))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("parse search result: %w", err)
	}

	return result, nil
}

type tsMode struct {
	Win      int      `json:"win"`
	Mission  int      `json:"mission"`
	Death    int      `json:"death"`
	Winrate  *float64 `json:"winrate"`
	KB       *float64 `json:"kb"`
	KD       *float64 `json:"kd"`
	KPS      *float64 `json:"kps"`
	Respawns *float64 `json:"respawns_per_battle"`
	Lifetime *int     `json:"lifetime"`
	KBair    *float64 `json:"kb_air"`
	KBground *float64 `json:"kb_ground"`
	KDair    *float64 `json:"kd_air"`
	KDground *float64 `json:"kd_ground"`
}

func ptrOrZero(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}

func ptrOrZeroInt(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}
