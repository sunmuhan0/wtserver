package service

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type BrowserClient struct {
	proxyURL   string
	mu         sync.RWMutex
	ready      bool
	token      string
	lastRefresh time.Time
}

var browser *BrowserClient

func StartBrowser() error {
	proxyURL := "http://127.0.0.1:8082"
	browser = &BrowserClient{proxyURL: proxyURL}

	log.Printf("[browser] checking proxy at %s...", proxyURL)
	for i := 0; i < 30; i++ {
		resp, err := http.Get(proxyURL + "/health")
		if err != nil {
			if i%5 == 0 {
				log.Printf("[browser] waiting for proxy... (%d/30)", i+1)
			}
			time.Sleep(2 * time.Second)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == 200 {
			browser.mu.Lock()
			browser.ready = true
			browser.token = "proxy-active"
			browser.lastRefresh = time.Now()
			browser.mu.Unlock()
			log.Printf("[browser] proxy ready!")
			return nil
		}
	}

	return fmt.Errorf("proxy not available at %s after 60s", proxyURL)
}

func StopBrowser() {
}

func GetBrowser() *BrowserClient {
	return browser
}

func (b *BrowserClient) Fetch(method, url string, headers map[string]string, body string) (int, string, error) {
	b.mu.RLock()
	if !b.ready {
		b.mu.RUnlock()
		return 0, "", fmt.Errorf("browser proxy not ready")
	}
	b.mu.RUnlock()

	reqURL := b.proxyURL + url
	var resp *http.Response
	var err error

	if method == "GET" {
		resp, err = http.Get(reqURL)
	} else {
		req, _ := http.NewRequest("POST", reqURL, bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		resp, err = http.DefaultTransport.RoundTrip(req)
	}

	if err != nil {
		return 0, "", fmt.Errorf("proxy request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(respBody), nil
}

func (b *BrowserClient) IsTokenValid() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.ready
}

func (b *BrowserClient) GetToken() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.token
}

func (b *BrowserClient) Refresh() error {
	log.Printf("[browser] requesting proxy session refresh...")
	resp, err := http.Get(b.proxyURL + "/refresh")
	if err != nil {
		return fmt.Errorf("refresh request: %w", err)
	}
	resp.Body.Close()

	time.Sleep(30 * time.Second)
	return nil
}

func ensureBrowser() error {
	if browser == nil {
		return StartBrowser()
	}
	if !browser.IsTokenValid() {
		return browser.Refresh()
	}
	return nil
}
