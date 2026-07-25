package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jabahum/keycloak-onboarder/backend/internal/approvals"
	"github.com/jabahum/keycloak-onboarder/backend/internal/audit"
	"github.com/jabahum/keycloak-onboarder/backend/internal/auth"
	"github.com/jabahum/keycloak-onboarder/backend/internal/delivery"
	"github.com/jabahum/keycloak-onboarder/backend/internal/middleware"
	"github.com/jabahum/keycloak-onboarder/backend/internal/notifications"
	"github.com/jabahum/keycloak-onboarder/backend/internal/response"
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
	api.HEAD("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	protected := api.Group("")
	if s.config.AuthMode == "keycloak" {
		issuer := s.config.KeycloakPublicURL + "/realms/" + s.config.KeycloakRealm
		jwks := auth.NewJWKSProvider(s.config.KeycloakInternalURL, s.config.KeycloakRealm)
		protected.Use(auth.JWTAuth(jwks, issuer, s.config.KeycloakAudience))
	} else {
		protected.Use(middleware.APIKeyAuth(s.config.APIAuthToken))
	}
	auditService := audit.NewService(s.db)
	notificationService := notifications.NewService(s.db)
	protected.Use(audit.Middleware(auditService))

	protected.GET("/auth/me", auth.RequirePermission(auth.PermissionRead), func(c *gin.Context) {
		user, ok := auth.GetUser(c)
		if !ok {
			response.Error(c, http.StatusUnauthorized, "authentication required")
			return
		}
		if err := notificationService.UpsertUser(c.Request.Context(), user); err != nil {
			response.Error(c, http.StatusInternalServerError, "failed to load user profile")
			return
		}
		response.OK(c, gin.H{
			"subject": user.Subject, "username": user.Username, "email": user.Email,
			"display_name": user.DisplayName, "realm_roles": user.RealmRoles,
			"effective_role": user.EffectiveRole(),
		})
	})
	s.applicationHandler().RegisterRoutes(protected)
	s.templateHandler().RegisterRoutes(protected)
	s.settingsHandler().RegisterRoutes(protected)
	s.provisioningHandler().RegisterRoutes(protected)
	delivery.NewHandler(s.delivery).RegisterRoutes(protected)

	notificationHandler := notifications.NewHandler(notificationService)
	protected.GET("/notifications", auth.RequirePermission(auth.PermissionRead), notificationHandler.List)
	protected.GET("/notifications/unread-count", auth.RequirePermission(auth.PermissionRead), notificationHandler.Unread)
	protected.PUT("/notifications/:id/read", auth.RequirePermission(auth.PermissionRead), notificationHandler.Read)
	protected.PUT("/notifications/read-all", auth.RequirePermission(auth.PermissionRead), notificationHandler.ReadAll)

	auditHandler := audit.NewHandler(auditService)
	protected.GET("/audit-logs", auth.RequirePermission(auth.PermissionViewAudit), auditHandler.List)
	protected.GET("/audit-logs/:id", auth.RequirePermission(auth.PermissionViewAudit), auditHandler.Get)

	approvalHandler := approvals.NewHandler(s.approvalService(notificationService))
	protected.POST("/applications/:id/approval-requests", auth.RequirePermission(auth.PermissionSubmitApproval), approvalHandler.Submit)
	protected.GET("/approval-requests", auth.RequirePermission(auth.PermissionRead), approvalHandler.List)
	protected.GET("/approval-requests/:id", auth.RequirePermission(auth.PermissionRead), approvalHandler.Get)
	protected.POST("/approval-requests/:id/approve", auth.RequirePermission(auth.PermissionReviewApproval), approvalHandler.Approve)
	protected.POST("/approval-requests/:id/reject", auth.RequirePermission(auth.PermissionReviewApproval), approvalHandler.Reject)
	protected.POST("/approval-requests/:id/cancel", auth.RequirePermission(auth.PermissionSubmitApproval), approvalHandler.Cancel)
	protected.POST("/approval-requests/:id/retry", auth.RequirePermission(auth.PermissionReviewApproval), approvalHandler.Retry)
}
