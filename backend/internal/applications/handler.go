package applications

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jabahum/keycloak-onboarder/backend/internal/response"
)

type Handler struct {
	service     *Service
	provisionFn func(ctx context.Context, applicationID string) (any, error)
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/applications", h.List)
	rg.POST("/applications", h.Create)
	rg.GET("/applications/:id", h.GetByID)
	rg.DELETE("/applications/:id", h.Delete)
	rg.POST("/applications/:id/provision", h.Provision)
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) SetProvisioner(fn func(ctx context.Context, applicationID string) (any, error)) {
	h.provisionFn = fn
}

func (h *Handler) Provision(c *gin.Context) {
	if h.provisionFn == nil {
		response.Error(c, http.StatusInternalServerError, "provisioner is not configured")
		return
	}

	job, err := h.provisionFn(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.OK(c, job)
}

func (h *Handler) List(c *gin.Context) {
	apps, err := h.service.List(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.OK(c, apps)
}

func (h *Handler) GetByID(c *gin.Context) {
	app, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	response.OK(c, app)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateApplicationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	app, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Created(c, app)
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.OK(c, gin.H{
		"message": "application deleted",
	})
}
