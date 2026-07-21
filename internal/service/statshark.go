package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/warthunder/assistant/internal/model"
)

const ssBase = "https://statshark.net"

func SearchPlayerSS(nickname string, token string) ([]model.SSPlayerSearchResult, error) {
	if token != "" {
		return searchPlayerSSAPI(nickname, token)
	}
	return searchPlayerSSFallback(nickname)
}

func SearchPlayerSSV3(nickname string, token string) ([]model.SSPlayerSearchResult, error) {
	if token == "" {
		return nil, fmt.Errorf("X-Turnstile-Token header is required")
	}
	return searchPlayerSSAPI(nickname, token)
}

func searchPlayerSSAPI(nickname string, token string) ([]model.SSPlayerSearchResult, error) {
	u := fmt.Sprintf("%s/api/stat/GetIdByName?Name=%s&IgnoreCase=true&MaxCount=25&Details=false",
		ssBase, url.QueryEscape(nickname))

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("X-Turnstile-Token", token)
	req.Header.Set("Referer", ssBase+"/players")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 406 {
		return nil, fmt.Errorf("statshark api requires valid turnstile token (got 406)")
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("statshark status %d: %s", resp.StatusCode, string(body))
	}

	var raw []struct {
		ID       int    `json:"id"`
		Nickname string `json:"nickname"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}

	var results []model.SSPlayerSearchResult
	for _, r := range raw {
		results = append(results, model.SSPlayerSearchResult{
			ID:       r.ID,
			Nickname: r.Nickname,
		})
	}
	return results, nil
}

func searchPlayerSSFallback(nickname string) ([]model.SSPlayerSearchResult, error) {
	result, err := SearchPlayer(nickname)
	if err != nil {
		return nil, err
	}
	var players []model.SSPlayerSearchResult
	for idStr, name := range result {
		var pid int
		fmt.Sscanf(idStr, "%d", &pid)
		players = append(players, model.SSPlayerSearchResult{
			ID:       pid,
			Nickname: name,
		})
	}
	if players == nil {
		return []model.SSPlayerSearchResult{}, nil
	}
	return players, nil
}

func GetPlayerSS(nickname string, token string) (*model.SSProfile, error) {
	if token != "" {
		return getPlayerSSAPI(nickname, token)
	}
	return getPlayerSSFallback(nickname)
}

func GetPlayerSSV3(nickname string, token string) (*model.SSProfile, error) {
	if token == "" {
		return nil, fmt.Errorf("X-Turnstile-Token header is required")
	}
	return getPlayerSSAPI(nickname, token)
}

func getPlayerSSAPI(nickname string, token string) (*model.SSProfile, error) {
	u := fmt.Sprintf("%s/api/stat/GetLeaderboardHistoryById/%s", ssBase, url.PathEscape(nickname))

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("X-Turnstile-Token", token)
	req.Header.Set("Referer", ssBase+"/player/"+url.PathEscape(nickname))

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 406 {
		return nil, fmt.Errorf("statshark api requires valid turnstile token (got 406)")
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("statshark status %d: %s", resp.StatusCode, string(body))
	}

	var raw struct {
		Result []json.RawMessage `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}

	if len(raw.Result) == 0 {
		return nil, fmt.Errorf("player not found")
	}

	return parseSSProfile(raw.Result[0])
}

