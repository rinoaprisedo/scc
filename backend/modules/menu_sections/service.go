package menu_sections

import "errors"

var ErrNotFound = errors.New("menu section not found")

// Service holds business rules for the menu_sections module.
type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List() ([]MenuSection, error) {
	return s.repo.List()
}

func (s *Service) Get(uuidStr string) (*MenuSection, error) {
	section, err := s.repo.FindByUUID(uuidStr)
	if err != nil {
		return nil, ErrNotFound
	}
	return section, nil
}

type SectionInput struct {
	Name    string
	Icon    string
	Order   int
	ActorID *uint64
}

func (s *Service) Create(in SectionInput) (*MenuSection, error) {
	section := MenuSection{Name: in.Name, Icon: in.Icon, Order: in.Order}
	section.CreatedBy = in.ActorID
	if err := s.repo.Create(&section); err != nil {
		return nil, err
	}
	return &section, nil
}

func (s *Service) Update(uuidStr string, in SectionInput) (before *MenuSection, after *MenuSection, err error) {
	section, err := s.repo.FindByUUID(uuidStr)
	if err != nil {
		return nil, nil, ErrNotFound
	}
	old := *section

	section.Name = in.Name
	section.Icon = in.Icon
	section.Order = in.Order
	section.UpdatedBy = in.ActorID

	if err := s.repo.Save(section); err != nil {
		return nil, nil, err
	}
	return &old, section, nil
}

func (s *Service) Delete(uuidStr string, actorID *uint64) (*MenuSection, error) {
	section, err := s.repo.FindByUUID(uuidStr)
	if err != nil {
		return nil, ErrNotFound
	}
	if err := s.repo.Delete(section, actorID); err != nil {
		return nil, err
	}
	return section, nil
}

type ReorderItem struct {
	UUID  string
	Order int
}

func (s *Service) Reorder(items []ReorderItem) {
	for _, item := range items {
		_ = s.repo.UpdateOrder(item.UUID, item.Order)
	}
}
