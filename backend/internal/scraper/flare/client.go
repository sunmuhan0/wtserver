package flare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type FlareSolverrClient struct {
	BaseURL    string
	HTTPClient *http.Client
	MaxTimeout int
}

type FlareSolverrRequest struct {
	Cmd             string `json:"cmd"`
	URL             string `json:"url"`
	MaxTimeout      int    `json:"maxTimeout,omitempty"`
	SessionID       string `json:"session,omitempty"`
	SessionTTLMin   int    `json:"session_ttl_min,omitempty"`
	ReturnRawHTML   bool   `json:"returnRawHtml,omitempty"`
}

type FlareSolverrResponse struct {
	Status       string `json:"status"`
	Message      string `json:"message"`
	Solution     *struct {
		URL      string `json:"url"`
		Status   int    `json:"status"`
		Response string `json:"response"`
		Cookies  []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"cookies"`
		UserAgent string `json:"userAgent"`
	} `json:"solution"`
	Version      string `json:"version"`
}

func NewClient(baseURL string) *FlareSolverrClient {
	if baseURL == "" {
		baseURL = "http://localhost:8191"
	}
	return &FlareSolverrClient{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 120 * time.Second},
		MaxTimeout: 60000,
	}
}

func (f *FlareSolverrClient) Get(url string) (*http.Response, error) {
	body, _, err := f.getResponse(url, false)
	if err != nil {
		return nil, err
	}

	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader([]byte(body))),
	}, nil
}

func (f *FlareSolverrClient) GetRaw(url string) (string, *FlareSolverrResponse, error) {
	return f.getResponse(url, true)
}

func (f *FlareSolverrClient) getResponse(url string, rawHTML bool) (string, *FlareSolverrResponse, error) {
	reqBody := FlareSolverrRequest{
		Cmd:           "request.get",
		URL:           url,
		MaxTimeout:    f.MaxTimeout,
		ReturnRawHTML: rawHTML,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := f.HTTPClient.Post(
		fmt.Sprintf("%s/v1", f.BaseURL),
		"application/json",
		bytes.NewReader(jsonBody),
	)
	if err != nil {
		return "", nil, fmt.Errorf("flare request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("read response: %w", err)
	}

	var fsResp FlareSolverrResponse
	if err := json.Unmarshal(body, &fsResp); err != nil {
		return "", nil, fmt.Errorf("parse response: %w", err)
	}

	if fsResp.Status != "ok" {
		msg := fsResp.Message
		if msg == "" {
			msg = "unknown error"
		}
		return "", &fsResp, fmt.Errorf("flare error: %s", msg)
	}

	if fsResp.Solution == nil {
		return "", &fsResp, fmt.Errorf("no solution in response")
	}

	responseBody := fsResp.Solution.Response
	if rawHTML {
		return responseBody, &fsResp, nil
	}

	return responseBody, &fsResp, nil
}

func (f *FlareSolverrClient) CreateSession(sessionID string, ttlMin int) error {
	req := FlareSolverrRequest{
		Cmd:           "sessions.create",
		SessionID:     sessionID,
		SessionTTLMin: ttlMin,
	}
	return f.sendCommand(req)
}

func (f *FlareSolverrClient) DestroySession(sessionID string) error {
	req := FlareSolverrRequest{
		Cmd:       "sessions.destroy",
		SessionID: sessionID,
	}
	return f.sendCommand(req)
}

func (f *FlareSolverrClient) ListSessions() error {
	req := FlareSolverrRequest{
		Cmd: "sessions.list",
	}
	return f.sendCommand(req)
}

func (f *FlareSolverrClient) sendCommand(req FlareSolverrRequest) error {
	jsonBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	resp, err := f.HTTPClient.Post(
		fmt.Sprintf("%s/v1", f.BaseURL),
		"application/json",
		bytes.NewReader(jsonBody),
	)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	var fsResp FlareSolverrResponse
	if err := json.Unmarshal(body, &fsResp); err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	if fsResp.Status != "ok" {
		return fmt.Errorf("flare command failed: %s", fsResp.Message)
	}

	return nil
}
