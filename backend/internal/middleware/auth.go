package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func APIKeyAuth(expectedToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		expectedToken = strings.TrimSpace(expectedToken)

		if expectedToken == "" {
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

		c.Next()
	}
}
