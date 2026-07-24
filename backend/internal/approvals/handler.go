package approvals

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jabahum/keycloak-onboarder/backend/internal/applications"
	"github.com/jabahum/keycloak-onboarder/backend/internal/auth"
	"github.com/jabahum/keycloak-onboarder/backend/internal/response"
	"net/http"
)

type Handler struct{ service *Service }

func NewHandler(s *Service) *Handler { return &Handler{service: s} }
func (h *Handler) Submit(c *gin.Context) {
	var input SubmitRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	u, _ := auth.GetUser(c)
	item, err := h.service.Submit(c.Request.Context(), c.Param("id"), u, input)
	if err != nil {
		write(c, err)
		return
	}
	response.Created(c, item)
}
func (h *Handler) List(c *gin.Context) {
	u, _ := auth.GetUser(c)
	items, err := h.service.List(c.Request.Context(), u)
	if err != nil {
		write(c, err)
		return
	}
	response.OK(c, items)
}
func (h *Handler) Get(c *gin.Context) {
	u, _ := auth.GetUser(c)
	item, err := h.service.Get(c.Request.Context(), c.Param("id"))
	if err == nil && u.EffectiveRole() != "admin" && item.RequestedBySubject != u.Subject {
		err = ErrForbidden
	}
	if err != nil {
		write(c, err)
		return
	}
	response.OK(c, item)
}
func (h *Handler) Approve(c *gin.Context) {
	var input DecisionRequest
	_ = c.ShouldBindJSON(&input)
	u, _ := auth.GetUser(c)
	item, err := h.service.Approve(c.Request.Context(), c.Param("id"), u, input.Comment)
	if err != nil {
		write(c, err)
		return
	}
	response.OK(c, item)
}
func (h *Handler) Reject(c *gin.Context) {
	var input DecisionRequest
	_ = c.ShouldBindJSON(&input)
	u, _ := auth.GetUser(c)
	item, err := h.service.Reject(c.Request.Context(), c.Param("id"), u, input.Comment)
	if err != nil {
		write(c, err)
		return
	}
	response.OK(c, item)
}
func (h *Handler) Cancel(c *gin.Context) {
	var input DecisionRequest
	_ = c.ShouldBindJSON(&input)
	u, _ := auth.GetUser(c)
	item, err := h.service.Cancel(c.Request.Context(), c.Param("id"), u, input.Comment)
	if err != nil {
		write(c, err)
		return
	}
	response.OK(c, item)
}
func (h *Handler) Retry(c *gin.Context) {
	u, _ := auth.GetUser(c)
	item, err := h.service.Retry(c.Request.Context(), c.Param("id"), u)
	if err != nil {
		write(c, err)
		return
	}
	response.OK(c, item)
}
func write(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound), errors.Is(err, applications.ErrNotFound):
		response.Error(c, 404, err.Error())
	case errors.Is(err, ErrForbidden):
		response.Error(c, 403, err.Error())
	case errors.Is(err, ErrConflict):
		response.Error(c, 409, err.Error())
	case errors.Is(err, applications.ErrValidation):
		response.Error(c, 422, err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, "approval operation failed")
	}
}