func parseSSProfile(data json.RawMessage) (*model.SSProfile, error) {
	var raw struct {
		Account struct {
			ID       int    `json:"id"`
			Nickname string `json:"nickname"`
			Rank     string `json:"rank"`
			Level    int    `json:"level"`
		} `json:"account"`
		Stats struct {
			Arcade    json.RawMessage `json:"a"`
			Realistic json.RawMessage `json:"r"`
			Simulator json.RawMessage `json:"s"`
		} `json:"stats"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse profile: %w", err)
	}

	profile := &model.SSProfile{
		AccountID: raw.Account.ID,
		Nickname:  raw.Account.Nickname,
		Rank:      raw.Account.Rank,
		Level:     raw.Account.Level,
	}

	if raw.Stats.Arcade != nil && string(raw.Stats.Arcade) != "null" {
		profile.Arcade = parseSSModeStats(raw.Stats.Arcade)
	}
	if raw.Stats.Realistic != nil && string(raw.Stats.Realistic) != "null" {
		profile.Realistic = parseSSModeStats(raw.Stats.Realistic)
	}
	if raw.Stats.Simulator != nil && string(raw.Stats.Simulator) != "null" {
		profile.Simulator = parseSSModeStats(raw.Stats.Simulator)
	}

	profile.Overall = computeOverallStats(profile.Arcade, profile.Realistic, profile.Simulator)

	return profile, nil
}

func parseSSModeStats(data json.RawMessage) *model.SSModeStats {
	var raw struct {
		Battles     int     `json:"battles"`
		Wins        int     `json:"wins"`
		WinRate     float64 `json:"winrate"`
		Kills       int     `json:"kills"`
		Deaths      int     `json:"deaths"`
		KD          float64 `json:"kd"`
		Respawns    int     `json:"respawns"`
		Lifetime    float64 `json:"lifetime"`
		Damage      int64   `json:"damage"`
		BestKS      int     `json:"best_killstreak"`
		SL          int64   `json:"sl"`
		RP          int64   `json:"rp"`
		AirKills    int     `json:"air_kills"`
		GroundKills int     `json:"ground_kills"`
		NavalKills  int     `json:"naval_kills"`
		KPB         float64 `json:"kpb"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil
	}
	return &model.SSModeStats{
		Battles:       raw.Battles,
		Wins:          raw.Wins,
		WinRate:       raw.WinRate,
		Kills:         raw.Kills,
		Deaths:        raw.Deaths,
		KD:            raw.KD,
		Respawns:      raw.Respawns,
		Lifetime:      raw.Lifetime,
		Damage:        raw.Damage,
		BestKillStreak: raw.BestKS,
		AirKills:      raw.AirKills,
		GroundKills:   raw.GroundKills,
		NavalKills:    raw.NavalKills,
		KPB:           raw.KPB,
		SL:            raw.SL,
		RP:            raw.RP,
	}
}

func computeOverallStats(modes ...*model.SSModeStats) *model.SSModeStats {
	overall := model.SSModeStats{}
	for _, m := range modes {
		if m != nil {
			addModeStats(&overall, m)
		}
	}
	if overall.Battles == 0 {
		return nil
	}
	overall.WinRate = float64(overall.Wins) / float64(overall.Battles) * 100
	overall.KPB = float64(overall.Kills) / float64(overall.Battles)
	overall.KD = float64(overall.Kills) / float64(overall.Deaths)
	overall.RespawnsPB = float64(overall.Respawns) / float64(overall.Battles)
	return &overall
}

func getPlayerSSFallback(nickname string) (*model.SSProfile, error) {
	ts, err := GetPlayerTS(nickname)
	if err != nil {
		return nil, fmt.Errorf("player not found on thunderskill and no turnstile token provided: %w", err)
	}

	profile := &model.SSProfile{
		Nickname: ts.Nick,
		Rank:     ts.Rank,
	}

	if ts.Arcade.Battles > 0 {
		a := ts.Arcade
		profile.Arcade = &model.SSModeStats{
			Battles:  a.Battles,
			Wins:     a.Wins,
			WinRate:  a.WinRate,
			Kills:    a.Kills,
			Deaths:   a.Deaths,
			KD:       a.KD,
			Respawns: int(a.Respawns * float64(a.Battles)),
			Lifetime: float64(a.Lifetime),
		}
	}
	if ts.Realistic.Battles > 0 {
		r := ts.Realistic
		profile.Realistic = &model.SSModeStats{
			Battles:  r.Battles,
			Wins:     r.Wins,
			WinRate:  r.WinRate,
			Kills:    r.Kills,
			Deaths:   r.Deaths,
			KD:       r.KD,
			Respawns: int(r.Respawns * float64(r.Battles)),
			Lifetime: float64(r.Lifetime),
		}
	}
	if ts.Simulator.Battles > 0 {
		s := ts.Simulator
		profile.Simulator = &model.SSModeStats{
			Battles:  s.Battles,
			Wins:     s.Wins,
			WinRate:  s.WinRate,
			Kills:    s.Kills,
			Deaths:   s.Deaths,
			KD:       s.KD,
			Respawns: int(s.Respawns * float64(s.Battles)),
			Lifetime: float64(s.Lifetime),
		}
	}

	profile.Overall = computeOverallStats(profile.Arcade, profile.Realistic, profile.Simulator)

	return profile, nil
}

func addModeStats(total *model.SSModeStats, mode *model.SSModeStats) {
	total.Battles += mode.Battles
	total.Wins += mode.Wins
	total.Kills += mode.Kills
	total.Deaths += mode.Deaths
	total.Respawns += mode.Respawns
	if mode.Lifetime > total.Lifetime {
		total.Lifetime = mode.Lifetime
	}
}

