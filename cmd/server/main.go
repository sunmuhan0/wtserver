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

	api := r.Group("/api/v1")
	{
		api.GET("/player-ts/:nickname", handler.GetPlayerTS)
		api.GET("/player-search/:nickname", handler.SearchPlayer)
		api.GET("/squadron/:name", handler.GetSquadron)
		api.GET("/globalstats", handler.GetGlobalStats)
		api.GET("/vehicle/:name", handler.GetVehicle)
		api.GET("/vehicles", handler.ListVehicles)
		api.GET("/vehicle-filters", handler.GetFilters)
		api.GET("/news", handler.GetNews)
		api.GET("/news/detail", handler.GetNewsDetail)
	}

	addr := ":" + cfg.Port
	log.Printf("server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
