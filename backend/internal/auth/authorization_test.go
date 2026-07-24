package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestEffectiveRolePrecedence(t *testing.T) {
	user := User{RealmRoles: []string{"viewer", "manager", "admin"}}
	if got := user.EffectiveRole(); got != "admin" {
		t.Fatalf("EffectiveRole() = %q", got)
	}
}

func TestRolePermissions(t *testing.T) {
	if !HasPermission(User{RealmRoles: []string{"viewer"}}, PermissionRead) {
		t.Fatal("viewer should have read permission")
	}
	if HasPermission(User{RealmRoles: []string{"viewer"}}, PermissionManageDrafts) {
		t.Fatal("viewer should not manage drafts")
	}
	if HasPermission(User{RealmRoles: []string{"manager"}}, PermissionReviewApproval) {
		t.Fatal("manager should not review approvals")
	}
	if !HasPermission(User{RealmRoles: []string{"admin"}}, PermissionViewAudit) {
		t.Fatal("admin should view audit")
	}
}

func TestRequirePermissionEnforcesBackendAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/admin", func(c *gin.Context) {
		SetUser(c, User{Subject: "viewer", RealmRoles: []string{"viewer"}})
		c.Next()
	}, RequirePermission(PermissionAdminClients), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/admin", nil))
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", recorder.Code)
	}
}
