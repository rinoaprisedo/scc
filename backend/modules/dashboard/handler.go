package dashboard

import (
	"baseadmin/backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler is the HTTP layer for the dashboard module.
type Handler struct {
	Service *Service
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{Service: NewService(NewRepository(db))}
}

// Summary godoc
// @Summary		Peserta dashboard summary
// @Tags			dashboard
// @Security		SessionCookie
// @Success		200	{object}	utils.Response
// @Router			/dashboard/summary [get]
func (h *Handler) Summary(c *gin.Context) {
	summary, err := h.Service.Summary()
	if err != nil {
		utils.Error(c, 500, "failed to load dashboard summary")
		return
	}
	utils.Success(c, 200, "ok", summary)
}
