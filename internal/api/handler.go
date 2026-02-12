package api

import (
	"github.com/gin-gonic/gin"

	"godockerui/internal/service"
)

type Handler struct {
	dockerService service.RuntimeService
}

func NewHandler(dockerService service.RuntimeService) *Handler {
	return &Handler{dockerService: dockerService}
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}
