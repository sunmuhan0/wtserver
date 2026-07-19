package model

type GlobalStats struct {
	Nations []string         `json:"nations"`
	Types   []string         `json:"types"`
	Cells   []GlobalStatCell `json:"cells"`
}

type GlobalStatCell struct {
	Nation      string  `json:"nation"`
	Type        string  `json:"type"`
	Count       int     `json:"count"`
	AvgBR       float64 `json:"avg_br"`
	WinRate     float64 `json:"win_rate"`
	GamesPlayed int     `json:"games_played"`
	PlayerCount int     `json:"player_count"`
}
