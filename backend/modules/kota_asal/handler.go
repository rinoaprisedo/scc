package kota_asal

import (
	"baseadmin/backend/modules/activity_logs"
	"baseadmin/backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler is the HTTP layer for the kota_asal module.
type Handler struct {
	Service *Service
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{Service: NewService(NewRepository(db))}
}

// List godoc
// @Summary		List kota asal
// @Tags			kota-asal
// @Security		SessionCookie
// @Param			page	query		int		false	"Page number"
// @Param			limit	query		int		false	"Page size"
// @Param			search	query		string	false	"Search by name/province"
// @Success		200		{object}	utils.Response
// @Router			/kota-asal [get]
func (h *Handler) List(c *gin.Context) {
	p := utils.GetPagination(c, "name")
	list, total, err := h.Service.List(p)
	if err != nil {
		utils.Error(c, 500, "failed to fetch kota asal")
		return
	}
	utils.SuccessList(c, "ok", list, utils.BuildMeta(p.Page, p.Limit, total))
}

// Options godoc
// @Summary		List every kota asal, unpaginated (for dropdowns)
// @Tags			kota-asal
// @Security		SessionCookie
// @Success		200	{object}	utils.Response
// @Router			/kota-asal/options [get]
func (h *Handler) Options(c *gin.Context) {
	list, err := h.Service.Options()
	if err != nil {
		utils.Error(c, 500, "failed to fetch kota asal")
		return
	}
	utils.Success(c, 200, "ok", list)
}

// Get godoc
// @Summary		Get a kota asal
// @Tags			kota-asal
// @Security		SessionCookie
// @Param			uuid	path		string	true	"Kota asal UUID"
// @Success		200		{object}	utils.Response
// @Failure		404		{object}	utils.Response
// @Router			/kota-asal/{uuid} [get]
func (h *Handler) Get(c *gin.Context) {
	kotaAsal, err := h.Service.Get(c.Param("uuid"))
	if err != nil {
		utils.Error(c, 404, "kota asal not found")
		return
	}
	utils.Success(c, 200, "ok", kotaAsal)
}

type kotaAsalRequest struct {
	Name        string `json:"name" binding:"required"`
	Province    string `json:"province"`
	Description string `json:"description"`
}

// Create godoc
// @Summary		Create a kota asal
// @Tags			kota-asal
// @Security		SessionCookie
// @Param			body	body		kotaAsalRequest	true	"Kota asal payload"
// @Success		201		{object}	utils.Response
// @Router			/kota-asal [post]
func (h *Handler) Create(c *gin.Context) {
	var req kotaAsalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "invalid request payload")
		return
	}
	actorID := utils.CurrentUserID(c)
	kotaAsal, err := h.Service.Create(Input{Name: req.Name, Province: req.Province, Description: req.Description, ActorID: actorID})
	if err != nil {
		utils.Error(c, 400, "failed to create kota asal")
		return
	}
	activity_logs.LogActivity(actorID, activity_logs.ActionCreate, "kota_asal", kotaAsal.UUID.String(), nil, kotaAsal, c.ClientIP(), c.Request.UserAgent())
	utils.Success(c, 201, "kota asal created", kotaAsal)
}

// Update godoc
// @Summary		Update a kota asal
// @Tags			kota-asal
// @Security		SessionCookie
// @Param			uuid	path		string			true	"Kota asal UUID"
// @Param			body	body		kotaAsalRequest	true	"Kota asal payload"
// @Success		200		{object}	utils.Response
// @Failure		404		{object}	utils.Response
// @Router			/kota-asal/{uuid} [put]
func (h *Handler) Update(c *gin.Context) {
	var req kotaAsalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "invalid request payload")
		return
	}
	actorID := utils.CurrentUserID(c)
	before, after, err := h.Service.Update(c.Param("uuid"), Input{Name: req.Name, Province: req.Province, Description: req.Description, ActorID: actorID})
	if err != nil {
		utils.Error(c, 404, "kota asal not found")
		return
	}
	activity_logs.LogActivity(actorID, activity_logs.ActionUpdate, "kota_asal", after.UUID.String(), before, after, c.ClientIP(), c.Request.UserAgent())
	utils.Success(c, 200, "kota asal updated", after)
}

// Delete godoc
// @Summary		Delete a kota asal
// @Tags			kota-asal
// @Security		SessionCookie
// @Param			uuid	path	string	true	"Kota asal UUID"
// @Success		200		{object}	utils.Response
// @Failure		404		{object}	utils.Response
// @Router			/kota-asal/{uuid} [delete]
func (h *Handler) Delete(c *gin.Context) {
	actorID := utils.CurrentUserID(c)
	kotaAsal, err := h.Service.Delete(c.Param("uuid"), actorID)
	if err != nil {
		utils.Error(c, 404, "kota asal not found")
		return
	}
	activity_logs.LogActivity(actorID, activity_logs.ActionDelete, "kota_asal", kotaAsal.UUID.String(), kotaAsal, nil, c.ClientIP(), c.Request.UserAgent())
	utils.Success(c, 200, "kota asal deleted", nil)
}
