package handler

import (
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var dataServiceURL = func() string {
	url := os.Getenv("DATA_SERVICE_URL")
	if url == "" {
		url = "http://127.0.0.1:3001"
	}
	return url
}()

func proxy(target string) gin.HandlerFunc {
	return func(c *gin.Context) {
		resp, err := http.Get(dataServiceURL + target)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "data service unavailable: " + err.Error()})
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		c.Data(resp.StatusCode, "application/json; charset=utf-8", body)
	}
}

func GetPlayer(c *gin.Context) {
	proxy("/api/v1/player/" + c.Param("nickname"))(c)
}

func GetSquadron(c *gin.Context) {
	proxy("/api/v1/squadron/" + c.Param("name"))(c)
}

func GetVehicle(c *gin.Context) {
	proxy("/api/v1/vehicle/" + c.Param("name"))(c)
}

func GetNews(c *gin.Context) {
	proxy("/api/v1/news")(c)
}
