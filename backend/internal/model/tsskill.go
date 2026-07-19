package model

type TSSkillStats struct {
	Nick     string         `json:"nick"`
	Rank     string         `json:"rank"`
	LastStat string         `json:"last_stat"`
	Arcade   TSModeStats    `json:"arcade"`
	Realistic TSModeStats   `json:"realistic"`
	Simulator TSModeStats   `json:"simulator"`
}

type TSModeStats struct {
	Battles   int      `json:"battles"`
	Wins      int      `json:"wins"`
	WinRate   float64  `json:"win_rate"`
	Kills     int      `json:"kills"`
	Deaths    int      `json:"deaths"`
	KD        float64  `json:"kd"`
	KPB       float64  `json:"kills_per_battle"`
	AirKills  int      `json:"air_kills"`
	GroundKills int    `json:"ground_kills"`
	KPS       float64  `json:"kps"`
	Respawns  float64  `json:"respawns_per_battle"`
	Lifetime  int      `json:"lifetime"`
}
