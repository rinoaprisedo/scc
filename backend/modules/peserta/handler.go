package peserta

import (
	"errors"
	"time"

	"baseadmin/backend/modules/activity_logs"
	"baseadmin/backend/modules/qr_gate"
	"baseadmin/backend/modules/settings"
	"baseadmin/backend/storage"
	"baseadmin/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Handler is the HTTP layer for the peserta module.
type Handler struct {
	Service *Service
	// QrGateService backs the admin "point history" action — reuses the
	// same scan-listing logic the participant's own History Scanner uses,
	// just looked up by an admin-chosen peserta UUID instead of "me".
	QrGateService *qr_gate.Service
	// SettingsService backs the registration_deadline/form_edit_deadline
	// checks below — read here at the HTTP layer (like QrGateService above)
	// rather than threading a settings dependency into peserta's own
	// business logic in Service.
	SettingsService *settings.Service
}

func NewHandler(db *gorm.DB, s storage.StorageInterface, rdb *redis.Client) *Handler {
	return &Handler{
		Service:         NewService(NewRepository(db), s, rdb),
		QrGateService:   qr_gate.NewService(qr_gate.NewRepository(db)),
		SettingsService: settings.NewService(settings.NewRepository(db), rdb, s),
	}
}

// deadlineLayout matches an HTML <input type="datetime-local"> value
// (admin-entered in the Settings page), interpreted in the server's local
// timezone since this project has no per-user timezone handling elsewhere.
const deadlineLayout = "2006-01-02T15:04"

// isPastDeadline reports whether the given deadline setting value is in the
// past. An empty/unset or malformed value means "no deadline" — it never
// blocks, rather than blocking on bad admin input.
func isPastDeadline(value string) bool {
	if value == "" {
		return false
	}
	t, err := time.ParseInLocation(deadlineLayout, value, time.Local)
	if err != nil {
		return false
	}
	return time.Now().After(t)
}

// List godoc
// @Summary		List peserta
// @Tags			peserta
// @Security		SessionCookie
// @Param			page	query		int		false	"Page number"
// @Param			limit	query		int		false	"Page size"
// @Param			search	query		string	false	"Search by name/email/phone/KTP/passport"
// @Success		200		{object}	utils.Response
// @Router			/peserta [get]
func (h *Handler) List(c *gin.Context) {
	p := utils.GetPagination(c, "created_at")
	list, total, err := h.Service.List(p)
	if err != nil {
		utils.Error(c, 500, "failed to fetch peserta")
		return
	}
	utils.SuccessList(c, "ok", list, utils.BuildMeta(p.Page, p.Limit, total))
}

// Get godoc
// @Summary		Get a peserta by UUID
// @Tags			peserta
// @Security		SessionCookie
// @Param			uuid	path		string	true	"Peserta UUID"
// @Success		200		{object}	utils.Response
// @Failure		404		{object}	utils.Response
// @Router			/peserta/{uuid} [get]
func (h *Handler) Get(c *gin.Context) {
	user, err := h.Service.Get(c.Param("uuid"))
	if err != nil {
		utils.Error(c, 404, "peserta not found")
		return
	}
	utils.Success(c, 200, "ok", user)
}

// PointHistory godoc
// @Summary		Get a peserta's QR gate scan/point history
// @Tags			peserta
// @Security		SessionCookie
// @Param			uuid	path		string	true	"Peserta UUID"
// @Success		200		{object}	utils.Response
// @Failure		404		{object}	utils.Response
// @Router			/peserta/{uuid}/points [get]
func (h *Handler) PointHistory(c *gin.Context) {
	user, err := h.Service.Get(c.Param("uuid"))
	if err != nil {
		utils.Error(c, 404, "peserta not found")
		return
	}
	scans, totalPoints, err := h.QrGateService.MyScans(user.ID)
	if err != nil {
		utils.Error(c, 500, "failed to fetch point history")
		return
	}
	utils.Success(c, 200, "ok", gin.H{"scans": scans, "total_points": totalPoints})
}

type profileRequest struct {
	Title                   string `json:"title"`
	FirstName               string `json:"first_name"`
	MiddleName              string `json:"middle_name"`
	LastName                string `json:"last_name"`
	BirthDate               string `json:"birth_date"`
	OriginCityUUID          string `json:"origin_city_uuid"`
	OriginCityOther         string `json:"origin_city_other"`
	NearestAirportUUID      string `json:"nearest_airport_uuid"`
	DietaryRestriction      string `json:"dietary_restriction"`
	DietaryRestrictionOther string `json:"dietary_restriction_other"`
	PhoneNumber             string `json:"phone_number"`
	KtpNumber               string `json:"ktp_number" binding:"required"`
	NomorKtp                string `json:"nomor_ktp"`
	PassportNumber          string `json:"passport_number"`
	PassportExpiry          string `json:"passport_expiry"`
	JacketSize              string `json:"jacket_size"`
	PoloSize                string `json:"polo_size"`
	// NomorMeja is admin-only (see User.NomorMeja) — deliberately absent
	// from selfProfileRequest below.
	NomorMeja string `json:"nomor_meja"`
}

func (r profileRequest) toInput() ProfileInput {
	return ProfileInput{
		Title:                   r.Title,
		FirstName:               r.FirstName,
		MiddleName:              r.MiddleName,
		LastName:                r.LastName,
		BirthDate:               r.BirthDate,
		OriginCityUUID:          r.OriginCityUUID,
		OriginCityOther:         r.OriginCityOther,
		NearestAirportUUID:      r.NearestAirportUUID,
		DietaryRestriction:      r.DietaryRestriction,
		DietaryRestrictionOther: r.DietaryRestrictionOther,
		PhoneNumber:             r.PhoneNumber,
		KtpNumber:               r.KtpNumber,
		NomorKtp:                r.NomorKtp,
		PassportNumber:          r.PassportNumber,
		PassportExpiry:          r.PassportExpiry,
		JacketSize:              r.JacketSize,
		PoloSize:                r.PoloSize,
		NomorMeja:               r.NomorMeja,
	}
}

type createRequest struct {
	Name string `json:"name" binding:"required"`
	// Email is optional for peserta — NIK (ktp_number, in profileRequest)
	// is the required identifier instead, since peserta accounts will
	// eventually log in with NIK + password rather than email.
	Email    string `json:"email" binding:"omitempty,email"`
	Password string `json:"password" binding:"required"`
	Status   string `json:"status"`
	profileRequest
}

// Create godoc
// @Summary		Create a peserta
// @Tags			peserta
// @Security		SessionCookie
// @Param			body	body		createRequest	true	"Peserta payload"
// @Success		201		{object}	utils.Response
// @Failure		400		{object}	utils.Response
// @Router			/peserta [post]
func (h *Handler) Create(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "invalid request payload")
		return
	}

	actorID := utils.CurrentUserID(c)
	user, err := h.Service.Create(CreateInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Status:   req.Status,
		Profile:  req.profileRequest.toInput(),
		ActorID:  actorID,
	})
	if err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	activity_logs.LogActivity(actorID, activity_logs.ActionCreate, "peserta", user.UUID.String(), nil, user, c.ClientIP(), c.Request.UserAgent())
	utils.Success(c, 201, "peserta created", user)
}

type updateRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Status   string `json:"status"`
	profileRequest
}

// Update godoc
// @Summary		Update a peserta
// @Tags			peserta
// @Security		SessionCookie
// @Param			uuid	path		string			true	"Peserta UUID"
// @Param			body	body		updateRequest	true	"Fields to update"
// @Success		200		{object}	utils.Response
// @Failure		404		{object}	utils.Response
// @Router			/peserta/{uuid} [put]
func (h *Handler) Update(c *gin.Context) {
	var req updateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "invalid request payload")
		return
	}

	actorID := utils.CurrentUserID(c)
	before, after, err := h.Service.Update(c.Param("uuid"), UpdateInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Status:   req.Status,
		Profile:  req.profileRequest.toInput(),
		ActorID:  actorID,
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			utils.Error(c, 404, "peserta not found")
			return
		}
		utils.Error(c, 400, "failed to update peserta")
		return
	}

	activity_logs.LogActivity(actorID, activity_logs.ActionUpdate, "peserta", after.UUID.String(), before, after, c.ClientIP(), c.Request.UserAgent())
	utils.Success(c, 200, "peserta updated", after)
}

// Delete godoc
// @Summary		Delete a peserta
// @Tags			peserta
// @Security		SessionCookie
// @Param			uuid	path	string	true	"Peserta UUID"
// @Success		200		{object}	utils.Response
// @Failure		404		{object}	utils.Response
// @Router			/peserta/{uuid} [delete]
func (h *Handler) Delete(c *gin.Context) {
	actorID := utils.CurrentUserID(c)
	user, err := h.Service.Delete(c.Param("uuid"), actorID)
	if err != nil {
		utils.Error(c, 404, "peserta not found")
		return
	}

	activity_logs.LogActivity(actorID, activity_logs.ActionDelete, "peserta", user.UUID.String(), user, nil, c.ClientIP(), c.Request.UserAgent())
	utils.Success(c, 200, "peserta deleted", nil)
}

