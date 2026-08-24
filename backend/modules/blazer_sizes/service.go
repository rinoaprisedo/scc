package blazer_sizes

import (
	"errors"

	"baseadmin/backend/utils"
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

func (s *Service) Options() ([]BlazerSize, error) {
	return s.repo.All()
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
