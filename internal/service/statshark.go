package service

import (
	"encoding/json"
	"fmt"
	"net/url"

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
	return searchPlayerSSAPI(nickname, token)
}

func browserFetch(method, path string, body string) (int, string, error) {
	b := GetBrowser()
	if b == nil {
		return 0, "", fmt.Errorf("browser not initialized")
	}
	return b.Fetch(method, path, nil, body)
}

func searchPlayerSSAPI(nickname string, token string) ([]model.SSPlayerSearchResult, error) {
	if err := ensureBrowser(); err != nil {
		return nil, fmt.Errorf("browser unavailable: %w", err)
	}

	path := fmt.Sprintf("/api/stat/GetIdByName?Name=%s&IgnoreCase=true&MaxCount=25&Details=false",
		url.QueryEscape(nickname))

	status, body, err := browserFetch("GET", path, "")
	if err != nil {
		return nil, fmt.Errorf("browser fetch: %w", err)
	}
	if status == 406 {
		go GetBrowser().Refresh()
		return nil, fmt.Errorf("statshark api requires valid turnstile token (got 406)")
	}
	if status != 200 {
		return nil, fmt.Errorf("statshark status %d: %s", status, body)
	}

	var searchResult map[string]string
	if err := json.Unmarshal([]byte(body), &searchResult); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}

	var results []model.SSPlayerSearchResult
	for idStr, name := range searchResult {
		var pid int
		fmt.Sscanf(idStr, "%d", &pid)
		results = append(results, model.SSPlayerSearchResult{
			ID:       pid,
			Nickname: name,
		})
	}
	if results == nil {
		results = []model.SSPlayerSearchResult{}
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
	return getPlayerSSAPI(nickname, token)
}

func getPlayerSSAPI(nickname string, token string) (*model.SSProfile, error) {
	if err := ensureBrowser(); err != nil {
		return nil, fmt.Errorf("browser unavailable: %w", err)
	}

	b := GetBrowser()

	searchPath := fmt.Sprintf("/api/stat/GetIdByName?Name=%s&IgnoreCase=true&MaxCount=1&Details=false",
		url.QueryEscape(nickname))
	status, body, err := b.Fetch("GET", searchPath, nil, "")
	if err != nil {
		return nil, fmt.Errorf("browser fetch search: %w", err)
	}
	if status != 200 {
		return nil, fmt.Errorf("statshark search status %d: %s", status, body)
	}
	var searchResult map[string]string
	if err := json.Unmarshal([]byte(body), &searchResult); err != nil {
		return nil, fmt.Errorf("parse search: %w", err)
	}
	var uid string
	for id := range searchResult {
		uid = id
		break
	}
	if uid == "" {
		return nil, fmt.Errorf("player %q not found", nickname)
	}

	path := fmt.Sprintf("/api/stat/GetLeaderboardHistoryById/%s", uid)
	status, body, err = b.Fetch("GET", path, nil, "")
	if err != nil {
		return nil, fmt.Errorf("browser fetch: %w", err)
	}
	if status == 406 {
		go b.Refresh()
		return nil, fmt.Errorf("statshark api requires valid turnstile token (got 406)")
	}
	if status != 200 {
		return nil, fmt.Errorf("statshark status %d: %s", status, body)
	}

	var rawArr []json.RawMessage
	if err := json.Unmarshal([]byte(body), &rawArr); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}

	if len(rawArr) == 0 {
		return nil, fmt.Errorf("player not found")
	}

	return parseSSProfile(rawArr[0])
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
	if err := ensureBrowser(); err != nil {
		return nil, fmt.Errorf("browser unavailable: %w", err)
	}

	b := GetBrowser()

	searchPath := fmt.Sprintf("/api/stat/GetIdByName?Name=%s&IgnoreCase=true&MaxCount=1&Details=false",
		url.QueryEscape(nickname))

	status, body, err := b.Fetch("GET", searchPath, nil, "")
	if err != nil {
		return nil, fmt.Errorf("browser fetch search: %w", err)
	}
	if status == 406 {
		go b.Refresh()
		return nil, fmt.Errorf("statshark api requires valid turnstile token (got 406)")
	}
	if status != 200 {
		return nil, fmt.Errorf("statshark search status %d: %s", status, body)
	}

	var searchResult map[string]string
	if err := json.Unmarshal([]byte(body), &searchResult); err != nil {
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

	detailPath := fmt.Sprintf("/api/stat/MakeStatRequestById/%s?update=true", uid)

	status2, body2, err := b.Fetch("POST", detailPath, nil, "{}")
	if err != nil {
		return nil, fmt.Errorf("browser fetch detail: %w", err)
	}
	if status2 == 406 {
		go b.Refresh()
		return nil, fmt.Errorf("statshark api requires valid turnstile token (got 406)")
	}
	if status2 != 200 {
		return nil, fmt.Errorf("statshark detail status %d: %s", status2, body2)
	}

	return parseSSPlayerDetail([]byte(body2))
}

