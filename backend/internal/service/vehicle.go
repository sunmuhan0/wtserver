package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/warthunder/assistant/internal/model"
)

var wtVehicleAPI = "https://wtvehiclesapi.duckdns.org"

type wtVehicleResp struct {
	Identifier      string  `json:"identifier"`
	Country         string  `json:"country"`
	VehicleType     string  `json:"vehicle_type"`
	Era             int     `json:"era"`
	ArcadeBr        float64 `json:"arcade_br"`
	RealisticBr     float64 `json:"realistic_br"`
	SimulatorBr     float64 `json:"simulator_br"`
	IsPremium       int     `json:"is_premium"`
	Value           float64 `json:"value"`
	ReqExp          float64 `json:"req_exp"`
	SquadronVehicle int     `json:"squadron_vehicle"`
}

func GetVehicle(name string) (*model.Vehicle, error) {
	url := fmt.Sprintf("%s/api/vehicles?search=%s", wtVehicleAPI, name)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("vehicle api unreachable: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var vehicles []wtVehicleResp
	if err := json.Unmarshal(body, &vehicles); err != nil {
		return nil, fmt.Errorf("parse vehicle: %w", err)
	}
	if len(vehicles) == 0 {
		return nil, fmt.Errorf("vehicle not found")
	}
	v := vehicles[0]

	return &model.Vehicle{
		Name:      v.Identifier,
		Country:   v.Country,
		Type:      v.VehicleType,
		Rank:      v.Era,
		BR:        fmt.Sprintf("%.1f/%.1f/%.1f", v.ArcadeBr, v.RealisticBr, v.SimulatorBr),
		IsPremium: v.IsPremium == 1,
	}, nil
}

func GetSquadron(name string) (*model.Squadron, error) {
	return nil, fmt.Errorf("squadron search unavailable: warthunder.com is behind Cloudflare")
}

func GetNews() ([]model.NewsItem, error) {
	url := fmt.Sprintf("%s/api/news?limit=10", wtVehicleAPI)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("news service unavailable: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if len(body) == 0 {
		return nil, fmt.Errorf("news service returned empty response")
	}

	var items []model.NewsItem
	if err := json.Unmarshal(body, &items); err == nil {
		return items, nil
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("news service unavailable: unexpected response format")
	}

	if list, ok := raw["news"].([]interface{}); ok {
		for _, n := range list {
			if m, ok := n.(map[string]interface{}); ok {
				items = append(items, model.NewsItem{
					Title: fmt.Sprintf("%v", m["title"]),
					URL:   fmt.Sprintf("%v", m["url"]),
					Date:  fmt.Sprintf("%v", m["date"]),
				})
			}
		}
		return items, nil
	}

	return nil, fmt.Errorf("news service unavailable: unexpected response format")
}
