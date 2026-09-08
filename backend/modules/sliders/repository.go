package sliders

import (
	"baseadmin/backend/utils"

	"gorm.io/gorm"
)

// Repository isolates all GORM/SQL access for the sliders module.
type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) List(p utils.Pagination, sliderType string) ([]Slider, int64, error) {
	q := r.DB.Model(&Slider{})
	if sliderType != "" {
		q = q.Where("type = ?", sliderType)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// "order" is a reserved SQL keyword — quote it when it's the requested
	// sort column, unlike every other column name this endpoint accepts.
	sortBy := p.SortBy
	if sortBy == "order" {
		sortBy = `"order"`
	}

	var list []Slider
	err := q.Order(sortBy + " " + p.SortDir).Offset(p.Offset).Limit(p.Limit).Find(&list).Error
	return list, total, err
}

// ListActive returns every active slider ordered for the website's
// carousel — id asc breaks ties among rows sharing the same order (e.g.
// freshly created rows before anyone has dragged yet).
func (r *Repository) ListActive() ([]Slider, error) {
	var list []Slider
	err := r.DB.Where("is_active = ?", true).Order(`"order" ASC, id ASC`).Find(&list).Error
	return list, err
}

// NextOrder returns the order value a newly created row of the given type
// should get so it's appended after every existing row of that same type —
// mobile and desktop are sequenced independently since they back separate
// carousels.
func (r *Repository) NextOrder(sliderType string) (int, error) {
	var maxOrder *int
	if err := r.DB.Model(&Slider{}).Where("type = ?", sliderType).Select(`MAX("order")`).Scan(&maxOrder).Error; err != nil {
		return 0, err
	}
	if maxOrder == nil {
		return 0, nil
	}
	return *maxOrder + 1, nil
}

func (r *Repository) UpdateOrder(uuidStr string, order int) error {
	return r.DB.Model(&Slider{}).Where("uuid = ?", uuidStr).Update("order", order).Error
}

func (r *Repository) FindByUUID(uuidStr string) (*Slider, error) {
	var slider Slider
	if err := r.DB.Where("uuid = ?", uuidStr).First(&slider).Error; err != nil {
		return nil, err
	}
	return &slider, nil
}

func (r *Repository) Create(slider *Slider) error {
	return r.DB.Create(slider).Error
}

// Save persists every column, including zero values — IsActive can
// legitimately be toggled to false, and plain db.Save(struct) skips
// zero-value fields on UPDATE. Same precedent as
// qris_cross_border.Repository.Save.
func (r *Repository) Save(slider *Slider) error {
	return r.DB.Model(slider).Select("*").Updates(slider).Error
}

func (r *Repository) Delete(slider *Slider, actorID *uint64) error {
	if actorID != nil {
		r.DB.Model(slider).Update("deleted_by", actorID)
	}
	return r.DB.Delete(slider).Error
}
