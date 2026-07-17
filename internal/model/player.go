package model

type Player struct {
	Nickname    string       `json:"nickname"`
	Title       string       `json:"title"`
	Level       int          `json:"level"`
	Country     string       `json:"country"`
	Clan        string       `json:"clan"`
	Statistics  *Statistics  `json:"statistics"`
	Vehicles    []VehicleRef `json:"vehicles"`
}

type Statistics struct {
	Battles     int     `json:"battles"`
	Wins        int     `json:"wins"`
	Losses      int     `json:"losses"`
	Kills       int     `json:"kills"`
	Deaths      int     `json:"deaths"`
	AirKills    int     `json:"air_kills"`
	GroundKills int     `json:"ground_kills"`
	WinRate     float64 `json:"win_rate"`
	KDR         float64 `json:"kd_ratio"`
}

type VehicleRef struct {
	Name string `json:"name"`
	Type string `json:"type"`
	BR   string `json:"br"`
}

type Squadron struct {
	Name        string     `json:"name"`
	Tag         string     `json:"tag"`
	Members     int        `json:"members"`
	Description string     `json:"description"`
	Leader      string     `json:"leader"`
}
