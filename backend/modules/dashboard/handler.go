package dashboard

import (
	"baseadmin/backend/modules/peserta"
	"baseadmin/backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler is the HTTP layer for the dashboard module.
type Handler struct {
	Service *Service
}

// NewHandler takes the already-constructed peserta.Service (rather than
// building its own) so the dashboard's attendance-status regenerate reuses
// the exact same business logic peserta's own CRUD paths use, instead of a
// second copy of recomputeAttendanceStatus's rules living in this module.
func NewHandler(db *gorm.DB, pesertaService *peserta.Service) *Handler {
	return &Handler{Service: NewService(NewRepository(db), pesertaService)}
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
