package blazer_sizes

import (
	"baseadmin/backend/modules/activity_logs"
	"baseadmin/backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler is the HTTP layer for the blazer_sizes module.
type Handler struct {
	Service *Service
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{Service: NewService(NewRepository(db))}
}

// List godoc
// @Summary		List blazer sizes
// @Tags			blazer-sizes
// @Security		SessionCookie
// @Param			page	query		int		false	"Page number"
// @Param			limit	query		int		false	"Page size"
// @Param			search	query		string	false	"Search by size"
// @Success		200		{object}	utils.Response
// @Router			/blazer-sizes [get]
func (h *Handler) List(c *gin.Context) {
	p := utils.GetPagination(c, "order")
	list, total, err := h.Service.List(p)
	if err != nil {
		utils.Error(c, 500, "failed to fetch blazer sizes")
		return
	}
	utils.SuccessList(c, "ok", list, utils.BuildMeta(p.Page, p.Limit, total))
}

// Options godoc
// @Summary		List every blazer size, unpaginated (for dropdowns)
// @Tags			blazer-sizes
// @Security		SessionCookie
// @Success		200	{object}	utils.Response
// @Router			/blazer-sizes/options [get]
func (h *Handler) Options(c *gin.Context) {
	list, err := h.Service.Options()
	if err != nil {
		utils.Error(c, 500, "failed to fetch blazer sizes")
		return
	}
	utils.Success(c, 200, "ok", list)
}

// Get godoc
// @Summary		Get a blazer size
// @Tags			blazer-sizes
// @Security		SessionCookie
// @Param			uuid	path		string	true	"Blazer size UUID"
// @Success		200		{object}	utils.Response
// @Failure		404		{object}	utils.Response
// @Router			/blazer-sizes/{uuid} [get]
func (h *Handler) Get(c *gin.Context) {
	blazerSize, err := h.Service.Get(c.Param("uuid"))
	if err != nil {
		utils.Error(c, 404, "blazer size not found")
		return
	}
	utils.Success(c, 200, "ok", blazerSize)
}

type blazerSizeRequest struct {
	Size  string `json:"size" binding:"required"`
	Stock int    `json:"stock" binding:"gte=0"`
}

// Create godoc
// @Summary		Create a blazer size
// @Tags			blazer-sizes
// @Security		SessionCookie
// @Param			body	body		blazerSizeRequest	true	"Blazer size payload"
// @Success		201		{object}	utils.Response
// @Router			/blazer-sizes [post]
func (h *Handler) Create(c *gin.Context) {
	var req blazerSizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "invalid request payload")
		return
	}
	actorID := utils.CurrentUserID(c)
	blazerSize, err := h.Service.Create(Input{Size: req.Size, Stock: req.Stock, ActorID: actorID})
	if err != nil {
		utils.Error(c, 400, "failed to create blazer size")
		return
	}
	activity_logs.LogActivity(actorID, activity_logs.ActionCreate, "blazer_size", blazerSize.UUID.String(), nil, blazerSize, c.ClientIP(), c.Request.UserAgent())
	utils.Success(c, 201, "blazer size created", blazerSize)
}

// Update godoc
// @Summary		Update a blazer size
// @Tags			blazer-sizes
// @Security		SessionCookie
// @Param			uuid	path		string				true	"Blazer size UUID"
// @Param			body	body		blazerSizeRequest	true	"Blazer size payload"
// @Success		200		{object}	utils.Response
// @Failure		404		{object}	utils.Response
// @Router			/blazer-sizes/{uuid} [put]
func (h *Handler) Update(c *gin.Context) {
	var req blazerSizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "invalid request payload")
		return
	}
	actorID := utils.CurrentUserID(c)
	before, after, err := h.Service.Update(c.Param("uuid"), Input{Size: req.Size, Stock: req.Stock, ActorID: actorID})
	if err != nil {
		utils.Error(c, 404, "blazer size not found")
		return
	}
	activity_logs.LogActivity(actorID, activity_logs.ActionUpdate, "blazer_size", after.UUID.String(), before, after, c.ClientIP(), c.Request.UserAgent())
	utils.Success(c, 200, "blazer size updated", after)
}

// Delete godoc
// @Summary		Delete a blazer size
// @Tags			blazer-sizes
// @Security		SessionCookie
// @Param			uuid	path	string	true	"Blazer size UUID"
// @Success		200		{object}	utils.Response
// @Failure		404		{object}	utils.Response
// @Router			/blazer-sizes/{uuid} [delete]
func (h *Handler) Delete(c *gin.Context) {
	actorID := utils.CurrentUserID(c)
	blazerSize, err := h.Service.Delete(c.Param("uuid"), actorID)
	if err != nil {
		utils.Error(c, 404, "blazer size not found")
		return
	}
	activity_logs.LogActivity(actorID, activity_logs.ActionDelete, "blazer_size", blazerSize.UUID.String(), blazerSize, nil, c.ClientIP(), c.Request.UserAgent())
	utils.Success(c, 200, "blazer size deleted", nil)
}

type reorderItem struct {
	UUID  string `json:"uuid" binding:"required"`
	Order int    `json:"order"`
}

// Reorder godoc
// @Summary		Reorder blazer sizes
// @Tags			blazer-sizes
// @Security		SessionCookie
// @Param			body	body		[]reorderItem	true	"Ordered list of blazer size UUIDs"
// @Success		200		{object}	utils.Response
// @Router			/blazer-sizes/reorder [put]
func (h *Handler) Reorder(c *gin.Context) {
	var items []reorderItem
	if err := c.ShouldBindJSON(&items); err != nil {
		utils.Error(c, 400, "invalid request payload")
		return
	}
	serviceItems := make([]ReorderItem, 0, len(items))
	for _, item := range items {
		serviceItems = append(serviceItems, ReorderItem{UUID: item.UUID, Order: item.Order})
	}
	h.Service.Reorder(serviceItems)
	utils.Success(c, 200, "reordered", nil)
}
