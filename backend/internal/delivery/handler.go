package delivery

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jabahum/keycloak-onboarder/backend/internal/applications"
	"github.com/jabahum/keycloak-onboarder/backend/internal/auth"
	"github.com/jabahum/keycloak-onboarder/backend/internal/credentials"
	"github.com/jabahum/keycloak-onboarder/backend/internal/response"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/environments", auth.RequirePermission(auth.PermissionRead), h.listEnvironments)
	rg.POST("/environments", auth.RequirePermission(auth.PermissionManageEnvironments), h.createEnvironment)
	rg.PUT("/environments/:id", auth.RequirePermission(auth.PermissionManageEnvironments), h.updateEnvironment)
	rg.DELETE("/environments/:id", auth.RequirePermission(auth.PermissionManageEnvironments), h.deleteEnvironment)
	rg.GET("/realm-connections", auth.RequirePermission(auth.PermissionRead), h.listConnections)
	rg.POST("/realm-connections", auth.RequirePermission(auth.PermissionManageRealmConnections), h.createConnection)
	rg.PUT("/realm-connections/:id", auth.RequirePermission(auth.PermissionManageRealmConnections), h.updateConnection)
	rg.POST("/realm-connections/:id/test", auth.RequirePermission(auth.PermissionManageRealmConnections), h.testConnection)
	rg.DELETE("/realm-connections/:id", auth.RequirePermission(auth.PermissionManageRealmConnections), h.disableConnection)

	rg.GET("/applications/:id/snapshots", auth.RequirePermission(auth.PermissionRead), h.listSnapshots)
	rg.POST("/applications/:id/snapshots", auth.RequirePermission(auth.PermissionManageDrafts), h.createSnapshot)
	rg.GET("/applications/:id/deployments", auth.RequirePermission(auth.PermissionRead), h.listApplicationDeployments)
	rg.POST("/applications/:id/deployments", auth.RequirePermission(auth.PermissionAdminClients), h.deploy)
	rg.GET("/deployments", auth.RequirePermission(auth.PermissionRead), h.listDeployments)
	rg.POST("/applications/:id/promotions", auth.RequirePermission(auth.PermissionPromoteApplications), h.promote)
	rg.POST("/deployments/:id/rollback", auth.RequirePermission(auth.PermissionAdminClients), h.rollback)

	rg.POST("/deployments/:id/drift-checks", auth.RequirePermission(auth.PermissionCheckDrift), h.checkDrift)
	rg.GET("/drift-runs", auth.RequirePermission(auth.PermissionRead), h.listDriftRuns)
	rg.POST("/deployments/:id/reconcile", auth.RequirePermission(auth.PermissionSubmitApproval), h.reconcile)
	rg.POST("/deployments/:id/rotate-secret", auth.RequirePermission(auth.PermissionRotateSecrets), h.rotateSecret)
	rg.POST("/secret-deliveries/:id/consume", auth.RequirePermission(auth.PermissionRead), h.consumeSecret)
}

