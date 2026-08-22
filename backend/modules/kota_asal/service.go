package kota_asal

import (
	"errors"

	"baseadmin/backend/utils"
)

var ErrNotFound = errors.New("kota asal not found")

// Service holds business rules for the kota_asal module.
type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(p utils.Pagination) ([]KotaAsal, int64, error) {
	return s.repo.List(p)
}

func (s *Service) Options() ([]KotaAsal, error) {
	return s.repo.All()
}

func (s *Service) Get(uuidStr string) (*KotaAsal, error) {
	kotaAsal, err := s.repo.FindByUUID(uuidStr)
	if err != nil {
		return nil, ErrNotFound
	}
	return kotaAsal, nil
}

type Input struct {
	Name     string
	Province string
	ActorID  *uint64
}

func (s *Service) Create(in Input) (*KotaAsal, error) {
	kotaAsal := KotaAsal{Name: in.Name, Province: in.Province}
	kotaAsal.CreatedBy = in.ActorID
	if err := s.repo.Create(&kotaAsal); err != nil {
		return nil, err
	}
	return &kotaAsal, nil
}

func (s *Service) Update(uuidStr string, in Input) (before *KotaAsal, after *KotaAsal, err error) {
	kotaAsal, err := s.repo.FindByUUID(uuidStr)
	if err != nil {
		return nil, nil, ErrNotFound
	}
	old := *kotaAsal

	kotaAsal.Name = in.Name
	kotaAsal.Province = in.Province
	kotaAsal.UpdatedBy = in.ActorID

	if err := s.repo.Save(kotaAsal); err != nil {
		return nil, nil, err
	}
	return &old, kotaAsal, nil
}

func (s *Service) Delete(uuidStr string, actorID *uint64) (*KotaAsal, error) {
	kotaAsal, err := s.repo.FindByUUID(uuidStr)
	if err != nil {
		return nil, ErrNotFound
	}
	if err := s.repo.Delete(kotaAsal, actorID); err != nil {
		return nil, err
	}
	return kotaAsal, nil
}
