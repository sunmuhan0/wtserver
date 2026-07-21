package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

const tokenFile = "/root/project/wtserver/token.json"

const ssTurnstileSiteKey = "0x4AAAAAAA4JlzGNrS08fzpp"

type TokenData struct {
	TurnstileToken string `json:"turnstile_token"`
	CfClearance    string `json:"cf_clearance"`
	UpdatedAt      string `json:"updated_at"`
}

var (
	tokenData  *TokenData
	tokenMu    sync.RWMutex
	refreshMu  sync.Mutex
	refreshing bool
	captchaKey string
)

func SetCaptchaKey(key string) {
	captchaKey = key
}

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
		if time.Since(updatedAt) < 3*time.Hour {
			defer tokenMu.RUnlock()
			return tokenData.TurnstileToken, nil
		}
	}
	tokenMu.RUnlock()

	loadTokenFile()

	tokenMu.RLock()
	defer tokenMu.RUnlock()

	if tokenData == nil || tokenData.TurnstileToken == "" {
		return "", fmt.Errorf("no token available")
	}

	updatedAt, _ := time.Parse(time.RFC3339, tokenData.UpdatedAt)
	if time.Since(updatedAt) > 3*time.Hour {
		return "", fmt.Errorf("token expired")
	}

	return tokenData.TurnstileToken, nil
}

func RefreshToken() (string, error) {
	refreshMu.Lock()
	if refreshing {
		refreshMu.Unlock()
		return "", fmt.Errorf("token refresh already in progress")
	}
	refreshing = true
	refreshMu.Unlock()
	defer func() {
		refreshMu.Lock()
		refreshing = false
		refreshMu.Unlock()
	}()

	if captchaKey != "" {
		log.Printf("[token] attempting capsolver...")
		token, err := SolveTurnstile(captchaKey, ssTurnstileSiteKey, "https://statshark.net/")
		if err != nil {
			log.Printf("[token] capsolver failed: %v", err)
		} else if token != "" {
			_ = SaveToken(token, "")
			log.Printf("[token] turnstile token obtained via capsolver")
			return token, nil
		}
	}

	log.Printf("[token] attempting chromedp refresh...")

	defer func() {
		if r := recover(); r != nil {
			log.Printf("[token] chromedp panic recovered: %v", r)
		}
	}()

	os.Setenv("DISPLAY", ":99")

	opts := []chromedp.ExecAllocatorOption{
		chromedp.ExecPath("/snap/bin/chromium"),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("disable-features", "IsolateOrigins,site-per-process"),
		chromedp.Flag("enable-features", "NetworkService,NetworkServiceInProcess"),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.0.0 Safari/537.36"),
	}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	var turnstileToken string
	var cfClearance string

	log.Printf("[token] navigating to statshark.net...")
	err := chromedp.Run(ctx,
		chromedp.Navigate("https://statshark.net/player/224501637"),
		chromedp.WaitReady("body"),
		chromedp.Sleep(5*time.Second),
	)
	if err != nil {
		log.Printf("[token] navigate error: %v", err)
		return "", fmt.Errorf("chromedp navigate failed: %w", err)
	}

	log.Printf("[token] waiting for Turnstile widget to solve...")

	pollJS := `document.querySelector('input[name="cf-turnstile-response"]')?.value || ''`
	for i := 0; i < 12; i++ {
		time.Sleep(5 * time.Second)

		_ = chromedp.Run(ctx,
			chromedp.ActionFunc(func(ctx context.Context) error {
				var val string
				if err := chromedp.Evaluate(pollJS, &val).Do(ctx); err != nil {
					return nil
				}
				if val != "" {
					turnstileToken = val
				}
				return nil
			}),
		)

		log.Printf("[token] poll %d/12: turnstile=%v", i+1, turnstileToken != "")

		if turnstileToken != "" {
			break
		}

		if i == 3 {
			log.Printf("[token] trying player page...")
			_ = chromedp.Run(ctx,
				chromedp.Navigate("https://statshark.net/player/224501637"),
				chromedp.WaitReady("body"),
				chromedp.Sleep(3*time.Second),
			)
		}
	}

	_ = chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			cookies, err := network.GetCookies().Do(ctx)
			if err != nil {
				return err
			}
			for _, c := range cookies {
				if c.Name == "cf_clearance" {
					cfClearance = c.Value
					log.Printf("[token] got cf_clearance cookie")
				}
			}
			return nil
		}),
	)

	if turnstileToken != "" || cfClearance != "" {
		_ = SaveToken(turnstileToken, cfClearance)
		log.Printf("[token] saved: turnstile=%v, cf_clearance=%v", turnstileToken != "", cfClearance != "")
		return turnstileToken, nil
	}

	return "", fmt.Errorf("failed to get token from chromedp")
}

func GetTokenWithRefresh() (string, error) {
	token, err := GetToken()
	if err == nil {
		return token, nil
	}

	log.Printf("[token] token unavailable (%v), refreshing...", err)
	return RefreshToken()
}

func TriggerBackgroundRefresh() {
	refreshMu.Lock()
	if refreshing {
		refreshMu.Unlock()
		log.Printf("[token] background refresh already in progress, skipping")
		return
	}
	refreshMu.Unlock()

	go func() {
		log.Printf("[token] starting background token refresh...")
		_, err := RefreshToken()
		if err != nil {
			log.Printf("[token] background refresh failed: %v", err)
		} else {
			log.Printf("[token] background refresh completed successfully")
		}
	}()
}
