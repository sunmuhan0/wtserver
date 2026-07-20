package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/warthunder/assistant/config"
	"github.com/warthunder/assistant/internal/handler"
)

func main() {
	cfg := config.Load()
	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		v1.GET("/player-ts/:nickname", handler.GetPlayerTS)
		v1.GET("/player-search/:nickname", handler.SearchPlayer)
		v1.GET("/squadron/:name", handler.GetSquadron)
		v1.GET("/player-ss/:nickname", handler.GetPlayerSS)
		v1.GET("/player-search-ss/:nickname", handler.SearchPlayerSS)
		v1.GET("/player-leaderboard-ss/:nickname", handler.GetLeaderboardHistorySS)
		v1.GET("/globalstats", handler.GetGlobalStats)
		v1.GET("/vehicle/:name", handler.GetVehicle)
		v1.GET("/vehicles", handler.ListVehicles)
		v1.GET("/vehicle-filters", handler.GetFilters)
		v1.GET("/news", handler.GetNews)
		v1.GET("/news/detail", handler.GetNewsDetail)
	}

	v3 := r.Group("/api/v3")
	{
		v3.GET("/player-ss/:nickname", handler.GetPlayerSSV3)
		v3.GET("/player-search-ss/:nickname", handler.SearchPlayerSSV3)
		v3.GET("/player-leaderboard-ss/:nickname", handler.GetLeaderboardHistorySSV3)
	}

	addr := ":" + cfg.Port
	log.Printf("server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
