package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/warthunder/assistant/config"
	"github.com/warthunder/assistant/internal/handler"
	"github.com/warthunder/assistant/internal/service"
)

func main() {
	cfg := config.Load()
	service.SetCaptchaKey(cfg.CaptchaAPIKey)

	log.Println("starting browser for statshark API...")
	if err := service.StartBrowser(); err != nil {
		log.Printf("WARNING: browser start failed: %v (will retry on first request)", err)
	}
	defer service.StopBrowser()

	r := gin.Default()

	h := handler.New(cfg)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/player-ts/:nickname", h.GetPlayerTS)
		v1.GET("/player-search/:nickname", h.SearchPlayer)
		v1.GET("/squadron/:name", h.GetSquadron)
		v1.GET("/player-ss/:nickname", h.GetPlayerSS)
		v1.GET("/player-search-ss/:nickname", h.SearchPlayerSS)
		v1.GET("/player-leaderboard-ss/:nickname", h.GetLeaderboardHistorySS)
		v1.GET("/globalstats", h.GetGlobalStats)
		v1.GET("/vehicle/:name", h.GetVehicle)
		v1.GET("/vehicles", h.ListVehicles)
		v1.GET("/vehicle-filters", h.GetFilters)
		v1.GET("/news", h.GetNews)
		v1.GET("/news/detail", h.GetNewsDetail)
		v1.POST("/token", h.SetToken)
		v1.GET("/token/status", h.GetTokenStatus)
	}

	v3 := r.Group("/api/v3")
	{
		v3.GET("/player-ss/:nickname", h.GetPlayerSSV3)
		v3.GET("/player-detail/:nickname", h.GetPlayerDetailV3)
		v3.GET("/player-search-ss/:nickname", h.SearchPlayerSSV3)
		v3.GET("/player-leaderboard-ss/:nickname", h.GetLeaderboardHistorySSV3)
	}

	addr := ":" + cfg.Port
	log.Printf("server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
