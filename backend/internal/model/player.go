package model

type Squadron struct {
	Name        string `json:"name"`
	Tag         string `json:"tag"`
	Members     int    `json:"members"`
	Description string `json:"description"`
	Leader      string `json:"leader"`
}
