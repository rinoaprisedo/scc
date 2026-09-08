package blazer_sizes

import (
	"errors"

	"baseadmin/backend/utils"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("blazer size not found")

// Service holds business rules for the blazer_sizes module.
type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(p utils.Pagination) ([]BlazerSize, int64, error) {
	return s.repo.List(p)
}

// SizeOption is what /blazer-sizes/options returns to dropdowns (website
// onboarding + admin peserta form) — Remaining is the computed Stock minus
// Assigned, not the raw Stock quota, so a size disappears from availability
// the moment enough peserta have picked it, not only when an admin manually
// zeroes its Stock.
type SizeOption struct {
	UUID      uuid.UUID `json:"uuid"`
	Size      string    `json:"size"`
	Stock     int       `json:"stock"`
	Remaining int64     `json:"remaining"`
}

func (s *Service) Options() ([]SizeOption, error) {
	rows, err := s.repo.OptionsWithRemaining()
	if err != nil {
		return nil, err
	}
	options := make([]SizeOption, 0, len(rows))
	for _, row := range rows {
		options = append(options, SizeOption{
			UUID:      row.UUID,
			Size:      row.Size,
			Stock:     row.Stock,
			Remaining: int64(row.Stock) - row.Assigned,
		})
	}
	return options, nil
}

func (s *Service) Get(uuidStr string) (*BlazerSize, error) {
	blazerSize, err := s.repo.FindByUUID(uuidStr)
	if err != nil {
		return nil, ErrNotFound
	}
	return blazerSize, nil
}

type Input struct {
	Size    string
	Stock   int
	ActorID *uint64
}

func (s *Service) Create(in Input) (*BlazerSize, error) {
	order, err := s.repo.NextOrder()
	if err != nil {
		return nil, err
	}
	blazerSize := BlazerSize{Size: in.Size, Stock: in.Stock, Order: order}
	blazerSize.CreatedBy = in.ActorID
	if err := s.repo.Create(&blazerSize); err != nil {
		return nil, err
	}
	return &blazerSize, nil
}

func (s *Service) Update(uuidStr string, in Input) (before *BlazerSize, after *BlazerSize, err error) {
	blazerSize, err := s.repo.FindByUUID(uuidStr)
	if err != nil {
		return nil, nil, ErrNotFound
	}
	old := *blazerSize

	blazerSize.Size = in.Size
	blazerSize.Stock = in.Stock
	blazerSize.UpdatedBy = in.ActorID

	if err := s.repo.Save(blazerSize); err != nil {
		return nil, nil, err
	}
	return &old, blazerSize, nil
}

func (s *Service) Delete(uuidStr string, actorID *uint64) (*BlazerSize, error) {
	blazerSize, err := s.repo.FindByUUID(uuidStr)
	if err != nil {
		return nil, ErrNotFound
	}
	if err := s.repo.Delete(blazerSize, actorID); err != nil {
		return nil, err
	}
	return blazerSize, nil
}

type ReorderItem struct {
	UUID  string
	Order int
}

// Reorder applies a full drag-reordered list in one shot — mirrors
// menu_sections.Service.Reorder (see there for why per-row errors are
// swallowed rather than aborting the whole batch).
func (s *Service) Reorder(items []ReorderItem) {
	for _, item := range items {
		_ = s.repo.UpdateOrder(item.UUID, item.Order)
	}
}
