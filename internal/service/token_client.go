package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

type TokenData struct {
	TurnstileToken string `json:"turnstile_token"`
	CfClearance    string `json:"cf_clearance"`
	UpdatedAt      string `json:"updated_at"`
}

var (
	tokenFile  = "/root/project/wtserver/token.json"
	tokenData  *TokenData
	tokenMu    sync.RWMutex
)

func loadTokenFile() {
	tokenMu.Lock()
	defer tokenMu.Unlock()

	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return
	}
	var t TokenData
	if err := json.Unmarshal(data, &t); err != nil {
		return
	}
	tokenData = &t
}

func SaveToken(token string, cfClearance string) error {
	tokenMu.Lock()
	defer tokenMu.Unlock()

	data := TokenData{
		TurnstileToken: token,
		CfClearance:    cfClearance,
		UpdatedAt:      time.Now().Format(time.RFC3339),
	}

	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(tokenFile, bytes, 0644); err != nil {
		return err
	}

	tokenData = &data
	return nil
}

func GetToken() (string, error) {
	tokenMu.RLock()
	if tokenData != nil && tokenData.TurnstileToken != "" {
		updatedAt, _ := time.Parse(time.RFC3339, tokenData.UpdatedAt)
		if time.Since(updatedAt) < 4*time.Hour {
			defer tokenMu.RUnlock()
			return tokenData.TurnstileToken, nil
		}
	}
	tokenMu.RUnlock()

	loadTokenFile()

	tokenMu.RLock()
	defer tokenMu.RUnlock()

	if tokenData == nil || tokenData.TurnstileToken == "" {
		return "", fmt.Errorf("no token available, please set token first")
	}

	updatedAt, _ := time.Parse(time.RFC3339, tokenData.UpdatedAt)
	if time.Since(updatedAt) > 4*time.Hour {
		return "", fmt.Errorf("token expired, please refresh")
	}

	return tokenData.TurnstileToken, nil
}

func GetCfClearance() string {
	tokenMu.RLock()
	defer tokenMu.RUnlock()
	if tokenData != nil {
		return tokenData.CfClearance
	}
	return ""
}

func init() {
	loadTokenFile()
}

func MakeStatsharkRequest(url string, token string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")

	if token != "" {
		req.Header.Set("X-Turnstile-Token", token)
	}

	if cf := GetCfClearance(); cf != "" {
		req.Header.Set("Cookie", "cf_clearance="+cf)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 406 {
		return nil, fmt.Errorf("statshark requires valid token (406)")
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("statshark status %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}
