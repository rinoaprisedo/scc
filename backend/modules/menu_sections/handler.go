package menu_sections

import (
	"baseadmin/backend/modules/activity_logs"
	"baseadmin/backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler is the HTTP layer for the menu_sections module.
type Handler struct {
	Service *Service
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{Service: NewService(NewRepository(db))}
}

// List godoc
// @Summary		List menu sections
// @Tags			menu-sections
// @Security		SessionCookie
// @Success		200	{object}	utils.Response
// @Router			/menu-sections [get]
func (h *Handler) List(c *gin.Context) {
	list, err := h.Service.List()
	if err != nil {
		utils.Error(c, 500, "failed to fetch menu sections")
		return
	}
	utils.Success(c, 200, "ok", list)
}

// Get godoc
// @Summary		Get a menu section
// @Tags			menu-sections
// @Security		SessionCookie
// @Param			uuid	path		string	true	"Menu section UUID"
// @Success		200		{object}	utils.Response
// @Failure		404		{object}	utils.Response
// @Router			/menu-sections/{uuid} [get]
func (h *Handler) Get(c *gin.Context) {
	section, err := h.Service.Get(c.Param("uuid"))
	if err != nil {
		utils.Error(c, 404, "menu section not found")
		return
	}
	utils.Success(c, 200, "ok", section)
}

type sectionRequest struct {
	Name  string `json:"name" binding:"required"`
	Icon  string `json:"icon"`
	Order int    `json:"order"`
}

// Create godoc
// @Summary		Create a menu section
// @Tags			menu-sections
// @Security		SessionCookie
// @Param			body	body		sectionRequest	true	"Menu section payload"
// @Success		201		{object}	utils.Response
// @Router			/menu-sections [post]
func (h *Handler) Create(c *gin.Context) {
	var req sectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "invalid request payload")
		return
	}
	actorID := utils.CurrentUserID(c)
	section, err := h.Service.Create(SectionInput{Name: req.Name, Icon: req.Icon, Order: req.Order, ActorID: actorID})
	if err != nil {
		utils.Error(c, 400, "failed to create menu section")
		return
	}
	activity_logs.LogActivity(actorID, activity_logs.ActionCreate, "menu_sections", section.UUID.String(), nil, section, c.ClientIP(), c.Request.UserAgent())
	utils.Success(c, 201, "menu section created", section)
}

// Update godoc
// @Summary		Update a menu section
// @Tags			menu-sections
// @Security		SessionCookie
// @Param			uuid	path		string			true	"Menu section UUID"
// @Param			body	body		sectionRequest	true	"Menu section payload"
// @Success		200		{object}	utils.Response
// @Failure		404		{object}	utils.Response
// @Router			/menu-sections/{uuid} [put]
func (h *Handler) Update(c *gin.Context) {
	var req sectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "invalid request payload")
		return
	}
	actorID := utils.CurrentUserID(c)
	before, after, err := h.Service.Update(c.Param("uuid"), SectionInput{Name: req.Name, Icon: req.Icon, Order: req.Order, ActorID: actorID})
	if err != nil {
		utils.Error(c, 404, "menu section not found")
		return
	}
	activity_logs.LogActivity(actorID, activity_logs.ActionUpdate, "menu_sections", after.UUID.String(), before, after, c.ClientIP(), c.Request.UserAgent())
	utils.Success(c, 200, "menu section updated", after)
}

// Delete godoc
// @Summary		Delete a menu section
// @Tags			menu-sections
// @Security		SessionCookie
// @Param			uuid	path	string	true	"Menu section UUID"
// @Success		200		{object}	utils.Response
// @Failure		404		{object}	utils.Response
// @Router			/menu-sections/{uuid} [delete]
func (h *Handler) Delete(c *gin.Context) {
	actorID := utils.CurrentUserID(c)
	section, err := h.Service.Delete(c.Param("uuid"), actorID)
	if err != nil {
		utils.Error(c, 404, "menu section not found")
		return
	}
	activity_logs.LogActivity(actorID, activity_logs.ActionDelete, "menu_sections", section.UUID.String(), section, nil, c.ClientIP(), c.Request.UserAgent())
	utils.Success(c, 200, "menu section deleted", nil)
}

type reorderItem struct {
	UUID  string `json:"uuid" binding:"required"`
	Order int    `json:"order"`
}

// Reorder godoc
// @Summary		Reorder menu sections
// @Tags			menu-sections
// @Security		SessionCookie
// @Param			body	body		[]reorderItem	true	"Ordered list of section UUIDs"
// @Success		200		{object}	utils.Response
// @Router			/menu-sections/reorder [put]
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
