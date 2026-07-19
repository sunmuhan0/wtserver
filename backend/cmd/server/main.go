package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/warthunder/assistant/config"
	"github.com/warthunder/assistant/internal/handler"
)

func main() {
	cfg := config.Load()
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	api := r.Group("/api/v1")
	{
		api.GET("/player-ts/:nickname", handler.GetPlayerTS)
		api.GET("/player-search/:nickname", handler.SearchPlayer)
		api.GET("/squadron/:name", handler.GetSquadron)
		api.GET("/globalstats", handler.GetGlobalStats)
		api.GET("/vehicle/:name", handler.GetVehicle)
		api.GET("/news", handler.GetNews)
	}

	scraper := r.Group("/api/v2")
	{
		scraper.GET("/player/:nickname", handler.GetPlayerTSV2)
		scraper.GET("/player/:nickname/vehicles", handler.GetPlayerVehicles)
		scraper.GET("/player/:nickname/export", handler.GetPlayerExportJSON)
		scraper.GET("/vehicles", handler.GetVehicleIndex)
		scraper.GET("/vehicles/:slug", handler.GetVehicleDetail)
		scraper.GET("/wt/vehicles", handler.GetWTVehicles)
		scraper.GET("/wt/squadron/:tag", handler.GetWTSquadron)
		scraper.GET("/wt/news", handler.GetWTNews)
		scraper.GET("/flare/status", handler.GetFlareSolverrStatus)
	}

	addr := ":" + cfg.Port
	log.Printf("server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
