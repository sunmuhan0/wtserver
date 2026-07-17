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
		api.GET("/player/:nickname", handler.GetPlayer)
		api.GET("/squadron/:name", handler.GetSquadron)
		api.GET("/vehicle/:name", handler.GetVehicle)
		api.GET("/news", handler.GetNews)
	}

	addr := ":" + cfg.Port
	log.Printf("server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
