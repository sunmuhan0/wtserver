package model

type Vehicle struct {
	Name          string        `json:"name"`
	Country       string        `json:"country"`
	Type          string        `json:"type"`
	Rank          int           `json:"rank"`
	BR            string        `json:"br"`
	IsPremium     bool          `json:"is_premium"`
	IsHidden      bool          `json:"is_hidden"`
	Crew          int           `json:"crew"`
	Mass          float64       `json:"mass"`
	EnginePower   int           `json:"engine_power"`
	MaxSpeed      float64       `json:"max_speed"`
	Armaments     []Armament    `json:"armaments"`
	Modifications []Modification `json:"modifications"`
}

type Armament struct {
	Name   string `json:"name"`
	Caliber string `json:"caliber"`
	Amount int    `json:"amount"`
}

type Modification struct {
	Name  string `json:"name"`
	Cost  int    `json:"cost"`
	RP    int    `json:"rp"`
}
