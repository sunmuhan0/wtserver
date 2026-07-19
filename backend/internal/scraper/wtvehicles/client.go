package wtvehicles

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const DefaultBaseURL = "https://wtvehiclesapi.duckdns.org"

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

type Vehicle struct {
	ID                int     `json:"id"`
	Identifier        string  `json:"identifier"`
	Name              string  `json:"name"`
	Country           string  `json:"country"`
	VehicleType       string  `json:"vehicle_type"`
	Rank              int     `json:"rank"`
	ArcadeBR          float64 `json:"arcade_br"`
	RealisticBR       float64 `json:"realistic_br"`
	SimulatorBR       float64 `json:"simulator_br"`
	WinRate           float64 `json:"win_rate"`
	KillsPerDeath     float64 `json:"kills_per_death"`
	KillsPerBattle    float64 `json:"kills_per_battle"`
	BattlesPlayed     int     `json:"games_played"`
	Players           int     `json:"players"`
}

type Squadron struct {
	ID               int     `json:"id"`
	Name             string  `json:"name"`
	Tag              string  `json:"tag"`
	MembersCount     int     `json:"members_count"`
	Description      string  `json:"description"`
	LeaderID         int     `json:"leader_id"`
	LeaderNickname   string  `json:"leader_nickname"`
}

type NewsItem struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Summary   string    `json:"summary"`
	URL       string    `json:"url"`
	Published time.Time `json:"published"`
	ImageURL  string    `json:"image_url"`
}

func NewClient() *Client {
	return &Client{
		BaseURL:    DefaultBaseURL,
		HTTPClient: &http.Client{Timeout: 20 * time.Second},
	}
}

func (c *Client) GetVehicles() ([]Vehicle, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/api/vehicles")
	if err != nil {
		return nil, fmt.Errorf("get vehicles: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("vehicles API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read vehicles: %w", err)
	}

	var vehicles []Vehicle
	if err := json.Unmarshal(body, &vehicles); err != nil {
		return nil, fmt.Errorf("parse vehicles: %w", err)
	}
	return vehicles, nil
}

func (c *Client) GetVehicleByID(identifier string) (*Vehicle, error) {
	resp, err := c.HTTPClient.Get(fmt.Sprintf("%s/api/vehicles/%s", c.BaseURL, identifier))
	if err != nil {
		return nil, fmt.Errorf("get vehicle %s: %w", identifier, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("vehicle %s returned status %d", identifier, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read vehicle %s: %w", identifier, err)
	}

	var vehicle Vehicle
	if err := json.Unmarshal(body, &vehicle); err != nil {
		return nil, fmt.Errorf("parse vehicle %s: %w", identifier, err)
	}
	return &vehicle, nil
}

func (c *Client) GetSquadron(tag string) (*Squadron, error) {
	resp, err := c.HTTPClient.Get(fmt.Sprintf("%s/api/squadrons/%s", c.BaseURL, tag))
	if err != nil {
		return nil, fmt.Errorf("get squadron %s: %w", tag, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("squadron %s returned status %d", tag, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read squadron %s: %w", tag, err)
	}

	var squadron Squadron
	if err := json.Unmarshal(body, &squadron); err != nil {
		return nil, fmt.Errorf("parse squadron %s: %w", tag, err)
	}
	return &squadron, nil
}

func (c *Client) GetNews() ([]NewsItem, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/api/news")
	if err != nil {
		return nil, fmt.Errorf("get news: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("news API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read news: %w", err)
	}

	var news []NewsItem
	if err := json.Unmarshal(body, &news); err != nil {
		return nil, fmt.Errorf("parse news: %w", err)
	}
	return news, nil
}