type statusRequest struct {
	Status string `json:"status" binding:"required"`
}

// UpdateStatus godoc
// @Summary		Change a peserta's status
// @Tags			peserta
// @Security		SessionCookie
// @Param			uuid	path		string			true	"Peserta UUID"
// @Param			body	body		statusRequest	true	"New status"
// @Success		200		{object}	utils.Response
// @Router			/peserta/{uuid}/status [put]
func (h *Handler) UpdateStatus(c *gin.Context) {
	var req statusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "invalid request payload")
		return
	}

	oldStatus, user, err := h.Service.UpdateStatus(c.Param("uuid"), req.Status)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			utils.Error(c, 404, "peserta not found")
			return
		}
		utils.Error(c, 500, "failed to update status")
		return
	}

	actorID := utils.CurrentUserID(c)
	activity_logs.LogActivity(actorID, activity_logs.ActionUpdate, "peserta", user.UUID.String(), gin.H{"status": oldStatus}, gin.H{"status": user.Status}, c.ClientIP(), c.Request.UserAgent())
	utils.Success(c, 200, "status updated", user)
}

// UpdateMe godoc
// @Summary		Update the current peserta's own profile
// @Tags			peserta
// @Security		SessionCookie
// @Param			body	body		profileRequest	true	"Profile fields"
// @Success		200		{object}	utils.Response
// @Failure		401		{object}	utils.Response
// @Router			/peserta/me [put]
// selfProfileRequest mirrors profileRequest but with every field a
// *string — the website saves the Form and Shirt Size tabs as two separate
// partial requests, so binding must be able to tell "field omitted" (nil,
// leave unchanged) apart from "field explicitly cleared" (non-nil "").
// profileRequest's plain strings (and its binding:"required" on KtpNumber,
// needed at admin-create time) can't express that, hence a separate type.
type selfProfileRequest struct {
	Title                   *string `json:"title"`
	FirstName               *string `json:"first_name"`
	MiddleName              *string `json:"middle_name"`
	LastName                *string `json:"last_name"`
	BirthDate               *string `json:"birth_date"`
	OriginCityUUID          *string `json:"origin_city_uuid"`
	OriginCityOther         *string `json:"origin_city_other"`
	NearestAirportUUID      *string `json:"nearest_airport_uuid"`
	DietaryRestriction      *string `json:"dietary_restriction"`
	DietaryRestrictionOther *string `json:"dietary_restriction_other"`
	PhoneNumber             *string `json:"phone_number"`
	// No ktp_number here — see SelfProfileInput's comment: the login NIK
	// isn't participant-editable. nomor_ktp is the separate field the
	// website's "Nomor KTP" form input actually writes to.
	NomorKtp       *string `json:"nomor_ktp"`
	PassportNumber *string `json:"passport_number"`
	PassportExpiry *string `json:"passport_expiry"`
	JacketSize     *string `json:"jacket_size"`
	PoloSize       *string `json:"polo_size"`
}

func (r selfProfileRequest) toInput() SelfProfileInput {
	return SelfProfileInput{
		Title:                   r.Title,
		FirstName:               r.FirstName,
		MiddleName:              r.MiddleName,
		LastName:                r.LastName,
		BirthDate:               r.BirthDate,
		OriginCityUUID:          r.OriginCityUUID,
		OriginCityOther:         r.OriginCityOther,
		NearestAirportUUID:      r.NearestAirportUUID,
		DietaryRestriction:      r.DietaryRestriction,
		DietaryRestrictionOther: r.DietaryRestrictionOther,
		PhoneNumber:             r.PhoneNumber,
		NomorKtp:                r.NomorKtp,
		PassportNumber:          r.PassportNumber,
		PassportExpiry:          r.PassportExpiry,
		JacketSize:              r.JacketSize,
		PoloSize:                r.PoloSize,
	}
}

