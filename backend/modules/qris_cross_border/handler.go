package qris_cross_border

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

// Handler is the HTTP layer for the qris_cross_border module.
type Handler struct {
	Service *Service
}

func NewHandler(db *gorm.DB, s storage.StorageInterface, ocrProvider, ocrAPIKey, ocrModel string) *Handler {
	return &Handler{Service: NewService(NewRepository(db), s, ocrProvider, ocrAPIKey, ocrModel)}
}

func parseFloatForm(c *gin.Context, key string) float64 {
	v, _ := strconv.ParseFloat(c.PostForm(key), 64)
	return v
}

// List godoc
// @Summary		List QRIS cross border submissions
// @Tags			qris-cross-border
// @Security		SessionCookie
// @Param			page	query		int		false	"Page number"
// @Param			limit	query		int		false	"Page size"
// @Param			search	query		string	false	"Search by peserta name/NIK"
// @Param			status			query		string	false	"Filter by status"
// @Param			peserta_uuid	query		string	false	"Filter by peserta"
// @Success		200				{object}	utils.Response
// @Router			/qris-cross-border [get]
func (h *Handler) List(c *gin.Context) {
	p := utils.GetPagination(c, "created_at")
	list, total, err := h.Service.List(p, c.Query("status"), c.Query("peserta_uuid"))
	if err != nil {
		utils.Error(c, 500, "failed to fetch qris cross border records")
		return
	}
	utils.SuccessList(c, "ok", list, utils.BuildMeta(p.Page, p.Limit, total))
}

// Get godoc
// @Summary		Get a QRIS cross border submission
// @Tags			qris-cross-border
// @Security		SessionCookie
// @Param			uuid	path		string	true	"Record UUID"
// @Success		200		{object}	utils.Response
// @Failure		404		{object}	utils.Response
// @Router			/qris-cross-border/{uuid} [get]
func (h *Handler) Get(c *gin.Context) {
	row, err := h.Service.Get(c.Param("uuid"))
	if err != nil {
		utils.Error(c, 404, "record not found")
		return
	}
	utils.Success(c, 200, "ok", row)
}

// Create godoc
// @Summary		Create a QRIS cross border submission
// @Tags			qris-cross-border
// @Security		SessionCookie
// @Accept			multipart/form-data
// @Param			peserta_uuid	formData	string	true	"Peserta UUID"
// @Param			image			formData	file	true	"Proof of payment image"
// @Success		201				{object}	utils.Response
// @Failure		400				{object}	utils.Response
// @Router			/qris-cross-border [post]
func (h *Handler) Create(c *gin.Context) {
	pesertaUUID := c.PostForm("peserta_uuid")
	if pesertaUUID == "" {
		utils.Error(c, 400, "peserta is required")
		return
	}
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
	row, err := h.Service.Create(CreateInput{
		PesertaUUID: pesertaUUID,
		ActorID:     actorID,
	}, file, fileHeader)
	if err != nil {
		if errors.Is(err, ErrPesertaNotFound) {
			utils.Error(c, 400, "peserta not found")
			return
		}
		utils.Error(c, 400, err.Error())
		return
	}

	activity_logs.LogActivity(actorID, activity_logs.ActionCreate, "qris_cross_border", row.UUID.String(), nil, row, c.ClientIP(), c.Request.UserAgent())
	utils.Success(c, 201, "qris cross border created", row)
}

// CreateSelf godoc
// @Summary		Submit a QRIS cross border proof of payment (participant self-service)
// @Tags			qris-cross-border
// @Security		SessionCookie
// @Accept			multipart/form-data
// @Param			image	formData	file	true	"Proof of payment image"
// @Success		201		{object}	utils.Response
// @Failure		400		{object}	utils.Response
// @Router			/qris-cross-border/my [post]
func (h *Handler) CreateSelf(c *gin.Context) {
	actorID := utils.CurrentUserID(c)
	if actorID == nil {
		utils.Error(c, 401, "unauthorized")
		return
	}
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

	row, err := h.Service.CreateSelf(*actorID, file, fileHeader)
	if err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	activity_logs.LogActivity(actorID, activity_logs.ActionCreate, "qris_cross_border", row.UUID.String(), nil, row, c.ClientIP(), c.Request.UserAgent())
	utils.Success(c, 201, "qris cross border submitted", row)
}

// MyList godoc
// @Summary		List the logged-in participant's own QRIS cross border submissions
// @Tags			qris-cross-border
// @Security		SessionCookie
// @Param			page	query		int	false	"Page number"
// @Param			limit	query		int	false	"Page size"
// @Success		200		{object}	utils.Response
// @Router			/qris-cross-border/my [get]
func (h *Handler) MyList(c *gin.Context) {
	actorID := utils.CurrentUserID(c)
	if actorID == nil {
		utils.Error(c, 401, "unauthorized")
		return
	}
	p := utils.GetPagination(c, "created_at")
	list, total, err := h.Service.MyList(*actorID, p)
	if err != nil {
		utils.Error(c, 500, "failed to fetch qris cross border records")
		return
	}
	utils.SuccessList(c, "ok", list, utils.BuildMeta(p.Page, p.Limit, total))
}

