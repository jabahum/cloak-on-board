package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Permission string

const (
	PermissionRead                   Permission = "read"
	PermissionManageDrafts           Permission = "manage_drafts"
	PermissionSubmitApproval         Permission = "submit_approval"
	PermissionAdminClients           Permission = "admin_clients"
	PermissionManageSettings         Permission = "manage_settings"
	PermissionReviewApproval         Permission = "review_approval"
	PermissionViewAudit              Permission = "view_audit"
	PermissionManageEnvironments     Permission = "manage_environments"
	PermissionManageRealmConnections Permission = "manage_realm_connections"
	PermissionPromoteApplications    Permission = "promote_applications"
	PermissionCheckDrift             Permission = "check_drift"
	PermissionReconcileDrift         Permission = "reconcile_drift"
	PermissionRotateSecrets          Permission = "rotate_secrets"
)

var rolePermissions = map[string]map[Permission]bool{
	"viewer": {
		PermissionRead: true,
	},
	"manager": {
		PermissionRead: true, PermissionManageDrafts: true, PermissionSubmitApproval: true,
		PermissionPromoteApplications: true, PermissionCheckDrift: true,
	},
	"admin": {
		PermissionRead: true, PermissionManageDrafts: true, PermissionSubmitApproval: true,
		PermissionAdminClients: true, PermissionManageSettings: true,
		PermissionReviewApproval: true, PermissionViewAudit: true,
		PermissionManageEnvironments: true, PermissionManageRealmConnections: true,
		PermissionPromoteApplications: true, PermissionCheckDrift: true,
		PermissionReconcileDrift: true, PermissionRotateSecrets: true,
	},
}

func HasPermission(user User, permission Permission) bool {
	return rolePermissions[user.EffectiveRole()][permission]
}

func RequirePermission(permission Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := GetUser(c)
		if !ok {
			unauthorized(c, "authentication required")
			return
		}
		if !HasPermission(user, permission) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "insufficient permission", "request_id": requestID(c),
			})
			return
		}
		c.Next()
	}
}
