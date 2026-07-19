package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/warthunder/assistant/internal/scraper/flare"
	"github.com/warthunder/assistant/internal/scraper/thunderskill"
	"github.com/warthunder/assistant/internal/scraper/wtvehicles"
)

var (
	tsClient   = thunderskill.NewClient()
	wtvClient  = wtvehicles.NewClient()
	flareClient = flare.NewClient("")
)

func GetPlayerTSV2(c *gin.Context) {
	profile, err := tsClient.FetchPlayerProfile(c.Param("nickname"))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, profile)
}

func GetPlayerVehicles(c *gin.Context) {
	vehicles, err := tsClient.FetchPlayerVehicles(c.Param("nickname"))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, vehicles)
}

func GetPlayerExportJSON(c *gin.Context) {
	data, err := tsClient.FetchPlayerExportJSON(c.Param("nickname"))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func GetVehicleIndex(c *gin.Context) {
	mode := c.DefaultQuery("mode", "R")
	vehicleType, _ := strconv.Atoi(c.DefaultQuery("type", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	maxPages, _ := strconv.Atoi(c.DefaultQuery("pages", "5"))

	entries, err := tsClient.FetchVehicleIndex(mode, vehicleType, limit, maxPages)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, entries)
}

func GetVehicleDetail(c *gin.Context) {
	detail, err := tsClient.FetchVehicleDetail(c.Param("slug"))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, detail)
}

func GetWTVehicles(c *gin.Context) {
	vehicles, err := wtvClient.GetVehicles()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, vehicles)
}

func GetWTSquadron(c *gin.Context) {
	squadron, err := wtvClient.GetSquadron(c.Param("tag"))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, squadron)
}

func GetWTNews(c *gin.Context) {
	news, err := wtvClient.GetNews()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, news)
}

func GetFlareSolverrStatus(c *gin.Context) {
	_, _, err := flareClient.GetRaw("https://thunderskill.com/en/stat/nonexistent")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": "unavailable",
			"error":  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "available"})
}
