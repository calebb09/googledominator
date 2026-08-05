package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"googledominator-backend/db"
)

func HealthCheck(c *gin.Context) {
	dbConnected := false
	if db.Instance != nil {
		dbConnected = db.Instance.IsConnected
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "online",
		"service":      "GoogleDominator API",
		"version":      "1.0.0",
		"db_connected": dbConnected,
		"timestamp":    time.Now().Format(time.RFC3339),
	})
}
