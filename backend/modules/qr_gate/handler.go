package qr_gate

import (
	"baseadmin/backend/modules/activity_logs"
	"baseadmin/backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler is the HTTP layer for the qr_gate module.
type Handler struct {
	Service *Service
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{Service: NewService(NewRepository(db))}
}

// List godoc
// @Summary		List QR gates
// @Tags			qr-gate
// @Security		SessionCookie
// @Param			page	query		int		false	"Page number"
// @Param			limit	query		int		false	"Page size"
// @Param			search	query		string	false	"Search by name/code"
// @Success		200		{object}	utils.Response
// @Router			/qr-gate [get]
func (h *Handler) List(c *gin.Context) {
	p := utils.GetPagination(c, "name")
	list, total, err := h.Service.List(p)
	if err != nil {
		utils.Error(c, 500, "failed to fetch qr gates")
		return
	}
	utils.SuccessList(c, "ok", list, utils.BuildMeta(p.Page, p.Limit, total))
}

// Get godoc
// @Summary		Get a QR gate
// @Tags			qr-gate
// @Security		SessionCookie
// @Param			uuid	path		string	true	"QR gate UUID"
// @Success		200		{object}	utils.Response
// @Failure		404		{object}	utils.Response
// @Router			/qr-gate/{uuid} [get]
func (h *Handler) Get(c *gin.Context) {
	gate, err := h.Service.Get(c.Param("uuid"))
	if err != nil {
		utils.Error(c, 404, "qr gate not found")
		return
	}
	utils.Success(c, 200, "ok", gate)
}

// GateScans godoc
// @Summary		List participants who scanned a QR gate
// @Tags			qr-gate
// @Security		SessionCookie
// @Param			uuid	path		string	true	"QR gate UUID"
// @Success		200		{object}	utils.Response
// @Failure		404		{object}	utils.Response
// @Router			/qr-gate/{uuid}/scans [get]
func (h *Handler) GateScans(c *gin.Context) {
	scans, err := h.Service.GateScans(c.Param("uuid"))
	if err != nil {
		utils.Error(c, 404, "qr gate not found")
		return
	}
	utils.Success(c, 200, "ok", scans)
}

type qrGateRequest struct {
	Name       string `json:"name" binding:"required"`
	Code       string `json:"code" binding:"required"`
	Points     int    `json:"points"`
	IsReusable bool   `json:"is_reusable"`
}

// Create godoc
// @Summary		Create a QR gate
// @Tags			qr-gate
// @Security		SessionCookie
// @Param			body	body		qrGateRequest	true	"QR gate payload"
// @Success		201		{object}	utils.Response
// @Router			/qr-gate [post]
func (h *Handler) Create(c *gin.Context) {
	var req qrGateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "invalid request payload")
		return
	}
	actorID := utils.CurrentUserID(c)
	gate, err := h.Service.Create(Input{Name: req.Name, Code: req.Code, Points: req.Points, IsReusable: req.IsReusable, ActorID: actorID})
	if err != nil {
		utils.Error(c, 400, "failed to create qr gate — code may already be in use")
		return
	}
	activity_logs.LogActivity(actorID, activity_logs.ActionCreate, "qr_gate", gate.UUID.String(), nil, gate, c.ClientIP(), c.Request.UserAgent())
	utils.Success(c, 201, "qr gate created", gate)
}

// Update godoc
// @Summary		Update a QR gate
// @Tags			qr-gate
// @Security		SessionCookie
// @Param			uuid	path		string			true	"QR gate UUID"
// @Param			body	body		qrGateRequest	true	"QR gate payload"
// @Success		200		{object}	utils.Response
// @Failure		404		{object}	utils.Response
// @Router			/qr-gate/{uuid} [put]
func (h *Handler) Update(c *gin.Context) {
	var req qrGateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "invalid request payload")
		return
	}
	actorID := utils.CurrentUserID(c)
	before, after, err := h.Service.Update(c.Param("uuid"), Input{Name: req.Name, Code: req.Code, Points: req.Points, IsReusable: req.IsReusable, ActorID: actorID})
	if err != nil {
		utils.Error(c, 404, "qr gate not found")
		return
	}
	activity_logs.LogActivity(actorID, activity_logs.ActionUpdate, "qr_gate", after.UUID.String(), before, after, c.ClientIP(), c.Request.UserAgent())
	utils.Success(c, 200, "qr gate updated", after)
}

