package model

type NewsItem struct {
	URL   string `json:"url"`
	Title string `json:"title"`
	Image string `json:"image"`
	Date  string `json:"date"`
}

type Vehicle struct {
	Name        string  `json:"name"`
	Country     string  `json:"country"`
	Type        string  `json:"type"`
	Rank        int     `json:"rank"`
	BR          string  `json:"br"`
	IsPremium   bool    `json:"is_premium"`
	IsHidden    bool    `json:"is_hidden"`
	Crew        int     `json:"crew"`
	Mass        float64 `json:"mass"`
	EnginePower int     `json:"engine_power"`
	MaxSpeed    float64 `json:"max_speed"`
}
