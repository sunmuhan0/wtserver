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

func ListVehicles(c *gin.Context) {
	country := c.DefaultQuery("country", "")
	vtype := c.DefaultQuery("type", "")
	search := c.DefaultQuery("search", "")
	results := service.ListVehicles(country, vtype, search, 50)
	c.JSON(http.StatusOK, results)
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
