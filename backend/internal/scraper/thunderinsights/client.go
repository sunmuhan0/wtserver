package thunderinsights

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const DefaultBaseURL = "https://api.thunderinsights.dk"

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	APIKey     string
}

type PlayerStats struct {
	Nickname  string       `json:"nickname"`
	AvatarURL string       `json:"avatar_url"`
	Clan      string       `json:"clan"`
	Level     int          `json:"level"`
	Modes     []ModeStats  `json:"modes"`
	Vehicles  []VehicleStat `json:"vehicles"`
}

type ModeStats struct {
	Mode        string  `json:"mode"`
	Battles     int     `json:"battles"`
	WinRate     float64 `json:"win_rate"`
	KillsPerDeath float64 `json:"kills_per_death"`
	KillsPerBattle float64 `json:"kills_per_battle"`
	Deaths      int     `json:"deaths"`
	AirKills    int     `json:"air_kills"`
	GroundKills int     `json:"ground_kills"`
	NavalKills  int     `json:"naval_kills"`
	Efficiency  float64 `json:"efficiency"`
}

type VehicleStat struct {
	Identifier  string  `json:"identifier"`
	Name        string  `json:"name"`
	Battles     int     `json:"battles"`
	WinRate     float64 `json:"win_rate"`
	KillsPerDeath float64 `json:"kills_per_death"`
	Efficiency  float64 `json:"efficiency"`
}

func NewClient(apiKey string) *Client {
	if apiKey != "" {
		return &Client{
			BaseURL:    DefaultBaseURL,
			HTTPClient: &http.Client{Timeout: 30 * time.Second},
			APIKey:     apiKey,
		}
	}
	return &Client{
		BaseURL:    DefaultBaseURL,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "WarThunderStats/1.0")
	req.Header.Set("Accept", "application/json")
	if c.APIKey != "" {
		req.Header.Set("X-API-Key", c.APIKey)
	}
}

func (c *Client) GetPlayer(nickname string) (*PlayerStats, error) {
	url := fmt.Sprintf("%s/api/v1/player/%s", c.BaseURL, strings.ToLower(nickname))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request player %s: %w", nickname, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("player %s not found", nickname)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned status %d for player %s", resp.StatusCode, nickname)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read player %s: %w", nickname, err)
	}

	var stats PlayerStats
	if err := json.Unmarshal(body, &stats); err != nil {
		return nil, fmt.Errorf("parse player %s: %w", nickname, err)
	}
	return &stats, nil
}

func (c *Client) GetVehicle(identifier string) (*VehicleStat, error) {
	url := fmt.Sprintf("%s/api/v1/vehicle/%s", c.BaseURL, identifier)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request vehicle %s: %w", identifier, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned status %d for vehicle %s", resp.StatusCode, identifier)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read vehicle %s: %w", identifier, err)
	}

	var stat VehicleStat
	if err := json.Unmarshal(body, &stat); err != nil {
		return nil, fmt.Errorf("parse vehicle %s: %w", identifier, err)
	}
	return &stat, nil
}
