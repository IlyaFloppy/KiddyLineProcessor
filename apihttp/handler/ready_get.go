package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type status struct {
	Status string `json:"status"`
}

// ReadyGet returns handler for /ready to check if db is ready
func ReadyGet(ready *bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if *ready {
			c.JSON(http.StatusOK, status{
				Status: "ok",
			})
		} else {
			c.JSON(http.StatusServiceUnavailable, status{
				Status: "not ok",
			})
		}
	}
}
