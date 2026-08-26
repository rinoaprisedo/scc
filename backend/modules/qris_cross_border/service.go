package qris_cross_border

import (
	"errors"
	"mime/multipart"
	"time"

	"baseadmin/backend/storage"
	"baseadmin/backend/utils"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/google/uuid"
)

var (
	ErrNotFound           = errors.New("qris cross border record not found")
	ErrPesertaNotFound    = errors.New("peserta not found")
	ErrInvalidStatus      = errors.New("invalid status")
	ErrDuplicateReference = errors.New("reference number already in use by another record")
)

// validStatuses gates the admin-facing UpdateStatus endpoint. "pending" is
// deliberately excluded — it's the system-managed state a new record starts
// in and the OCR job transitions out of; an admin never sets it directly.
var validStatuses = map[string]bool{
	StatusWaitingApproval: true,
	StatusApproved:        true,
	StatusRejected:        true,
}

// deepSeekAnthropicBaseURL is DeepSeek's Anthropic-wire-compatible endpoint —
// same request/response shape as the Claude API (image content blocks,
// forced tool_choice, ToolUseBlock results), so ocr.go needs no changes to
// switch providers; only which client the SDK talks to changes.
const deepSeekAnthropicBaseURL = "https://api.deepseek.com/anthropic"

// Service holds business rules for the qris_cross_border module. Repository
// stays a thin data-access layer, matching every other module.
type Service struct {
	repo    *Repository
	storage storage.StorageInterface
	client  anthropic.Client
	model   string
}

// NewService builds the OCR client for whichever provider is configured.
// provider is config.OCRProvider ("claude" or "deepseek"); apiKey/model are
// already resolved to the matching pair by the caller (see main.go).
func NewService(repo *Repository, s storage.StorageInterface, provider, apiKey, model string) *Service {
	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if provider == "deepseek" {
		opts = append(opts, option.WithBaseURL(deepSeekAnthropicBaseURL))
	}
	return &Service{
		repo:    repo,
		storage: s,
		client:  anthropic.NewClient(opts...),
		model:   model,
	}
}

