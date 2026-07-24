package audit

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jabahum/keycloak-onboarder/backend/internal/response"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }
func (h *Handler) List(c *gin.Context) {
	items, err := h.service.List(c.Request.Context(), Filter{
		Actor: c.Query("actor"), Action: c.Query("action"), ResourceType: c.Query("resource_type"),
		ApplicationID: c.Query("application_id"), Result: c.Query("result"), From: c.Query("from"), To: c.Query("to"),
		Page: ParseInt(c.Query("page"), 1), PageSize: ParseInt(c.Query("page_size"), 25),
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list audit logs")
		return
	}
	response.OK(c, items)
}
func (h *Handler) Get(c *gin.Context) {
	item, err := h.service.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusNotFound, "audit log not found")
		return
	}
	response.OK(c, item)
}
