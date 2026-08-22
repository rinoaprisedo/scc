package bandara

import (
	"baseadmin/backend/utils"

	"gorm.io/gorm"
)

// Repository isolates all GORM/SQL access for the bandara module.
type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) List(p utils.Pagination) ([]Bandara, int64, error) {
	q := r.DB.Model(&Bandara{})

	if p.Search != "" {
		like := "%" + p.Search + "%"
		q = q.Where("name ILIKE ?", like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []Bandara
	err := q.Order(p.SortBy + " " + p.SortDir).Offset(p.Offset).Limit(p.Limit).Find(&list).Error
	return list, total, err
}

// All returns every row ordered by name — backs the /bandara/options
// self-service lookup, which has no pagination since the table is small
// reference data.
func (r *Repository) All() ([]Bandara, error) {
	var list []Bandara
	err := r.DB.Order("name asc").Find(&list).Error
	return list, err
}

func (r *Repository) FindByUUID(uuidStr string) (*Bandara, error) {
	var bandara Bandara
	if err := r.DB.Where("uuid = ?", uuidStr).First(&bandara).Error; err != nil {
		return nil, err
	}
	return &bandara, nil
}

func (r *Repository) Create(bandara *Bandara) error {
	return r.DB.Create(bandara).Error
}

func (r *Repository) Save(bandara *Bandara) error {
	return r.DB.Save(bandara).Error
}

func (r *Repository) Delete(bandara *Bandara, actorID *uint64) error {
	if actorID != nil {
		r.DB.Model(bandara).Update("deleted_by", actorID)
	}
	return r.DB.Delete(bandara).Error
}
