package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const capsolverBase = "https://api.capsolver.com"

type capsolverTask struct {
	Type                string `json:"type"`
	WebsiteURL          string `json:"websiteURL"`
	WebsiteKey          string `json:"websiteKey"`
	Metadata            *struct {
		Action string `json:"action"`
	} `json:"metadata,omitempty"`
}

type capsolverCreateReq struct {
	ClientKey string         `json:"clientKey"`
	Task      capsolverTask  `json:"task"`
}

type capsolverCreateResp struct {
	ErrorId    int    `json:"errorId"`
	ErrorMsg   string `json:"errorMsg"`
	TaskId     string `json:"taskId"`
}

type capsolverResultReq struct {
	ClientKey string `json:"clientKey"`
	TaskId    string `json:"taskId"`
}

type capsolverResultResp struct {
	ErrorId    int    `json:"errorId"`
	ErrorMsg   string `json:"errorMsg"`
	Status     string `json:"status"`
	Solution   struct {
		Token     string `json:"token"`
		UserAgent string `json:"userAgent"`
	} `json:"solution"`
}

func SolveTurnstile(apiKey string, siteKey string, pageURL string) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("capsolver API key not configured")
	}

	createBody, _ := json.Marshal(capsolverCreateReq{
		ClientKey: apiKey,
		Task: capsolverTask{
			Type:       "AntiTurnstileTaskProxyLess",
			WebsiteURL: pageURL,
			WebsiteKey: siteKey,
		},
	})

	resp, err := http.Post(capsolverBase+"/createTask", "application/json", bytes.NewReader(createBody))
	if err != nil {
		return "", fmt.Errorf("capsolver create task failed: %w", err)
	}
	defer resp.Body.Close()

	var createResp capsolverCreateResp
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		return "", fmt.Errorf("parse capsolver response: %w", err)
	}
	if createResp.ErrorId != 0 {
		return "", fmt.Errorf("capsolver error %d: %s", createResp.ErrorId, createResp.ErrorMsg)
	}
	if createResp.TaskId == "" {
		return "", fmt.Errorf("capsolver returned no task id")
	}

	for i := 0; i < 30; i++ {
		time.Sleep(2 * time.Second)

		resultBody, _ := json.Marshal(capsolverResultReq{
			ClientKey: apiKey,
			TaskId:    createResp.TaskId,
		})

		resultResp, err := http.Post(capsolverBase+"/getTaskResult", "application/json", bytes.NewReader(resultBody))
		if err != nil {
			continue
		}

		var result capsolverResultResp
		json.NewDecoder(resultResp.Body).Decode(&result)
		resultResp.Body.Close()

		if result.ErrorId != 0 {
			return "", fmt.Errorf("capsolver result error %d: %s", result.ErrorId, result.ErrorMsg)
		}
		if result.Status == "ready" && result.Solution.Token != "" {
			return result.Solution.Token, nil
		}
	}

	return "", fmt.Errorf("capsolver timeout waiting for solution")
}
