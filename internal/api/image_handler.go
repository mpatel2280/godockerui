package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListImages(c *gin.Context) {
	images, err := h.dockerService.ListImages(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, images)
}
