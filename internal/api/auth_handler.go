package api

import "github.com/gin-gonic/gin"

// Login is a placeholder endpoint for future JWT auth integration.
func (h *Handler) Login(c *gin.Context) {
	c.JSON(200, gin.H{"token": "demo-token", "note": "auth is simulated"})
}