func (h *Handler) UpdateMe(c *gin.Context) {
	userID := utils.CurrentUserID(c)
	if userID == nil {
		utils.Error(c, 401, "authentication required")
		return
	}
	var req selfProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "invalid request payload")
		return
	}

	settingsMap, _ := h.SettingsService.List()
	if isPastDeadline(settingsMap["form_edit_deadline"]) {
		utils.Error(c, 403, "Maaf, batas waktu pengisian/edit formulir sudah lewat")
		return
	}

	user, err := h.Service.UpdateProfile(*userID, req.toInput())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			utils.Error(c, 404, "peserta not found")
			return
		}
		utils.Error(c, 400, "failed to update profile")
		return
	}
	utils.Success(c, 200, "profile updated", user)
}

type attendanceRequest struct {
	Confirmed bool `json:"confirmed"`
}

// SetMyAttendance godoc
// @Summary		Answer the post-login attendance prompt
// @Tags			peserta
// @Security		SessionCookie
// @Param			body	body		attendanceRequest	true	"Attendance answer"
// @Success		200		{object}	utils.Response
// @Failure		401		{object}	utils.Response
// @Router			/peserta/me/attendance [put]
func (h *Handler) SetMyAttendance(c *gin.Context) {
	userID := utils.CurrentUserID(c)
	if userID == nil {
		utils.Error(c, 401, "authentication required")
		return
	}
	var req attendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "invalid request payload")
		return
	}

	settingsMap, _ := h.SettingsService.List()
	if isPastDeadline(settingsMap["registration_deadline"]) {
		utils.Error(c, 403, "Maaf, registrasi sudah ditutup")
		return
	}

	user, err := h.Service.SetAttendance(*userID, req.Confirmed)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			utils.Error(c, 404, "peserta not found")
			return
		}
		utils.Error(c, 400, "failed to save attendance")
		return
	}
	utils.Success(c, 200, "attendance saved", user)
}

// UploadMyKtp godoc
// @Summary		Upload the current peserta's own KTP scan
// @Tags			peserta
// @Security		SessionCookie
// @Accept			multipart/form-data
// @Param			ktp_file	formData	file	true	"KTP scan image"
// @Success		200			{object}	utils.Response
// @Failure		401			{object}	utils.Response
// @Router			/peserta/me/ktp [post]
func (h *Handler) UploadMyKtp(c *gin.Context) {
	userID := utils.CurrentUserID(c)
	if userID == nil {
		utils.Error(c, 401, "authentication required")
		return
	}
	fileHeader, err := c.FormFile("ktp_file")
	if err != nil {
		utils.Error(c, 400, "ktp file is required")
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		utils.Error(c, 400, "failed to read uploaded file")
		return
	}
	defer file.Close()

	url, err := h.Service.UploadKtpSelf(*userID, file, fileHeader)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			utils.Error(c, 404, "peserta not found")
			return
		}
		utils.Error(c, 400, err.Error())
		return
	}
	utils.Success(c, 200, "ktp uploaded", gin.H{"ktp_file": url})
}

// UploadKtp godoc
// @Summary		Upload a peserta's KTP scan
// @Tags			peserta
// @Security		SessionCookie
// @Param			uuid		path	string	true	"Peserta UUID"
// @Accept			multipart/form-data
// @Param			ktp_file	formData	file	true	"KTP scan image"
// @Success		200			{object}	utils.Response
// @Router			/peserta/{uuid}/ktp [post]
func (h *Handler) UploadKtp(c *gin.Context) {
	fileHeader, err := c.FormFile("ktp_file")
	if err != nil {
		utils.Error(c, 400, "ktp file is required")
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		utils.Error(c, 400, "failed to read uploaded file")
		return
	}
	defer file.Close()

	url, err := h.Service.UploadKtp(c.Param("uuid"), file, fileHeader)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			utils.Error(c, 404, "peserta not found")
			return
		}
		utils.Error(c, 400, err.Error())
		return
	}
	utils.Success(c, 200, "ktp uploaded", gin.H{"ktp_file": url})
}
