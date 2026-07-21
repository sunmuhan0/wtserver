package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/warthunder/assistant/config"
	"github.com/warthunder/assistant/internal/service"
)

type Handler struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Handler {
	return &Handler{cfg: cfg}
}

func (h *Handler) GetPlayerTS(c *gin.Context) {
	stats, err := service.GetPlayerTS(c.Param("nickname"))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *Handler) SearchPlayer(c *gin.Context) {
	results, err := service.SearchPlayer(c.Param("nickname"))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

func (h *Handler) GetVehicle(c *gin.Context) {
	vehicle, err := service.GetVehicle(c.Param("name"))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, vehicle)
}

func (h *Handler) ListVehicles(c *gin.Context) {
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

func (h *Handler) GetFilters(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"countries": service.GetCountries(),
		"types":     service.GetTypes(),
	})
}

func (h *Handler) GetSquadron(c *gin.Context) {
	data, err := service.GetSquadron(c.Param("name"))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *Handler) GetGlobalStats(c *gin.Context) {
	data, err := service.GetGlobalStats()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *Handler) GetNews(c *gin.Context) {
	lang := c.DefaultQuery("lang", "zh")
	news, err := service.GetNews(lang)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, news)
}

func (h *Handler) GetPlayerSS(c *gin.Context) {
	data, err := service.GetPlayerSS(c.Param("nickname"), "")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *Handler) SearchPlayerSS(c *gin.Context) {
	data, err := service.SearchPlayerSS(c.Param("nickname"), "")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *Handler) GetLeaderboardHistorySS(c *gin.Context) {
	data, err := service.GetLeaderboardHistorySS(c.Param("nickname"), "")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *Handler) GetPlayerSSV3(c *gin.Context) {
	data, err := service.GetPlayerSSV3(c.Param("nickname"), "")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *Handler) SearchPlayerSSV3(c *gin.Context) {
	data, err := service.SearchPlayerSSV3(c.Param("nickname"), "")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *Handler) GetLeaderboardHistorySSV3(c *gin.Context) {
	data, err := service.GetLeaderboardHistorySSV3(c.Param("nickname"), "")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *Handler) GetPlayerDetailV3(c *gin.Context) {
	data, err := service.GetPlayerDetailV3(c.Param("nickname"), "")
	if err != nil {
		if err.Error() == "statshark api requires valid turnstile token (got 406)" {
			log.Printf("[handler] token expired, triggering background refresh...")
			service.TriggerBackgroundRefresh()
			c.JSON(http.StatusBadGateway, gin.H{"error": "token expired, refreshing in background. retry after a few seconds"})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *Handler) SetToken(c *gin.Context) {
	var req struct {
		Token       string `json:"token"`
		CfClearance string `json:"cf_clearance"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := service.SaveToken(req.Token, req.CfClearance); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) GetTokenStatus(c *gin.Context) {
	token, err := service.GetToken()
	hasToken := err == nil && token != ""
	c.JSON(http.StatusOK, gin.H{
		"has_token":       hasToken,
		"token_expired":   err != nil && err.Error() == "token expired",
		"has_captcha_key": h.cfg.CaptchaAPIKey != "",
	})
}

func (h *Handler) GetNewsDetail(c *gin.Context) {
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