// Delete godoc
// @Summary		Delete a QR gate
// @Tags			qr-gate
// @Security		SessionCookie
// @Param			uuid	path	string	true	"QR gate UUID"
// @Success		200		{object}	utils.Response
// @Failure		404		{object}	utils.Response
// @Router			/qr-gate/{uuid} [delete]
func (h *Handler) Delete(c *gin.Context) {
	actorID := utils.CurrentUserID(c)
	gate, err := h.Service.Delete(c.Param("uuid"), actorID)
	if err != nil {
		utils.Error(c, 404, "qr gate not found")
		return
	}
	activity_logs.LogActivity(actorID, activity_logs.ActionDelete, "qr_gate", gate.UUID.String(), gate, nil, c.ClientIP(), c.Request.UserAgent())
	utils.Success(c, 200, "qr gate deleted", nil)
}

type scanRequest struct {
	Code string `json:"code" binding:"required"`
}

// Scan godoc
// @Summary		Redeem a QR gate's code as the logged-in participant
// @Tags			qr-gate
// @Security		SessionCookie
// @Param			body	body		scanRequest	true	"Scanned code"
// @Success		200		{object}	utils.Response
// @Failure		404		{object}	utils.Response
// @Failure		409		{object}	utils.Response
// @Router			/qr-gate/scan [post]
func (h *Handler) Scan(c *gin.Context) {
	userID := utils.CurrentUserID(c)
	if userID == nil {
		utils.Error(c, 401, "unauthorized")
		return
	}

	var req scanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "invalid request payload")
		return
	}

	result, err := h.Service.Scan(*userID, req.Code)
	if err != nil {
		switch err {
		case ErrCodeNotFound:
			utils.Error(c, 404, "QR code tidak ditemukan")
		case ErrAlreadyClaimed:
			utils.Error(c, 409, "QR ini sudah digunakan peserta lain")
		default:
			utils.Error(c, 500, "failed to record scan")
		}
		return
	}

	if result.AlreadyScanned {
		// Not a failure — the code is valid, they just already have credit
		// for it. Frontend renders this as a success state with a
		// "you already scanned this" message rather than an error.
		utils.Success(c, 200, "already scanned", gin.H{
			"already_scanned": true,
			"gate_name":       result.Gate.Name,
			"points_awarded":  0,
		})
		return
	}

	activity_logs.LogActivity(userID, activity_logs.ActionCreate, "qr_gate_scan", result.ScanUUID.String(), nil, result, c.ClientIP(), c.Request.UserAgent())
	utils.Success(c, 200, "scan recorded", gin.H{
		"already_scanned": false,
		"gate_name":       result.Gate.Name,
		"points_awarded":  result.PointsAwarded,
		"scanned_at":      result.ScannedAt,
	})
}

// MyScans godoc
// @Summary		List the logged-in participant's own successful scans
// @Tags			qr-gate
// @Security		SessionCookie
// @Success		200	{object}	utils.Response
// @Router			/qr-gate/my-scans [get]
func (h *Handler) MyScans(c *gin.Context) {
	userID := utils.CurrentUserID(c)
	if userID == nil {
		utils.Error(c, 401, "unauthorized")
		return
	}

	scans, totalPoints, err := h.Service.MyScans(*userID)
	if err != nil {
		utils.Error(c, 500, "failed to fetch scan history")
		return
	}
	utils.Success(c, 200, "ok", gin.H{"scans": scans, "total_points": totalPoints})
}
