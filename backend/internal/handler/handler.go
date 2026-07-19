package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/warthunder/assistant/internal/service"
)

func GetPlayerTS(c *gin.Context) {
	stats, err := service.GetPlayerTS(c.Param("nickname"))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func SearchPlayer(c *gin.Context) {
	results, err := service.SearchPlayer(c.Param("nickname"))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

func GetVehicle(c *gin.Context) {
	vehicle, err := service.GetVehicle(c.Param("name"))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, vehicle)
}

func GetSquadron(c *gin.Context) {
	data, err := service.GetSquadron(c.Param("name"))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func GetGlobalStats(c *gin.Context) {
	data, err := service.GetGlobalStats()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func GetNews(c *gin.Context) {
	news, err := service.GetNews()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, news)
}