// DeleteSelf godoc
// @Summary		Delete the logged-in participant's own QRIS cross border submission
// @Tags			qris-cross-border
// @Security		SessionCookie
// @Param			uuid	path	string	true	"Record UUID"
// @Success		200		{object}	utils.Response
// @Failure		404		{object}	utils.Response
// @Router			/qris-cross-border/my/{uuid} [delete]
func (h *Handler) DeleteSelf(c *gin.Context) {
	actorID := utils.CurrentUserID(c)
	if actorID == nil {
		utils.Error(c, 401, "unauthorized")
		return
	}
	row, err := h.Service.DeleteSelf(c.Param("uuid"), *actorID)
	if err != nil {
		utils.Error(c, 404, "record not found")
		return
	}
	activity_logs.LogActivity(actorID, activity_logs.ActionDelete, "qris_cross_border", row.UUID.String(), row, nil, c.ClientIP(), c.Request.UserAgent())
	utils.Success(c, 200, "qris cross border deleted", nil)
}

// Update godoc
// @Summary		Update a QRIS cross border submission
// @Tags			qris-cross-border
// @Security		SessionCookie
// @Accept			multipart/form-data
// @Param			uuid				path		string	true	"Record UUID"
// @Param			peserta_uuid		formData	string	false	"Peserta UUID"
// @Param			image				formData	file	false	"Proof of payment image"
// @Param			nominal_asing		formData	number	false	"Amount in foreign currency"
// @Param			nominal_rupiah		formData	number	false	"Amount in IDR"
// @Param			merchant_name		formData	string	false	"Merchant name"
// @Param			reference_number	formData	string	false	"Reference number"
// @Param			status				formData	string	false	"Status"
// @Success		200					{object}	utils.Response
// @Failure		404					{object}	utils.Response
// @Router			/qris-cross-border/{uuid} [put]
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

	actorID := utils.CurrentUserID(c)
	before, after, err := h.Service.Update(c.Param("uuid"), UpdateInput{
		PesertaUUID:     c.PostForm("peserta_uuid"),
		NominalAsing:    parseFloatForm(c, "nominal_asing"),
		NominalRupiah:   parseFloatForm(c, "nominal_rupiah"),
		MerchantName:    c.PostForm("merchant_name"),
		ReferenceNumber: c.PostForm("reference_number"),
		Status:          c.PostForm("status"),
		ActorID:         actorID,
	}, file, fileHeader)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			utils.Error(c, 404, "record not found")
			return
		}
		if errors.Is(err, ErrPesertaNotFound) {
			utils.Error(c, 400, "peserta not found")
			return
		}
		if errors.Is(err, ErrDuplicateReference) {
			utils.Error(c, 400, "reference number already in use by another record")
			return
		}
		utils.Error(c, 400, "failed to update record")
		return
	}

	activity_logs.LogActivity(actorID, activity_logs.ActionUpdate, "qris_cross_border", after.UUID.String(), before, after, c.ClientIP(), c.Request.UserAgent())
	utils.Success(c, 200, "qris cross border updated", after)
}

type statusRequest struct {
	Status string `json:"status" binding:"required"`
}

// UpdateStatus godoc
// @Summary		Approve or reject a QRIS cross border submission
// @Tags			qris-cross-border
// @Security		SessionCookie
// @Param			uuid	path		string			true	"Record UUID"
// @Param			body	body		statusRequest	true	"New status"
// @Success		200		{object}	utils.Response
// @Router			/qris-cross-border/{uuid}/status [put]
func (h *Handler) UpdateStatus(c *gin.Context) {
	var req statusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "invalid request payload")
		return
	}

	actorID := utils.CurrentUserID(c)
	oldStatus, after, err := h.Service.UpdateStatus(c.Param("uuid"), req.Status, actorID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			utils.Error(c, 404, "record not found")
			return
		}
		if errors.Is(err, ErrInvalidStatus) {
			utils.Error(c, 400, "invalid status")
			return
		}
		utils.Error(c, 500, "failed to update status")
		return
	}

	activity_logs.LogActivity(actorID, activity_logs.ActionUpdate, "qris_cross_border", after.UUID.String(), gin.H{"status": oldStatus}, gin.H{"status": after.Status}, c.ClientIP(), c.Request.UserAgent())
	utils.Success(c, 200, "status updated", after)
}

// Delete godoc
// @Summary		Delete a QRIS cross border submission
// @Tags			qris-cross-border
// @Security		SessionCookie
// @Param			uuid	path	string	true	"Record UUID"
// @Success		200		{object}	utils.Response
// @Failure		404		{object}	utils.Response
// @Router			/qris-cross-border/{uuid} [delete]
func (h *Handler) Delete(c *gin.Context) {
	actorID := utils.CurrentUserID(c)
	row, err := h.Service.Delete(c.Param("uuid"), actorID)
	if err != nil {
		utils.Error(c, 404, "record not found")
		return
	}
	activity_logs.LogActivity(actorID, activity_logs.ActionDelete, "qris_cross_border", row.UUID.String(), row, nil, c.ClientIP(), c.Request.UserAgent())
	utils.Success(c, 200, "qris cross border deleted", nil)
}
