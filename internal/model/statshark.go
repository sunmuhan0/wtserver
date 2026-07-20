package model

import "time"

type SSProfile struct {
	AccountID       int            `json:"account_id"`
	Nickname        string         `json:"nickname"`
	Rank            string         `json:"rank"`
	Level           int            `json:"level"`
	LastSeen        time.Time      `json:"last_seen"`
	LastBattle      time.Time      `json:"last_battle"`
	Registration    time.Time      `json:"registration"`
	Clan            string         `json:"clan"`
	Avatar          string         `json:"avatar"`
	Arcade          *SSModeStats   `json:"arcade"`
	Realistic       *SSModeStats   `json:"realistic"`
	Simulator       *SSModeStats   `json:"simulator"`
	Overall         *SSModeStats   `json:"overall"`
}

type SSModeStats struct {
	Battles       int     `json:"battles"`
	Wins          int     `json:"wins"`
	WinRate       float64 `json:"win_rate"`
	Kills         int     `json:"kills"`
	GroundKills   int     `json:"ground_kills"`
	AirKills      int     `json:"air_kills"`
	NavalKills    int     `json:"naval_kills"`
	KPB           float64 `json:"kills_per_battle"`
	Deaths        int     `json:"deaths"`
	KD            float64 `json:"kd"`
	KDRatio       float64 `json:"kd_ratio"`
	Respawns      int     `json:"respawns"`
	RespawnsPB    float64 `json:"respawns_per_battle"`
	Lifetime      float64 `json:"lifetime"`
	Damage        int64   `json:"damage"`
	DamagePB      float64 `json:"damage_per_battle"`
	SL            int64   `json:"sl"`
	SLPB          int64   `json:"sl_per_battle"`
	RP            int64   `json:"rp"`
	RPPB          int64   `json:"rp_per_battle"`
	BestKillStreak int    `json:"best_kill_streak"`
}

type SSLeaderboardEntry struct {
	ID               int    `json:"id"`
	Nickname         string `json:"nickname"`
	Rank             int    `json:"rank"`
	Score            int    `json:"score"`
	PreviousRank     *int   `json:"previous_rank"`
	Battles          int    `json:"battles"`
	ArcadeBattles    int    `json:"arcade_battles"`
	RealisticBattles int    `json:"realistic_battles"`
	SimulatorBattles int    `json:"simulator_battles"`
	KD               float64 `json:"kd"`
	WinRate          float64 `json:"win_rate"`
	Squadron         *string `json:"squadron"`
}

type SSPlayerSearchResult struct {
	ID       int    `json:"id"`
	Nickname string `json:"nickname"`
}

type SSLeaderboardHistory struct {
	ID        int                   `json:"id"`
	Nickname  string                `json:"nickname"`
	History   []SSHistoryEntry      `json:"history"`
}

type SSHistoryEntry struct {
	Date   string  `json:"date"`
	Score  int     `json:"score"`
	Rank   int     `json:"rank"`
}
