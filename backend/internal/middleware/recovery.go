package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		log.Printf(
			"[PANIC] request_id=%s path=%s error=%v",
			GetRequestID(c),
			c.Request.URL.Path,
			recovered,
		)

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":      "internal server error",
			"request_id": GetRequestID(c),
		})
	})
}
