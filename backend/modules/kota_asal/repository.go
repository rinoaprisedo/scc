package kota_asal

import (
	"baseadmin/backend/utils"

	"gorm.io/gorm"
)

// Repository isolates all GORM/SQL access for the kota_asal module.
type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) List(p utils.Pagination) ([]KotaAsal, int64, error) {
	q := r.DB.Model(&KotaAsal{})

	if p.Search != "" {
		like := "%" + p.Search + "%"
		q = q.Where("name ILIKE ? OR province ILIKE ? OR description ILIKE ?", like, like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []KotaAsal
	err := q.Order(p.SortBy + " " + p.SortDir).Offset(p.Offset).Limit(p.Limit).Find(&list).Error
	return list, total, err
}

// All returns every row ordered by name — backs the /kota-asal/options
// self-service lookup, which has no pagination since the table is small
// reference data.
func (r *Repository) All() ([]KotaAsal, error) {
	var list []KotaAsal
	err := r.DB.Order("name asc").Find(&list).Error
	return list, err
}

func (r *Repository) FindByUUID(uuidStr string) (*KotaAsal, error) {
	var kotaAsal KotaAsal
	if err := r.DB.Where("uuid = ?", uuidStr).First(&kotaAsal).Error; err != nil {
		return nil, err
	}
	return &kotaAsal, nil
}

func (r *Repository) Create(kotaAsal *KotaAsal) error {
	return r.DB.Create(kotaAsal).Error
}

func (r *Repository) Save(kotaAsal *KotaAsal) error {
	return r.DB.Save(kotaAsal).Error
}

func (r *Repository) Delete(kotaAsal *KotaAsal, actorID *uint64) error {
	if actorID != nil {
		r.DB.Model(kotaAsal).Update("deleted_by", actorID)
	}
	return r.DB.Delete(kotaAsal).Error
}
