package templates

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jabahum/keycloak-onboarder/backend/internal/auth"
	"github.com/jabahum/keycloak-onboarder/backend/internal/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/templates", auth.RequirePermission(auth.PermissionRead), h.List)
	rg.GET("/templates/:id", auth.RequirePermission(auth.PermissionRead), h.GetByID)
	rg.POST("/templates/seed", auth.RequirePermission(auth.PermissionManageSettings), h.SeedDefaults)
}

func (h *Handler) List(c *gin.Context) {
	items, err := h.service.List(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.OK(c, items)
}

func (h *Handler) GetByID(c *gin.Context) {
	item, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	response.OK(c, item)
}

func (h *Handler) SeedDefaults(c *gin.Context) {
	if err := h.service.SeedDefaults(c.Request.Context()); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.OK(c, gin.H{
		"message": "default templates seeded",
	})
}
