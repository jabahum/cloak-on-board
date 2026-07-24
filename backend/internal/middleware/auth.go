package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jabahum/keycloak-onboarder/backend/internal/auth"
)

func APIKeyAuth(expectedToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		expectedToken = strings.TrimSpace(expectedToken)

		if expectedToken == "" {
			auth.SetUser(c, auth.User{Subject: "development", Username: "development", DisplayName: "Development Admin", RealmRoles: []string{"admin"}})
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		token := strings.TrimPrefix(authHeader, "Bearer ")
		token = strings.TrimSpace(token)

		if token == "" || token != expectedToken {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":      "unauthorized",
				"request_id": GetRequestID(c),
			})
			return
		}

		auth.SetUser(c, auth.User{Subject: "api-key", Username: "api-key", DisplayName: "API Key Admin", RealmRoles: []string{"admin"}})
		c.Next()
	}
}
