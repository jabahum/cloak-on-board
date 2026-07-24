package applications

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jabahum/keycloak-onboarder/backend/internal/auth"
	"github.com/jabahum/keycloak-onboarder/backend/internal/keycloak"
	"github.com/jabahum/keycloak-onboarder/backend/internal/response"
)

type ClientManager interface {
	ListKeycloakClients(context.Context, string) ([]keycloak.KeycloakClient, error)
	ImportClient(context.Context, ImportApplicationRequest) (Application, error)
	UpdateApplication(context.Context, string, UpdateApplicationRequest) (Application, error)
	DeleteApplication(context.Context, string, bool) error
	GetClientScopes(context.Context, string) (keycloak.ClientScopeAssignments, error)
	AssignClientScope(context.Context, string, string, string) error
	RemoveClientScope(context.Context, string, string, string) error
	ListProtocolMappers(context.Context, string) ([]keycloak.ProtocolMapper, error)
	CreateProtocolMapper(context.Context, string, keycloak.ProtocolMapperRequest) error
	UpdateProtocolMapper(context.Context, string, string, keycloak.ProtocolMapperRequest) error
	DeleteProtocolMapper(context.Context, string, string) error
}

type Handler struct {
	service     *Service
	provisionFn func(ctx context.Context, applicationID string) (any, error)
	manager     ClientManager
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/applications", auth.RequirePermission(auth.PermissionRead), h.List)
	rg.POST("/applications", auth.RequirePermission(auth.PermissionManageDrafts), h.Create)
	rg.POST("/applications/import", auth.RequirePermission(auth.PermissionManageDrafts), h.Import)
	rg.GET("/applications/:id", auth.RequirePermission(auth.PermissionRead), h.GetByID)
	rg.PUT("/applications/:id", auth.RequirePermission(auth.PermissionManageDrafts), h.Update)
	rg.DELETE("/applications/:id", auth.RequirePermission(auth.PermissionAdminClients), h.Delete)
	rg.POST("/applications/:id/provision", auth.RequirePermission(auth.PermissionAdminClients), h.Provision)
	rg.GET("/keycloak/clients", auth.RequirePermission(auth.PermissionManageDrafts), h.ListKeycloakClients)
	rg.GET("/applications/:id/client-scopes", auth.RequirePermission(auth.PermissionRead), h.ListClientScopes)
	rg.PUT("/applications/:id/client-scopes/:scopeId", auth.RequirePermission(auth.PermissionAdminClients), h.AssignClientScope)
	rg.DELETE("/applications/:id/client-scopes/:scopeId", auth.RequirePermission(auth.PermissionAdminClients), h.RemoveClientScope)
	rg.GET("/applications/:id/protocol-mappers", auth.RequirePermission(auth.PermissionRead), h.ListProtocolMappers)
	rg.POST("/applications/:id/protocol-mappers", auth.RequirePermission(auth.PermissionAdminClients), h.CreateProtocolMapper)
	rg.PUT("/applications/:id/protocol-mappers/:mapperId", auth.RequirePermission(auth.PermissionAdminClients), h.UpdateProtocolMapper)
	rg.DELETE("/applications/:id/protocol-mappers/:mapperId", auth.RequirePermission(auth.PermissionAdminClients), h.DeleteProtocolMapper)
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) SetProvisioner(fn func(ctx context.Context, applicationID string) (any, error)) {
	h.provisionFn = fn
}

func (h *Handler) SetClientManager(manager ClientManager) {
	h.manager = manager
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

func (h *Handler) Update(c *gin.Context) {
	var req UpdateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	current, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.writeError(c, err)
		return
	}
	user, _ := auth.GetUser(c)
	if user.EffectiveRole() != "admin" && current.KeycloakClientUUID != "" {
		response.Error(c, http.StatusForbidden, "linked client updates require an approval request")
		return
	}
	if current.Status == "pending_approval" {
		response.Error(c, http.StatusConflict, "application has a pending approval request")
		return
	}
	app, err := h.manager.UpdateApplication(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, app)
}

func (h *Handler) Import(c *gin.Context) {
	var req ImportApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	app, err := h.manager.ImportClient(c.Request.Context(), req)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Created(c, app)
}

func (h *Handler) Delete(c *gin.Context) {
	deleteKeycloak := c.Query("delete_keycloak") == "true"
	if err := h.manager.DeleteApplication(c.Request.Context(), c.Param("id"), deleteKeycloak); err != nil {
		h.writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) ListKeycloakClients(c *gin.Context) {
	clients, err := h.manager.ListKeycloakClients(c.Request.Context(), c.Query("search"))
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, clients)
}

func (h *Handler) ListClientScopes(c *gin.Context) {
	scopes, err := h.manager.GetClientScopes(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, scopes)
}

func (h *Handler) AssignClientScope(c *gin.Context) {
	var req AssignClientScopeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.manager.AssignClientScope(c.Request.Context(), c.Param("id"), c.Param("scopeId"), req.Type); err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, gin.H{"message": "client scope assigned"})
}

func (h *Handler) RemoveClientScope(c *gin.Context) {
	if err := h.manager.RemoveClientScope(c.Request.Context(), c.Param("id"), c.Param("scopeId"), c.Query("type")); err != nil {
		h.writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) ListProtocolMappers(c *gin.Context) {
	mappers, err := h.manager.ListProtocolMappers(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, mappers)
}

func (h *Handler) CreateProtocolMapper(c *gin.Context) {
	var req keycloak.ProtocolMapperRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.manager.CreateProtocolMapper(c.Request.Context(), c.Param("id"), req); err != nil {
		h.writeError(c, err)
		return
	}
	response.Created(c, gin.H{"message": "protocol mapper created"})
}

func (h *Handler) UpdateProtocolMapper(c *gin.Context) {
	var req keycloak.ProtocolMapperRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.manager.UpdateProtocolMapper(c.Request.Context(), c.Param("id"), c.Param("mapperId"), req); err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, gin.H{"message": "protocol mapper updated"})
}

func (h *Handler) DeleteProtocolMapper(c *gin.Context) {
	if err := h.manager.DeleteProtocolMapper(c.Request.Context(), c.Param("id"), c.Param("mapperId")); err != nil {
		h.writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound), errors.Is(err, keycloak.ErrNotFound):
		response.Error(c, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrConflict):
		response.Error(c, http.StatusConflict, err.Error())
	case errors.Is(err, ErrValidation):
		response.Error(c, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, ErrNotProvisioned):
		response.Error(c, http.StatusConflict, err.Error())
	default:
		response.Error(c, http.StatusBadGateway, err.Error())
	}
}
