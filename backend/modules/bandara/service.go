package bandara

import (
	"errors"

	"baseadmin/backend/utils"
)

var ErrNotFound = errors.New("bandara not found")

// Service holds business rules for the bandara module.
type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(p utils.Pagination) ([]Bandara, int64, error) {
	return s.repo.List(p)
}

func (s *Service) Options() ([]Bandara, error) {
	return s.repo.All()
}

func (s *Service) Get(uuidStr string) (*Bandara, error) {
	bandara, err := s.repo.FindByUUID(uuidStr)
	if err != nil {
		return nil, ErrNotFound
	}
	return bandara, nil
}

type Input struct {
	Name    string
	ActorID *uint64
}

func (s *Service) Create(in Input) (*Bandara, error) {
	bandara := Bandara{Name: in.Name}
	bandara.CreatedBy = in.ActorID
	if err := s.repo.Create(&bandara); err != nil {
		return nil, err
	}
	return &bandara, nil
}

func (s *Service) Update(uuidStr string, in Input) (before *Bandara, after *Bandara, err error) {
	bandara, err := s.repo.FindByUUID(uuidStr)
	if err != nil {
		return nil, nil, ErrNotFound
	}
	old := *bandara

	bandara.Name = in.Name
	bandara.UpdatedBy = in.ActorID

	if err := s.repo.Save(bandara); err != nil {
		return nil, nil, err
	}
	return &old, bandara, nil
}

func (s *Service) Delete(uuidStr string, actorID *uint64) (*Bandara, error) {
	bandara, err := s.repo.FindByUUID(uuidStr)
	if err != nil {
		return nil, ErrNotFound
	}
	if err := s.repo.Delete(bandara, actorID); err != nil {
		return nil, err
	}
	return bandara, nil
}
