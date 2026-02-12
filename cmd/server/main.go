package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"godockerui/internal/api"
	"godockerui/internal/repository"
	"godockerui/internal/service"
)

func main() {
	dockerBin, dockerErr := repository.DockerBinaryPath()
	dockerService := service.NewDockerService(dockerBin, dockerErr)

	r := gin.Default()
	r.LoadHTMLGlob("web/templates/*")
	r.Static("/static", "web/static")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	h := api.NewHandler(dockerService)

	apiV1 := r.Group("/api/v1")
	{
		apiV1.GET("/dashboard", h.Dashboard)
		apiV1.GET("/containers", h.ListContainers)
		apiV1.POST("/containers/:id/start", h.StartContainer)
		apiV1.POST("/containers/:id/stop", h.StopContainer)
		apiV1.POST("/containers/:id/restart", h.RestartContainer)
		apiV1.GET("/images", h.ListImages)
		apiV1.GET("/health", h.Health)
	}

	if dockerErr != nil {
		log.Printf("docker CLI unavailable, using simulator data: %v", dockerErr)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("godockerui listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
