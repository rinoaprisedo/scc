package sliders

import (
	"errors"
	"mime/multipart"
	"strconv"

	"baseadmin/backend/modules/activity_logs"
	"baseadmin/backend/storage"
	"baseadmin/backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler is the HTTP layer for the sliders module.
type Handler struct {
	Service *Service
}

func NewHandler(db *gorm.DB, s storage.StorageInterface) *Handler {
	return &Handler{Service: NewService(NewRepository(db), s)}
}

// List godoc
// @Summary		List sliders
// @Tags			sliders
// @Security		SessionCookie
// @Param			page	query		int		false	"Page number"
// @Param			limit	query		int		false	"Page size"
// @Param			type	query		string	false	"Filter by type (mobile/desktop)"
// @Success		200		{object}	utils.Response
// @Router			/sliders [get]
func (h *Handler) List(c *gin.Context) {
	p := utils.GetPagination(c, "order")
	list, total, err := h.Service.List(p, c.Query("type"))
	if err != nil {
		utils.Error(c, 500, "failed to fetch sliders")
		return
	}
	utils.SuccessList(c, "ok", list, utils.BuildMeta(p.Page, p.Limit, total))
}

// Active godoc
// @Summary		List every active slider, unpaginated (for the website carousel)
// @Tags			sliders
// @Security		SessionCookie
// @Success		200	{object}	utils.Response
// @Router			/sliders/active [get]
func (h *Handler) Active(c *gin.Context) {
	list, err := h.Service.Active()
	if err != nil {
		utils.Error(c, 500, "failed to fetch sliders")
		return
	}
	utils.Success(c, 200, "ok", list)
}

// Get godoc
// @Summary		Get a slider
// @Tags			sliders
// @Security		SessionCookie
// @Param			uuid	path		string	true	"Slider UUID"
// @Success		200		{object}	utils.Response
// @Failure		404		{object}	utils.Response
// @Router			/sliders/{uuid} [get]
func (h *Handler) Get(c *gin.Context) {
	slider, err := h.Service.Get(c.Param("uuid"))
	if err != nil {
		utils.Error(c, 404, "slider not found")
		return
	}
	utils.Success(c, 200, "ok", slider)
}

// Create godoc
// @Summary		Create a slider
// @Tags			sliders
// @Security		SessionCookie
// @Accept			multipart/form-data
// @Param			image	formData	file	true	"Slider image"
// @Param			type	formData	string	true	"Slider type (mobile/desktop)"
// @Success		201		{object}	utils.Response
// @Failure		400		{object}	utils.Response
// @Router			/sliders [post]
func (h *Handler) Create(c *gin.Context) {
	fileHeader, err := c.FormFile("image")
	if err != nil {
		utils.Error(c, 400, "image is required")
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		utils.Error(c, 400, "failed to read uploaded file")
		return
	}
	defer file.Close()

	actorID := utils.CurrentUserID(c)
	slider, err := h.Service.Create(actorID, c.PostForm("type"), file, fileHeader)
	if err != nil {
		if errors.Is(err, ErrInvalidType) {
			utils.Error(c, 400, "type must be mobile or desktop")
			return
		}
		utils.Error(c, 400, "failed to create slider")
		return
	}
	activity_logs.LogActivity(actorID, activity_logs.ActionCreate, "sliders", slider.UUID.String(), nil, slider, c.ClientIP(), c.Request.UserAgent())
	utils.Success(c, 201, "slider created", slider)
}

// Update godoc
// @Summary		Update a slider
// @Tags			sliders
// @Security		SessionCookie
// @Accept			multipart/form-data
// @Param			uuid		path		string	true	"Slider UUID"
// @Param			image		formData	file	false	"Slider image"
// @Param			type		formData	string	true	"Slider type (mobile/desktop)"
// @Param			is_active	formData	bool	false	"Active status"
// @Success		200			{object}	utils.Response
// @Failure		404			{object}	utils.Response
// @Router			/sliders/{uuid} [put]
func (h *Handler) Update(c *gin.Context) {
	var file multipart.File
	var fileHeader *multipart.FileHeader
	if fh, err := c.FormFile("image"); err == nil {
		f, err := fh.Open()
		if err != nil {
			utils.Error(c, 400, "failed to read uploaded file")
			return
		}
		defer f.Close()
		file = f
		fileHeader = fh
	}

	isActive, _ := strconv.ParseBool(c.DefaultPostForm("is_active", "true"))

	actorID := utils.CurrentUserID(c)
	before, after, err := h.Service.Update(c.Param("uuid"), UpdateInput{Type: c.PostForm("type"), IsActive: isActive, ActorID: actorID}, file, fileHeader)
	if err != nil {
		if errors.Is(err, ErrInvalidType) {
			utils.Error(c, 400, "type must be mobile or desktop")
			return
		}
		utils.Error(c, 404, "slider not found")
		return
	}
	activity_logs.LogActivity(actorID, activity_logs.ActionUpdate, "sliders", after.UUID.String(), before, after, c.ClientIP(), c.Request.UserAgent())
	utils.Success(c, 200, "slider updated", after)
}

// Delete godoc
// @Summary		Delete a slider
// @Tags			sliders
// @Security		SessionCookie
// @Param			uuid	path	string	true	"Slider UUID"
// @Success		200		{object}	utils.Response
// @Failure		404		{object}	utils.Response
// @Router			/sliders/{uuid} [delete]
func (h *Handler) Delete(c *gin.Context) {
	actorID := utils.CurrentUserID(c)
	slider, err := h.Service.Delete(c.Param("uuid"), actorID)
	if err != nil {
		utils.Error(c, 404, "slider not found")
		return
	}
	activity_logs.LogActivity(actorID, activity_logs.ActionDelete, "sliders", slider.UUID.String(), slider, nil, c.ClientIP(), c.Request.UserAgent())
	utils.Success(c, 200, "slider deleted", nil)
}

type reorderItem struct {
	UUID  string `json:"uuid" binding:"required"`
	Order int    `json:"order"`
}

// Reorder godoc
// @Summary		Reorder sliders
// @Tags			sliders
// @Security		SessionCookie
// @Param			body	body		[]reorderItem	true	"Ordered list of slider UUIDs"
// @Success		200		{object}	utils.Response
// @Router			/sliders/reorder [put]
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