func GetPlayerDetailV3(nickname string, token string) (*model.SSPlayerDetail, error) {
	if token == "" {
		return nil, fmt.Errorf("X-Turnstile-Token header is required")
	}

	searchURL := fmt.Sprintf("%s/api/stat/GetIdByName?Name=%s&IgnoreCase=true&MaxCount=1&Details=false",
		ssBase, url.QueryEscape(nickname))

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create search request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("X-Turnstile-Token", token)
	req.Header.Set("Referer", ssBase+"/players")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 406 {
		return nil, fmt.Errorf("statshark api requires valid turnstile token (got 406)")
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("statshark search status %d: %s", resp.StatusCode, string(body))
	}

	var searchResult map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("parse search result: %w", err)
	}

	var uid string
	for id, name := range searchResult {
		if name == nickname {
			uid = id
			break
		}
	}
	if uid == "" {
		for id := range searchResult {
			uid = id
			break
		}
	}
	if uid == "" {
		return nil, fmt.Errorf("player %q not found", nickname)
	}

	detailURL := fmt.Sprintf("%s/api/stat/MakeStatRequestById/%s?update=true", ssBase, uid)
	req2, err := http.NewRequest("POST", detailURL, bytes.NewReader([]byte("{}")))
	if err != nil {
		return nil, fmt.Errorf("create detail request: %w", err)
	}
	req2.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req2.Header.Set("Accept", "application/json, text/plain, */*")
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-Turnstile-Token", token)
	req2.Header.Set("Origin", ssBase)
	req2.Header.Set("Referer", ssBase+"/player/"+uid)

	resp2, err := client.Do(req2)
	if err != nil {
		return nil, fmt.Errorf("detail request failed: %w", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode == 406 {
		return nil, fmt.Errorf("statshark api requires valid turnstile token (got 406)")
	}
	if resp2.StatusCode != 200 {
		body, _ := io.ReadAll(resp2.Body)
		return nil, fmt.Errorf("statshark detail status %d: %s", resp2.StatusCode, string(body))
	}

	body, _ := io.ReadAll(resp2.Body)
	return parseSSPlayerDetail(body)
}

func parseSSPlayerDetail(data []byte) (*model.SSPlayerDetail, error) {
	var raw struct {
		Basics struct {
			Nickname    string `json:"nickname"`
			UID         string `json:"uid"`
			Level       int    `json:"level"`
			LevelProg   float64 `json:"levelProgress"`
			Title       string `json:"title"`
			SquadronName *string `json:"SquadronName"`
			PFP         string `json:"pfp"`
			BanStatus   string `json:"banStatus"`
			LastUpdate  string `json:"lastupdate"`
		} `json:"Basics"`
		Profile struct {
			Arcade struct {
				PvPPlayed map[string]interface{} `json:"pvp_played"`
			} `json:"arcade"`
			Rb struct {
				PvPPlayed map[string]interface{} `json:"pvp_played"`
			} `json:"rb"`
			Sim struct {
				PvPPlayed map[string]interface{} `json:"pvp_played"`
			} `json:"sim"`
		} `json:"Profile"`
		Misc struct {
			RegisterDay int64 `json:"registerDay"`
			LastDayOnline int64 `json:"lastDayOnline"`
		} `json:"Misc"`
		SpadeInfo map[string]interface{} `json:"SpadeInfo"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse detail: %w", err)
	}

	squadron := ""
	if raw.Basics.SquadronName != nil {
		squadron = *raw.Basics.SquadronName
	}

	uid := 0
	fmt.Sscanf(raw.Basics.UID, "%d", &uid)

	spadedTotal := 0
	for _, v := range raw.SpadeInfo {
		if m, ok := v.(map[string]interface{}); ok {
			spadedTotal += len(m)
		}
	}

	detail := &model.SSPlayerDetail{
		Nickname:    raw.Basics.Nickname,
		UID:         uid,
		Level:       raw.Basics.Level,
		LevelProg:   raw.Basics.LevelProg,
		Title:       raw.Basics.Title,
		Squadron:    squadron,
		Avatar:      raw.Basics.PFP,
		BanStatus:   raw.Basics.BanStatus,
		LastUpdate:  raw.Basics.LastUpdate,
		RegisterDay: raw.Misc.RegisterDay,
		LastOnline:  raw.Misc.LastDayOnline,
		SpadedTotal: spadedTotal,
		Arcade:    parseSSDetailMode(raw.Profile.Arcade.PvPPlayed),
		Realistic: parseSSDetailMode(raw.Profile.Rb.PvPPlayed),
		Simulator: parseSSDetailMode(raw.Profile.Sim.PvPPlayed),
	}

	return detail, nil
}

func parseSSDetailMode(pvp map[string]interface{}) *model.SSDetailMode {
	if pvp == nil {
		return nil
	}
	games := toInt(pvp["games"])
	if games == 0 {
		return nil
	}
	wins := toInt(pvp["wins"])
	airK := toInt(pvp["airKillsP"])
	groundK := toInt(pvp["groundKillsP"])
	navalK := toInt(pvp["navalKillsP"])
	aiAir := toInt(pvp["airKillsAIAndBot"])
	aiGround := toInt(pvp["groundKillsAIAndBot"])
	kills := airK + groundK + navalK
	respawns := toInt(pvp["respawns"])
	kd := 0.0
	if respawns > wins {
		kd = float64(kills) / float64(respawns-wins)
	}
	kpb := 0.0
	if games > 0 {
		kpb = float64(kills) / float64(games)
	}
	wr := 0.0
	if games > 0 {
		wr = float64(wins) / float64(games) * 100
	}

	return &model.SSDetailMode{
		PvP: &model.SSDetailPvP{
			Games:         games,
			Wins:          wins,
			WinRate:       wr,
			AirKills:      airK,
			GroundKills:   groundK,
			NavalKills:    navalK,
			Kills:         kills,
			AIBotKills:    aiAir + aiGround,
			Respawns:      respawns,
			KPB:           kpb,
			KD:            kd,
			TimePlayed:    toInt(pvp["timePlayed"]),
			FighterTime:   toInt(pvp["fighterTimePlayed"]),
			BomberTime:    toInt(pvp["bomberTimePlayed"]),
			TankTime:      toInt(pvp["tankTimePlayed"]),
			HeavyTankTime: toInt(pvp["heavy_tankTimePlayed"]),
			TDTime:        toInt(pvp["tank_destroyerTimePlayed"]),
			SPAATime:      toInt(pvp["SPAATimePlayed"]),
			ShipTime:      toInt(pvp["shipTimePlayed"]),
			HeliTime:      toInt(pvp["helicopterTimePlayed"]),
		},
	}
}

func toInt(v interface{}) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case json.Number:
		v, _ := n.Int64()
		return int(v)
	default:
		return 0
	}
}

func GetLeaderboardHistorySS(nickname string, token string) (*model.SSLeaderboardHistory, error) {
	if token != "" {
		return getLeaderboardHistorySSAPI(nickname, token)
	}
	return nil, fmt.Errorf("statshark api requires turnstile token, pass X-Turnstile-Token header")
}

func GetLeaderboardHistorySSV3(nickname string, token string) (*model.SSLeaderboardHistory, error) {
	if token == "" {
		return nil, fmt.Errorf("X-Turnstile-Token header is required")
	}
	return getLeaderboardHistorySSAPI(nickname, token)
}

func getLeaderboardHistorySSAPI(nickname string, token string) (*model.SSLeaderboardHistory, error) {
	u := fmt.Sprintf("%s/api/stat/GetLeaderboardHistoryById/%s", ssBase, url.PathEscape(nickname))

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("X-Turnstile-Token", token)
	req.Header.Set("Referer", ssBase+"/player/"+url.PathEscape(nickname))

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 406 {
		return nil, fmt.Errorf("statshark api requires valid turnstile token (got 406)")
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("statshark status %d: %s", resp.StatusCode, string(body))
	}

	var raw struct {
		Result []json.RawMessage `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}

	if len(raw.Result) == 0 {
		return nil, fmt.Errorf("player not found")
	}

	history := &model.SSLeaderboardHistory{}

	if len(raw.Result) >= 1 {
		var account struct {
			ID       int    `json:"id"`
			Nickname string `json:"nickname"`
		}
		if err := json.Unmarshal(raw.Result[0], &account); err == nil {
			history.ID = account.ID
			history.Nickname = account.Nickname
		}
	}

	if len(raw.Result) >= 2 {
		var entries []struct {
			Date  string `json:"date"`
			Score int    `json:"score"`
			Rank  int    `json:"rank"`
		}
		if err := json.Unmarshal(raw.Result[1], &entries); err == nil {
			for _, e := range entries {
				history.History = append(history.History, model.SSHistoryEntry{
					Date:  e.Date,
					Score: e.Score,
					Rank:  e.Rank,
				})
			}
		}
	}

	return history, nil
}
