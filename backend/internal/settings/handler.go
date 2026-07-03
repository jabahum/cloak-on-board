package settings

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jabahum/keycloak-onboarder/backend/internal/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/settings", h.Get)
	rg.PUT("/settings", h.Save)
}

func (h *Handler) Get(c *gin.Context) {
	item, err := h.service.Get(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.OK(c, item)
}

func (h *Handler) Save(c *gin.Context) {
	var req SaveSettingsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.service.Save(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.OK(c, item)
}
