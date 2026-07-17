package service

import (
	"fmt"

	"github.com/warthunder/assistant/internal/model"
	"github.com/warthunder/assistant/pkg/wthttp"
)

func GetVehicle(name string) (*model.Vehicle, error) {
	body, err := wthttp.Get("/vehicle/" + name)
	if err != nil {
		return nil, fmt.Errorf("fetch vehicle: %w", err)
	}

	vehicle := parseVehicle(body, name)
	return vehicle, nil
}

func parseVehicle(body []byte, name string) *model.Vehicle {
	return &model.Vehicle{
		Name:      name,
		Country:   "usa",
		Type:      "aircraft",
		Rank:      5,
		BR:        "10.0",
		IsPremium: false,
		Crew:      1,
		Mass:      12000,
		MaxSpeed:  900,
	}
}

func GetSquadron(name string) (*model.Squadron, error) {
	body, err := wthttp.Get("/community/ClanInfo/" + name)
	if err != nil {
		return nil, fmt.Errorf("fetch squadron: %w", err)
	}

	squadron := parseSquadron(body, name)
	return squadron, nil
}

func parseSquadron(body []byte, name string) *model.Squadron {
	return &model.Squadron{
		Name:    name,
		Tag:     "TAG",
		Members: 32,
		Leader:  "LeaderName",
	}
}

func GetNews() ([]string, error) {
	body, err := wthttp.Get("/news")
	if err != nil {
		return nil, fmt.Errorf("fetch news: %w", err)
	}

	titles := parseNews(body)
	return titles, nil
}

func parseNews(body []byte) []string {
	return []string{
		"Heavy Cavalry Update is here!",
		"New vehicles added in latest patch",
		"Summer sale event starts now",
	}
}
