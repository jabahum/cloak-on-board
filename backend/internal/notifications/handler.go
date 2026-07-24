package notifications

import (
	"github.com/gin-gonic/gin"
	"github.com/jabahum/keycloak-onboarder/backend/internal/auth"
	"github.com/jabahum/keycloak-onboarder/backend/internal/response"
	"net/http"
)

type Handler struct{ service *Service }

func NewHandler(s *Service) *Handler { return &Handler{service: s} }
func user(c *gin.Context) (auth.User, bool) {
	u, ok := auth.GetUser(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required")
	}
	return u, ok
}
func (h *Handler) List(c *gin.Context) {
	u, ok := user(c)
	if !ok {
		return
	}
	items, err := h.service.List(c.Request.Context(), u.Subject)
	if err != nil {
		response.Error(c, 500, "failed to list notifications")
		return
	}
	response.OK(c, items)
}
func (h *Handler) Unread(c *gin.Context) {
	u, ok := user(c)
	if !ok {
		return
	}
	n, err := h.service.UnreadCount(c.Request.Context(), u.Subject)
	if err != nil {
		response.Error(c, 500, "failed to count notifications")
		return
	}
	response.OK(c, gin.H{"count": n})
}
func (h *Handler) Read(c *gin.Context) {
	u, ok := user(c)
	if !ok {
		return
	}
	if err := h.service.MarkRead(c.Request.Context(), u.Subject, c.Param("id")); err != nil {
		response.Error(c, 404, "notification not found")
		return
	}
	response.OK(c, gin.H{"message": "notification marked read"})
}
func (h *Handler) ReadAll(c *gin.Context) {
	u, ok := user(c)
	if !ok {
		return
	}
	if err := h.service.MarkAllRead(c.Request.Context(), u.Subject); err != nil {
		response.Error(c, 500, "failed to mark notifications read")
		return
	}
	response.OK(c, gin.H{"message": "notifications marked read"})
}
