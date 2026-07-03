package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jabahum/keycloak-onboarder/backend/internal/middleware"
)

func (s *Server) registerRoutes() {
	api := s.router.Group("/api/v1")

	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": s.config.AppName,
			"env":     s.config.AppEnv,
		})
	})

	protected := api.Group("")
	protected.Use(middleware.APIKeyAuth(s.config.APIAuthToken))

	s.applicationHandler().RegisterRoutes(protected)
	s.templateHandler().RegisterRoutes(protected)
	s.settingsHandler().RegisterRoutes(protected)
	s.provisioningHandler().RegisterRoutes(protected)
}