func parseSSPlayerDetail(data []byte) (*model.SSPlayerDetail, error) {
	var raw struct {
		Basics struct {
			Nickname      string  `json:"nickname"`
			UID           string  `json:"uid"`
			Level         int     `json:"level"`
			LevelProg     float64 `json:"levelProgress"`
			Title         string  `json:"title"`
			TitleDisc     string  `json:"titleDisc"`
			SquadronName  *string `json:"SquadronName"`
			SquadronID    string  `json:"squadid"`
			PFP           string  `json:"pfp"`
			BanStatus     string  `json:"banStatus"`
			TotalViews    int     `json:"totalviews"`
			RecentViews   int     `json:"recentviews"`
			LastUpdate    string  `json:"lastupdate"`
		} `json:"Basics"`
		Profile struct {
			Arcade struct {
				PvPPlayed   map[string]interface{} `json:"pvp_played"`
				SkirmishPlayed map[string]interface{} `json:"skirmish_played"`
			} `json:"arcade"`
			Rb struct {
				PvPPlayed   map[string]interface{} `json:"pvp_played"`
				SkirmishPlayed map[string]interface{} `json:"skirmish_played"`
			} `json:"rb"`
			Sim struct {
				PvPPlayed   map[string]interface{} `json:"pvp_played"`
				SkirmishPlayed map[string]interface{} `json:"skirmish_played"`
			} `json:"sim"`
			Leaderboard json.RawMessage `json:"Leaderboard"`
		} `json:"Profile"`
		Misc struct {
			RegisterDay     int64 `json:"registerDay"`
			LastDayOnline   int64 `json:"lastDayOnline"`
			SquadronHistory []struct {
				ClanID   int    `json:"ClanID"`
				ClanTag  string `json:"ClanTag"`
				Date     string `json:"Date"`
			} `json:"SquadronHistory"`
			NameHistory []struct {
				IGN  string `json:"IGN"`
				Date string `json:"Date"`
			} `json:"NameHistory"`
			Titles []struct {
				Name string `json:"name"`
				Lang string `json:"lang"`
				Disc string `json:"disc"`
			} `json:"titles"`
		} `json:"Misc"`
		SpadeInfo map[string]map[string]interface{} `json:"SpadeInfo"`
		Ratings struct {
			Total map[string]*struct {
				Rating float64 `json:"rating"`
			} `json:"total"`
		} `json:"ratings"`
		RatingsNeo map[string]*struct {
			Monthly struct {
				Wt8         float64 `json:"wt8"`
				OriginalWt8 float64 `json:"original_wt8"`
				IsSuspicious bool   `json:"isSuspicious"`
				Contributions struct {
					Kills    float64 `json:"kills"`
					Position float64 `json:"position"`
					Win      float64 `json:"win"`
					Score    float64 `json:"score"`
				} `json:"contributions"`
			} `json:"monthly"`
		} `json:"ratingsNeo"`
		UserInfo struct {
			Likes    int `json:"Likes"`
			Dislikes int `json:"Dislikes"`
		} `json:"UserInfo"`
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

	squadID := 0
	fmt.Sscanf(raw.Basics.SquadronID, "%d", &squadID)

	spadedTotal := 0
	spadedByCountry := make(map[string]int)
	for country, vehicles := range raw.SpadeInfo {
		count := len(vehicles)
		spadedByCountry[country] = count
		spadedTotal += count
	}

	detail := &model.SSPlayerDetail{
		Nickname:        raw.Basics.Nickname,
		UID:             uid,
		Level:           raw.Basics.Level,
		LevelProg:       raw.Basics.LevelProg,
		Title:           raw.Basics.Title,
		TitleDisc:       raw.Basics.TitleDisc,
		Squadron:        squadron,
		SquadronID:      squadID,
		Avatar:          raw.Basics.PFP,
		BanStatus:       raw.Basics.BanStatus,
		TotalViews:      raw.Basics.TotalViews,
		RecentViews:     raw.Basics.RecentViews,
		LastUpdate:      raw.Basics.LastUpdate,
		RegisterDay:     raw.Misc.RegisterDay,
		LastOnline:      raw.Misc.LastDayOnline,
		SpadedTotal:     spadedTotal,
		SpadedByCountry: spadedByCountry,
		Arcade:          parseSSDetailModeFull(raw.Profile.Arcade.PvPPlayed, raw.Profile.Arcade.SkirmishPlayed),
		Realistic:       parseSSDetailModeFull(raw.Profile.Rb.PvPPlayed, raw.Profile.Rb.SkirmishPlayed),
		Simulator:       parseSSDetailModeFull(raw.Profile.Sim.PvPPlayed, raw.Profile.Sim.SkirmishPlayed),
	}

	if raw.Profile.Leaderboard != nil {
		detail.Leaderboard = parseLeaderboard(raw.Profile.Leaderboard)
	}
	if raw.Ratings.Total != nil {
		detail.Ratings = &model.SSRatings{Total: make(map[string]*model.SSRatingsEntry)}
		for k, v := range raw.Ratings.Total {
			detail.Ratings.Total[k] = &model.SSRatingsEntry{Rating: v.Rating}
		}
	}
	if raw.RatingsNeo != nil {
		detail.RatingsNeo = &model.SSRatingsNeo{}
		if e, ok := raw.RatingsNeo["tank_realistic"]; ok {
			detail.RatingsNeo.TankRealistic = &model.SSRatingsNeoEntry{}
			detail.RatingsNeo.TankRealistic.Monthly.Wt8 = e.Monthly.Wt8
			detail.RatingsNeo.TankRealistic.Monthly.OriginalWt8 = e.Monthly.OriginalWt8
			detail.RatingsNeo.TankRealistic.Monthly.IsSuspicious = e.Monthly.IsSuspicious
			detail.RatingsNeo.TankRealistic.Monthly.Contributions.Kills = e.Monthly.Contributions.Kills
			detail.RatingsNeo.TankRealistic.Monthly.Contributions.Position = e.Monthly.Contributions.Position
			detail.RatingsNeo.TankRealistic.Monthly.Contributions.Win = e.Monthly.Contributions.Win
			detail.RatingsNeo.TankRealistic.Monthly.Contributions.Score = e.Monthly.Contributions.Score
		}
		if e, ok := raw.RatingsNeo["tank_arcade"]; ok {
			detail.RatingsNeo.TankArcade = &model.SSRatingsNeoEntry{}
			detail.RatingsNeo.TankArcade.Monthly.Wt8 = e.Monthly.Wt8
			detail.RatingsNeo.TankArcade.Monthly.IsSuspicious = e.Monthly.IsSuspicious
		}
		if e, ok := raw.RatingsNeo["air_realistic"]; ok {
			detail.RatingsNeo.AirRealistic = &model.SSRatingsNeoEntry{}
			detail.RatingsNeo.AirRealistic.Monthly.Wt8 = e.Monthly.Wt8
			detail.RatingsNeo.AirRealistic.Monthly.IsSuspicious = e.Monthly.IsSuspicious
		}
		if e, ok := raw.RatingsNeo["air_arcade"]; ok {
			detail.RatingsNeo.AirArcade = &model.SSRatingsNeoEntry{}
			detail.RatingsNeo.AirArcade.Monthly.Wt8 = e.Monthly.Wt8
			detail.RatingsNeo.AirArcade.Monthly.IsSuspicious = e.Monthly.IsSuspicious
		}
	}

	for _, h := range raw.Misc.NameHistory {
		detail.NameHistory = append(detail.NameHistory, model.SSNameHistory{IGN: h.IGN, Date: h.Date})
	}
	for _, h := range raw.Misc.SquadronHistory {
		detail.SquadronHistory = append(detail.SquadronHistory, model.SSSquadronHistory{ClanID: h.ClanID, ClanTag: h.ClanTag, Date: h.Date})
	}
	for _, t := range raw.Misc.Titles {
		detail.Titles = append(detail.Titles, model.SSTitle{Name: t.Name, Lang: t.Lang, Disc: t.Disc})
	}

	detail.UserInfo = &model.SSUserInfo{Likes: raw.UserInfo.Likes, Dislikes: raw.UserInfo.Dislikes}

	return detail, nil
}

func parseLeaderboard(raw json.RawMessage) *model.SSLeaderboardData {
	var lbRaw map[string]json.RawMessage
	if err := json.Unmarshal(raw, &lbRaw); err != nil {
		return nil
	}

	lb := &model.SSLeaderboardData{}
	parseLbCat := func(data json.RawMessage) map[string]*model.SSLeaderboardEntry {
		if data == nil {
			return nil
		}
		var catRaw map[string]json.RawMessage
		if err := json.Unmarshal(data, &catRaw); err != nil {
			return nil
		}
		result := make(map[string]*model.SSLeaderboardEntry)
		for _, subKey := range []string{"value_total", "value_inhistory"} {
			subRaw, ok := catRaw[subKey]
			if !ok {
				continue
			}
			var entries map[string]json.RawMessage
			if err := json.Unmarshal(subRaw, &entries); err != nil {
				continue
			}
			prefix := ""
			if subKey == "value_inhistory" {
				prefix = "recent_"
			}
			for k, v := range entries {
				var entry struct {
					ValueTotal float64 `json:"value_total"`
					Idx        int     `json:"idx"`
				}
				if err := json.Unmarshal(v, &entry); err != nil {
					continue
				}
				if entry.Idx <= 0 {
					continue
				}
				result[prefix+k] = &model.SSLeaderboardEntry{Value: entry.ValueTotal, Rank: entry.Idx}
			}
		}
		return result
	}
	lb.AirArcade = parseLbCat(lbRaw["air_arcade"])
	lb.AirRealistic = parseLbCat(lbRaw["air_realistic"])
	lb.AirSimulation = parseLbCat(lbRaw["air_simulation"])
	lb.Arcade = parseLbCat(lbRaw["arcade"])
	lb.HelicopterArcade = parseLbCat(lbRaw["helicopter_arcade"])
	lb.Historical = parseLbCat(lbRaw["historical"])
	lb.Simulation = parseLbCat(lbRaw["simulation"])
	lb.TankArcade = parseLbCat(lbRaw["tank_arcade"])
	lb.TankRealistic = parseLbCat(lbRaw["tank_realistic"])
	lb.TankSimulation = parseLbCat(lbRaw["tank_simulation"])
	return lb
}

func parseSSDetailMode(pvp map[string]interface{}) *model.SSDetailMode {
	return parseSSDetailModeFull(pvp, nil)
}

func parseSSDetailModeFull(pvp map[string]interface{}, skirmish map[string]interface{}) *model.SSDetailMode {
	if pvp == nil {
		return nil
	}
	mode := &model.SSDetailMode{}
	mode.PvP = parseSSDetailPvP(pvp)
	if skirmish != nil {
		mode.Skirmish = parseSSDetailSkirmish(skirmish)
	}
	return mode
}

func parseSSDetailPvP(pvp map[string]interface{}) *model.SSDetailPvP {
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

	return &model.SSDetailPvP{
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
		AssaultTime:   toInt(pvp["assaultTimePlayed"]),
		TankTime:      toInt(pvp["tankTimePlayed"]),
		HeavyTankTime: toInt(pvp["heavy_tankTimePlayed"]),
		TDTime:        toInt(pvp["tank_destroyerTimePlayed"]),
		SPAATime:      toInt(pvp["SPAATimePlayed"]),
		ShipTime:      toInt(pvp["shipTimePlayed"]),
		TorpedoBoatTime: toInt(pvp["torpedo_boatTimePlayed"]),
		GunBoatTime:     toInt(pvp["gun_boatTimePlayed"]),
		DestroyerTime:   toInt(pvp["destroyerTimePlayed"]),
		CruiserTime:     toInt(pvp["cruiserTimePlayed"]),
		HeliTime:        toInt(pvp["helicopterTimePlayed"]),
		HumanTime:       toInt(pvp["humanTimePlayed"]),
	}
}

func parseSSDetailSkirmish(s map[string]interface{}) *model.SSDetailSkirmish {
	games := toInt(s["games"])
	if games == 0 {
		return nil
	}
	return &model.SSDetailSkirmish{
		Games:       games,
		Wins:        toInt(s["wins"]),
		AirKills:    toInt(s["airKillsP"]),
		GroundKills: toInt(s["groundKillsP"]),
		NavalKills:  toInt(s["navalKillsP"]),
		Respawns:    toInt(s["respawns"]),
		TimePlayed:  toInt(s["timePlayed"]),
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
	return getLeaderboardHistorySSAPI(nickname, token)
}

func getLeaderboardHistorySSAPI(nickname string, token string) (*model.SSLeaderboardHistory, error) {
	if err := ensureBrowser(); err != nil {
		return nil, fmt.Errorf("browser unavailable: %w", err)
	}

	b := GetBrowser()

	searchPath := fmt.Sprintf("/api/stat/GetIdByName?Name=%s&IgnoreCase=true&MaxCount=1&Details=false",
		url.QueryEscape(nickname))
	status, body, err := b.Fetch("GET", searchPath, nil, "")
	if err != nil {
		return nil, fmt.Errorf("browser fetch search: %w", err)
	}
	if status != 200 {
		return nil, fmt.Errorf("statshark search status %d: %s", status, body)
	}
	var searchResult map[string]string
	if err := json.Unmarshal([]byte(body), &searchResult); err != nil {
		return nil, fmt.Errorf("parse search: %w", err)
	}
	var uid string
	for id := range searchResult {
		uid = id
		break
	}
	if uid == "" {
		return nil, fmt.Errorf("player %q not found", nickname)
	}

	path := fmt.Sprintf("/api/stat/GetLeaderboardHistoryById/%s", uid)
	status, body, err = b.Fetch("GET", path, nil, "")
	if err != nil {
		return nil, fmt.Errorf("browser fetch: %w", err)
	}
	if status == 406 {
		go b.Refresh()
		return nil, fmt.Errorf("statshark api requires valid turnstile token (got 406)")
	}
	if status != 200 {
		return nil, fmt.Errorf("statshark status %d: %s", status, body)
	}

	var rawArr []json.RawMessage
	if err := json.Unmarshal([]byte(body), &rawArr); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}

	if len(rawArr) == 0 {
		return nil, fmt.Errorf("player not found")
	}

	history := &model.SSLeaderboardHistory{}

	if len(rawArr) >= 1 {
		var entry struct {
			Date string `json:"date"`
			Data struct {
				Arcade struct {
					T interface{} `json:"t"`
				} `json:"arcade"`
				Historical struct {
					T interface{} `json:"t"`
				} `json:"historical"`
			} `json:"data"`
		}
		if err := json.Unmarshal(rawArr[0], &entry); err == nil {
			history.Nickname = nickname
		}
	}

	var entries []struct {
		Date  string `json:"date"`
	}
	if err := json.Unmarshal([]byte(body), &entries); err == nil {
		for _, e := range entries {
			history.History = append(history.History, model.SSHistoryEntry{
				Date: e.Date,
			})
		}
	}

	return history, nil
}