// Response flattens the peserta relation into the fields the admin table/
// form actually needs — same precedent as qr_gate.ScanByUser — instead of
// serializing the full users.User struct (password, every profile column).
type Response struct {
	UUID            uuid.UUID `json:"uuid"`
	PesertaUUID     uuid.UUID `json:"peserta_uuid"`
	PesertaName     string    `json:"peserta_name"`
	PesertaKtp      string    `json:"peserta_ktp_number"`
	Image           string    `json:"image"`
	NominalAsing    float64   `json:"nominal_asing"`
	NominalRupiah   float64   `json:"nominal_rupiah"`
	MerchantName    string    `json:"merchant_name"`
	ReferenceNumber string    `json:"reference_number"`
	RejectReason    string    `json:"reject_reason"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func toResponse(row QrisCrossBorder) Response {
	resp := Response{
		UUID:          row.UUID,
		Image:         row.Image,
		NominalAsing:  row.NominalAsing,
		NominalRupiah: row.NominalRupiah,
		Status:        row.Status,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
	if row.MerchantName != nil {
		resp.MerchantName = *row.MerchantName
	}
	if row.ReferenceNumber != nil {
		resp.ReferenceNumber = *row.ReferenceNumber
	}
	if row.RejectReason != nil {
		resp.RejectReason = *row.RejectReason
	}
	if row.Peserta != nil {
		resp.PesertaUUID = row.Peserta.UUID
		resp.PesertaName = row.Peserta.Name
		if row.Peserta.KtpNumber != nil {
			resp.PesertaKtp = *row.Peserta.KtpNumber
		}
	}
	return resp
}

func (s *Service) List(p utils.Pagination, status, pesertaUUID string) ([]Response, int64, error) {
	rows, total, err := s.repo.List(p, status, pesertaUUID)
	if err != nil {
		return nil, 0, err
	}
	list := make([]Response, 0, len(rows))
	for _, row := range rows {
		list = append(list, toResponse(row))
	}
	return list, total, nil
}

func (s *Service) Get(uuidStr string) (*Response, error) {
	row, err := s.repo.FindByUUID(uuidStr)
	if err != nil {
		return nil, ErrNotFound
	}
	resp := toResponse(*row)
	return &resp, nil
}

type CreateInput struct {
	PesertaUUID string
	ActorID     *uint64
}

// Create only takes a peserta + proof image — nominal, merchant name,
// reference number, and status are filled in by the async OCR job kicked
// off below, not entered up front. Every new record starts at StatusPending.
func (s *Service) Create(in CreateInput, file multipart.File, header *multipart.FileHeader) (*Response, error) {
	pesertaID, err := s.repo.ResolvePesertaID(in.PesertaUUID)
	if err != nil {
		return nil, ErrPesertaNotFound
	}
	return s.createRow(pesertaID, in.ActorID, file, header)
}

// CreateSelf is the website's participant-facing counterpart to Create: the
// logged-in peserta submits proof of payment for themselves, so there's no
// peserta_uuid to resolve — the session's own user ID is both the record
// owner and the actor.
func (s *Service) CreateSelf(pesertaID uint64, file multipart.File, header *multipart.FileHeader) (*Response, error) {
	return s.createRow(pesertaID, &pesertaID, file, header)
}

func (s *Service) createRow(pesertaID uint64, actorID *uint64, file multipart.File, header *multipart.FileHeader) (*Response, error) {
	path, err := s.storage.Upload(file, header, "qris_cross_border")
	if err != nil {
		return nil, err
	}

	row := QrisCrossBorder{
		PesertaID: pesertaID,
		Image:     path,
		Status:    StatusPending,
	}
	row.CreatedBy = actorID

	if err := s.repo.Create(&row); err != nil {
		return nil, err
	}
	saved, err := s.repo.FindByUUID(row.UUID.String())
	if err != nil {
		return nil, err
	}

	go s.processOCR(saved.UUID.String())

	resp := toResponse(*saved)
	return &resp, nil
}

// MyList returns a page of the logged-in peserta's own submission history
// for the website — image, nominal, and status per record, newest first.
func (s *Service) MyList(pesertaID uint64, p utils.Pagination) ([]Response, int64, error) {
	rows, total, err := s.repo.ListByPeserta(pesertaID, p)
	if err != nil {
		return nil, 0, err
	}
	list := make([]Response, 0, len(rows))
	for _, row := range rows {
		list = append(list, toResponse(row))
	}
	return list, total, nil
}

// DeleteSelf lets a peserta delete their own submission — scoped by
// PesertaID so one participant can't delete another's record by guessing a
// UUID, unlike the admin Delete which trusts the permission guard instead.
func (s *Service) DeleteSelf(uuidStr string, pesertaID uint64) (*Response, error) {
	row, err := s.repo.FindByUUID(uuidStr)
	if err != nil || row.PesertaID != pesertaID {
		return nil, ErrNotFound
	}
	if err := s.repo.Delete(row, &pesertaID); err != nil {
		return nil, err
	}
	resp := toResponse(*row)
	return &resp, nil
}

type UpdateInput struct {
	PesertaUUID     string
	NominalAsing    float64
	NominalRupiah   float64
	MerchantName    string
	ReferenceNumber string
	Status          string
	ActorID         *uint64
}

// Update returns the pre-update snapshot alongside the saved record so the
// handler can log a before/after activity diff, matching users.Service.Update.
// This is also where an admin corrects whatever the AI extraction got wrong.
func (s *Service) Update(uuidStr string, in UpdateInput, file multipart.File, header *multipart.FileHeader) (before, after *Response, err error) {
	row, err := s.repo.FindByUUID(uuidStr)
	if err != nil {
		return nil, nil, ErrNotFound
	}
	old := toResponse(*row)

	if in.PesertaUUID != "" {
		pesertaID, err := s.repo.ResolvePesertaID(in.PesertaUUID)
		if err != nil {
			return nil, nil, ErrPesertaNotFound
		}
		row.PesertaID = pesertaID
		// Stale preloaded association must be cleared too, not just the FK
		// column — see peserta.Service.Update for why a full-struct Save
		// would otherwise silently re-sync PesertaID from the old Peserta.
		row.Peserta = nil
	}

	if file != nil {
		path, err := s.storage.Upload(file, header, "qris_cross_border")
		if err != nil {
			return nil, nil, err
		}
		row.Image = path
	}

	if in.ReferenceNumber != "" {
		exists, err := s.repo.ReferenceNumberExists(in.ReferenceNumber, uuidStr)
		if err != nil {
			return nil, nil, err
		}
		if exists {
			return nil, nil, ErrDuplicateReference
		}
	}

	row.NominalAsing = in.NominalAsing
	row.NominalRupiah = in.NominalRupiah
	row.MerchantName = nilIfEmpty(in.MerchantName)
	row.ReferenceNumber = nilIfEmpty(in.ReferenceNumber)
	if in.Status != "" && validStatuses[in.Status] {
		row.Status = in.Status
	}
	row.UpdatedBy = in.ActorID

	if err := s.repo.Save(row); err != nil {
		return nil, nil, err
	}
	saved, err := s.repo.FindByUUID(row.UUID.String())
	if err != nil {
		return nil, nil, err
	}
	afterResp := toResponse(*saved)
	return &old, &afterResp, nil
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (s *Service) UpdateStatus(uuidStr, status string, actorID *uint64) (oldStatus string, after *Response, err error) {
	if !validStatuses[status] {
		return "", nil, ErrInvalidStatus
	}
	row, err := s.repo.FindByUUID(uuidStr)
	if err != nil {
		return "", nil, ErrNotFound
	}
	oldStatus = row.Status
	row.Status = status
	row.UpdatedBy = actorID
	if err := s.repo.Save(row); err != nil {
		return "", nil, err
	}
	resp := toResponse(*row)
	return oldStatus, &resp, nil
}

func (s *Service) Delete(uuidStr string, actorID *uint64) (*Response, error) {
	row, err := s.repo.FindByUUID(uuidStr)
	if err != nil {
		return nil, ErrNotFound
	}
	if err := s.repo.Delete(row, actorID); err != nil {
		return nil, err
	}
	resp := toResponse(*row)
	return &resp, nil
}
