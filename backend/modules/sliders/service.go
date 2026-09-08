package sliders

import (
	"errors"
	"mime/multipart"

	"baseadmin/backend/storage"
	"baseadmin/backend/utils"

	"github.com/google/uuid"
)

var (
	ErrNotFound    = errors.New("slider not found")
	ErrInvalidType = errors.New("invalid slider type")
)

// Service holds business rules for the sliders module.
type Service struct {
	repo    *Repository
	storage storage.StorageInterface
}

func NewService(repo *Repository, s storage.StorageInterface) *Service {
	return &Service{repo: repo, storage: s}
}

func (s *Service) List(p utils.Pagination, sliderType string) ([]Slider, int64, error) {
	return s.repo.List(p, sliderType)
}

// ActiveSlider is what /sliders/active returns to the website carousel —
// just the fields it actually renders, not the full admin row (order,
// audit columns).
type ActiveSlider struct {
	UUID  uuid.UUID `json:"uuid"`
	Image string    `json:"image"`
	Type  string    `json:"type"`
}

func (s *Service) Active() ([]ActiveSlider, error) {
	rows, err := s.repo.ListActive()
	if err != nil {
		return nil, err
	}
	list := make([]ActiveSlider, 0, len(rows))
	for _, row := range rows {
		list = append(list, ActiveSlider{UUID: row.UUID, Image: row.Image, Type: row.Type})
	}
	return list, nil
}

func (s *Service) Get(uuidStr string) (*Slider, error) {
	slider, err := s.repo.FindByUUID(uuidStr)
	if err != nil {
		return nil, ErrNotFound
	}
	return slider, nil
}

func (s *Service) Create(actorID *uint64, sliderType string, file multipart.File, header *multipart.FileHeader) (*Slider, error) {
	if !ValidTypes[sliderType] {
		return nil, ErrInvalidType
	}
	path, err := s.storage.Upload(file, header, "sliders")
	if err != nil {
		return nil, err
	}
	order, err := s.repo.NextOrder(sliderType)
	if err != nil {
		return nil, err
	}
	slider := Slider{Image: path, Type: sliderType, Order: order, IsActive: true}
	slider.CreatedBy = actorID
	if err := s.repo.Create(&slider); err != nil {
		return nil, err
	}
	return &slider, nil
}

type UpdateInput struct {
	Type     string
	IsActive bool
	ActorID  *uint64
}

// Update only re-uploads the image when a new file is sent — file is nil on
// a plain metadata (is_active/type) edit, same precedent as
// qris_cross_border.Service.Update. Changing Type does NOT reset Order —
// the row keeps whatever order value it had, which only matters again once
// it's re-sorted within its new type's NextOrder/reorder scope.
func (s *Service) Update(uuidStr string, in UpdateInput, file multipart.File, header *multipart.FileHeader) (before, after *Slider, err error) {
	if !ValidTypes[in.Type] {
		return nil, nil, ErrInvalidType
	}
	slider, err := s.repo.FindByUUID(uuidStr)
	if err != nil {
		return nil, nil, ErrNotFound
	}
	old := *slider

	if file != nil {
		path, err := s.storage.Upload(file, header, "sliders")
		if err != nil {
			return nil, nil, err
		}
		slider.Image = path
	}
	slider.Type = in.Type
	slider.IsActive = in.IsActive
	slider.UpdatedBy = in.ActorID

	if err := s.repo.Save(slider); err != nil {
		return nil, nil, err
	}
	return &old, slider, nil
}

func (s *Service) Delete(uuidStr string, actorID *uint64) (*Slider, error) {
	slider, err := s.repo.FindByUUID(uuidStr)
	if err != nil {
		return nil, ErrNotFound
	}
	if err := s.repo.Delete(slider, actorID); err != nil {
		return nil, err
	}
	return slider, nil
}

type ReorderItem struct {
	UUID  string
	Order int
}

// Reorder applies a full drag-reordered list in one shot — mirrors
// blazer_sizes.Service.Reorder (see there for why per-row errors are
// swallowed rather than aborting the whole batch).
func (s *Service) Reorder(items []ReorderItem) {
	for _, item := range items {
		_ = s.repo.UpdateOrder(item.UUID, item.Order)
	}
}
