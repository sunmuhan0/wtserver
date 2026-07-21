package model

type SSPlayerDetail struct {
	Nickname  string `json:"nickname"`
	UID       int    `json:"uid"`
	Level     int    `json:"level"`
	LevelProg float64 `json:"level_progress"`
	Title     string `json:"title"`
	Squadron  string `json:"squadron"`
	Avatar    string `json:"avatar"`
	BanStatus string `json:"ban_status"`
	LastUpdate string `json:"last_update"`
	RegisterDay int64 `json:"register_day"`
	LastOnline  int64 `json:"last_online"`
	SpadedTotal int `json:"spaded_total"`
	Arcade  *SSDetailMode `json:"arcade"`
	Realistic *SSDetailMode `json:"realistic"`
	Simulator *SSDetailMode `json:"simulator"`
}

type SSDetailMode struct {
	PvP *SSDetailPvP `json:"pvp"`
}

type SSDetailPvP struct {
	Games       int `json:"games"`
	Wins        int `json:"wins"`
	WinRate     float64 `json:"win_rate"`
	AirKills    int `json:"air_kills"`
	GroundKills int `json:"ground_kills"`
	NavalKills  int `json:"naval_kills"`
	Kills       int `json:"kills"`
	AIBotKills  int `json:"ai_bot_kills"`
	Respawns    int `json:"respawns"`
	KPB         float64 `json:"kills_per_battle"`
	KD          float64 `json:"kd"`
	TimePlayed  int `json:"time_played"`
	FighterTime int `json:"fighter_time"`
	BomberTime  int `json:"bomber_time"`
	TankTime    int `json:"tank_time"`
	HeavyTankTime int `json:"heavy_tank_time"`
	TDTime      int `json:"td_time"`
	SPAATime    int `json:"spaa_time"`
	ShipTime    int `json:"ship_time"`
	HeliTime    int `json:"heli_time"`
}
