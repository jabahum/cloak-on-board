package provisioning

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
	rg.GET("/jobs", auth.RequirePermission(auth.PermissionRead), h.ListJobs)
	rg.GET("/jobs/:id", auth.RequirePermission(auth.PermissionRead), h.GetJobByID)
}

func (h *Handler) ListJobs(c *gin.Context) {
	jobs, err := h.service.ListJobs(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.OK(c, jobs)
}

func (h *Handler) GetJobByID(c *gin.Context) {
	job, err := h.service.GetJobByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	response.OK(c, job)
}

func (h *Handler) CreateJob(c *gin.Context) {
	var req CreateJobRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	job, err := h.service.CreateJob(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Created(c, job)
}