func (h *Handler) listEnvironments(c *gin.Context) {
	items, err := h.service.ListEnvironments(c.Request.Context())
	h.result(c, items, err, false)
}
func (h *Handler) createEnvironment(c *gin.Context) {
	var input CreateEnvironmentRequest
	if !bind(c, &input) {
		return
	}
	item, err := h.service.CreateEnvironment(c.Request.Context(), input)
	h.result(c, item, err, true)
}
func (h *Handler) updateEnvironment(c *gin.Context) {
	var input CreateEnvironmentRequest
	if !bind(c, &input) {
		return
	}
	item, err := h.service.UpdateEnvironment(c.Request.Context(), c.Param("id"), input)
	h.result(c, item, err, false)
}
func (h *Handler) deleteEnvironment(c *gin.Context) {
	if err := h.service.DeleteEnvironment(c.Request.Context(), c.Param("id")); err != nil {
		h.writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
func (h *Handler) listConnections(c *gin.Context) {
	items, err := h.service.ListConnections(c.Request.Context())
	h.result(c, items, err, false)
}
func (h *Handler) createConnection(c *gin.Context) {
	var input CreateConnectionRequest
	if !bind(c, &input) {
		return
	}
	item, err := h.service.CreateConnection(c.Request.Context(), input)
	h.result(c, item, err, true)
}
func (h *Handler) updateConnection(c *gin.Context) {
	var input UpdateConnectionRequest
	if !bind(c, &input) {
		return
	}
	item, err := h.service.UpdateConnection(c.Request.Context(), c.Param("id"), input)
	h.result(c, item, err, false)
}
func (h *Handler) testConnection(c *gin.Context) {
	item, err := h.service.TestConnection(c.Request.Context(), c.Param("id"))
	h.result(c, item, err, false)
}
func (h *Handler) disableConnection(c *gin.Context) {
	if err := h.service.DisableConnection(c.Request.Context(), c.Param("id")); err != nil {
		h.writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
func (h *Handler) createSnapshot(c *gin.Context) {
	user, _ := auth.GetUser(c)
	item, err := h.service.CreateSnapshot(c.Request.Context(), c.Param("id"), user)
	h.result(c, item, err, true)
}
func (h *Handler) listSnapshots(c *gin.Context) {
	items, err := h.service.ListSnapshots(c.Request.Context(), c.Param("id"))
	h.result(c, items, err, false)
}
func (h *Handler) listApplicationDeployments(c *gin.Context) {
	items, err := h.service.ListDeployments(c.Request.Context(), c.Param("id"))
	h.result(c, items, err, false)
}
func (h *Handler) listDeployments(c *gin.Context) {
	items, err := h.service.ListDeployments(c.Request.Context(), c.Query("application_id"))
	h.result(c, items, err, false)
}
func (h *Handler) deploy(c *gin.Context) {
	var input CreateDeploymentRequest
	if !bind(c, &input) {
		return
	}
	user, _ := auth.GetUser(c)
	item, _, err := h.service.Deploy(c.Request.Context(), c.Param("id"), input, user.Subject, false)
	h.result(c, item, err, true)
}
func (h *Handler) promote(c *gin.Context) {
	var input PromotionRequest
	if !bind(c, &input) {
		return
	}
	user, _ := auth.GetUser(c)
	item, _, err := h.service.Promote(c.Request.Context(), c.Param("id"), input, user.Subject, false)
	h.result(c, item, err, true)
}
func (h *Handler) rollback(c *gin.Context) {
	user, _ := auth.GetUser(c)
	item, _, err := h.service.Rollback(c.Request.Context(), c.Param("id"), user.Subject, false)
	h.result(c, item, err, false)
}
func (h *Handler) checkDrift(c *gin.Context) {
	user, _ := auth.GetUser(c)
	item, err := h.service.CheckDrift(c.Request.Context(), c.Param("id"), user.Subject)
	h.result(c, item, err, true)
}
func (h *Handler) listDriftRuns(c *gin.Context) {
	items, err := h.service.ListDriftRuns(c.Request.Context(), c.Query("deployment_id"))
	h.result(c, items, err, false)
}
func (h *Handler) reconcile(c *gin.Context) {
	response.Error(c, http.StatusForbidden, "reconciliation requires an approved reconcile_drift request")
}
func (h *Handler) rotateSecret(c *gin.Context) {
	response.Error(c, http.StatusForbidden, "secret rotation requires an approved rotate_client_secret request")
}
func (h *Handler) consumeSecret(c *gin.Context) {
	user, _ := auth.GetUser(c)
	secret, delivery, err := h.service.ConsumeSecret(c.Request.Context(), c.Param("id"), user.Subject)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.Header("Cache-Control", "no-store, max-age=0")
	c.Header("Pragma", "no-cache")
	response.OK(c, gin.H{"secret": secret, "delivery": delivery})
}

func bind(c *gin.Context, target any) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		response.Error(c, http.StatusBadRequest, "malformed request")
		return false
	}
	return true
}

func (h *Handler) result(c *gin.Context, value any, err error, created bool) {
	if err != nil {
		h.writeError(c, err)
		return
	}
	if created {
		response.Created(c, value)
	} else {
		response.OK(c, value)
	}
}

func (h *Handler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound), errors.Is(err, applications.ErrNotFound):
		response.Error(c, http.StatusNotFound, "resource not found")
	case errors.Is(err, ErrForbidden):
		response.Error(c, http.StatusForbidden, err.Error())
	case errors.Is(err, ErrConflict):
		response.Error(c, http.StatusConflict, err.Error())
	case errors.Is(err, ErrValidation):
		response.Error(c, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, ErrGone):
		c.Header("Cache-Control", "no-store, max-age=0")
		response.Error(c, http.StatusGone, err.Error())
	case errors.Is(err, ErrUnavailable):
		response.Error(c, http.StatusServiceUnavailable, "target environment is unavailable")
	case errors.Is(err, credentials.ErrNoKeys):
		response.Error(c, http.StatusInternalServerError, "credential encryption is unavailable")
	default:
		response.Error(c, http.StatusInternalServerError, "delivery operation failed")
	}
}
