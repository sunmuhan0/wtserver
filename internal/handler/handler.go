package handler

import (
	"fmt"
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

func ListVehicles(c *gin.Context) {
	country := c.DefaultQuery("country", "")
	vtype := c.DefaultQuery("type", "")
	search := c.DefaultQuery("search", "")
	offset := 0
	limit := 30
	fmt.Sscanf(c.DefaultQuery("offset", "0"), "%d", &offset)
	fmt.Sscanf(c.DefaultQuery("limit", "30"), "%d", &limit)
	results, total := service.ListVehicles(country, vtype, search, offset, limit)
	c.JSON(http.StatusOK, gin.H{"items": results, "total": total})
}

func GetFilters(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"countries": service.GetCountries(),
		"types":     service.GetTypes(),
	})
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
	lang := c.DefaultQuery("lang", "zh")
	news, err := service.GetNews(lang)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, news)
}

func GetPlayerSS(c *gin.Context) {
	data, err := service.GetPlayerSS(c.Param("nickname"), "")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func SearchPlayerSS(c *gin.Context) {
	data, err := service.SearchPlayerSS(c.Param("nickname"), "")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func GetLeaderboardHistorySS(c *gin.Context) {
	data, err := service.GetLeaderboardHistorySS(c.Param("nickname"), "")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func GetPlayerSSV3(c *gin.Context) {
	token := c.GetHeader("X-Turnstile-Token")
	data, err := service.GetPlayerSSV3(c.Param("nickname"), token)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func SearchPlayerSSV3(c *gin.Context) {
	token := c.GetHeader("X-Turnstile-Token")
	data, err := service.SearchPlayerSSV3(c.Param("nickname"), token)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func GetLeaderboardHistorySSV3(c *gin.Context) {
	token := c.GetHeader("X-Turnstile-Token")
	data, err := service.GetLeaderboardHistorySSV3(c.Param("nickname"), token)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func GetNewsDetail(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url is required"})
		return
	}
	detail, err := service.GetNewsDetail(url)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, detail)
}
