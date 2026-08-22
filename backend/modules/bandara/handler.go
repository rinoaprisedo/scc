package bandara

import (
	"baseadmin/backend/modules/activity_logs"
	"baseadmin/backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler is the HTTP layer for the bandara module.
type Handler struct {
	Service *Service
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{Service: NewService(NewRepository(db))}
}

// List godoc
// @Summary		List bandara
// @Tags			bandara
// @Security		SessionCookie
// @Param			page	query		int		false	"Page number"
// @Param			limit	query		int		false	"Page size"
// @Param			search	query		string	false	"Search by name"
// @Success		200		{object}	utils.Response
// @Router			/bandara [get]
func (h *Handler) List(c *gin.Context) {
	p := utils.GetPagination(c, "name")
	list, total, err := h.Service.List(p)
	if err != nil {
		utils.Error(c, 500, "failed to fetch bandara")
		return
	}
	utils.SuccessList(c, "ok", list, utils.BuildMeta(p.Page, p.Limit, total))
}

// Options godoc
// @Summary		List every bandara, unpaginated (for dropdowns)
// @Tags			bandara
// @Security		SessionCookie
// @Success		200	{object}	utils.Response
// @Router			/bandara/options [get]
func (h *Handler) Options(c *gin.Context) {
	list, err := h.Service.Options()
	if err != nil {
		utils.Error(c, 500, "failed to fetch bandara")
		return
	}
	utils.Success(c, 200, "ok", list)
}

// Get godoc
// @Summary		Get a bandara
// @Tags			bandara
// @Security		SessionCookie
// @Param			uuid	path		string	true	"Bandara UUID"
// @Success		200		{object}	utils.Response
// @Failure		404		{object}	utils.Response
// @Router			/bandara/{uuid} [get]
func (h *Handler) Get(c *gin.Context) {
	bandara, err := h.Service.Get(c.Param("uuid"))
	if err != nil {
		utils.Error(c, 404, "bandara not found")
		return
	}
	utils.Success(c, 200, "ok", bandara)
}

type bandaraRequest struct {
	Name string `json:"name" binding:"required"`
}

// Create godoc
// @Summary		Create a bandara
// @Tags			bandara
// @Security		SessionCookie
// @Param			body	body		bandaraRequest	true	"Bandara payload"
// @Success		201		{object}	utils.Response
// @Router			/bandara [post]
func (h *Handler) Create(c *gin.Context) {
	var req bandaraRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "invalid request payload")
		return
	}
	actorID := utils.CurrentUserID(c)
	bandara, err := h.Service.Create(Input{Name: req.Name, ActorID: actorID})
	if err != nil {
		utils.Error(c, 400, "failed to create bandara")
		return
	}
	activity_logs.LogActivity(actorID, activity_logs.ActionCreate, "bandara", bandara.UUID.String(), nil, bandara, c.ClientIP(), c.Request.UserAgent())
	utils.Success(c, 201, "bandara created", bandara)
}

// Update godoc
// @Summary		Update a bandara
// @Tags			bandara
// @Security		SessionCookie
// @Param			uuid	path		string			true	"Bandara UUID"
// @Param			body	body		bandaraRequest	true	"Bandara payload"
// @Success		200		{object}	utils.Response
// @Failure		404		{object}	utils.Response
// @Router			/bandara/{uuid} [put]
func (h *Handler) Update(c *gin.Context) {
	var req bandaraRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "invalid request payload")
		return
	}
	actorID := utils.CurrentUserID(c)
	before, after, err := h.Service.Update(c.Param("uuid"), Input{Name: req.Name, ActorID: actorID})
	if err != nil {
		utils.Error(c, 404, "bandara not found")
		return
	}
	activity_logs.LogActivity(actorID, activity_logs.ActionUpdate, "bandara", after.UUID.String(), before, after, c.ClientIP(), c.Request.UserAgent())
	utils.Success(c, 200, "bandara updated", after)
}

// Delete godoc
// @Summary		Delete a bandara
// @Tags			bandara
// @Security		SessionCookie
// @Param			uuid	path	string	true	"Bandara UUID"
// @Success		200		{object}	utils.Response
// @Failure		404		{object}	utils.Response
// @Router			/bandara/{uuid} [delete]
func (h *Handler) Delete(c *gin.Context) {
	actorID := utils.CurrentUserID(c)
	bandara, err := h.Service.Delete(c.Param("uuid"), actorID)
	if err != nil {
		utils.Error(c, 404, "bandara not found")
		return
	}
	activity_logs.LogActivity(actorID, activity_logs.ActionDelete, "bandara", bandara.UUID.String(), bandara, nil, c.ClientIP(), c.Request.UserAgent())
	utils.Success(c, 200, "bandara deleted", nil)
}
