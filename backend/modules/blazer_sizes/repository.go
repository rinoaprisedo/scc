package blazer_sizes

import (
	"baseadmin/backend/utils"

	"gorm.io/gorm"
)

// Repository isolates all GORM/SQL access for the blazer_sizes module.
type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) List(p utils.Pagination) ([]BlazerSize, int64, error) {
	q := r.DB.Model(&BlazerSize{})

	if p.Search != "" {
		q = q.Where("size ILIKE ?", "%"+p.Search+"%")
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

	var list []BlazerSize
	err := q.Order(sortBy + " " + p.SortDir).Offset(p.Offset).Limit(p.Limit).Find(&list).Error
	return list, total, err
}

// All returns every row ordered by the admin-controlled display order —
// backs the /blazer-sizes/options self-service lookup the website's size
// picker renders from, so dragging a size in the admin list reorders the
// website's options too. id asc breaks ties among rows that share the same
// order (e.g. freshly seeded rows before anyone has dragged yet).
func (r *Repository) All() ([]BlazerSize, error) {
	var list []BlazerSize
	err := r.DB.Order(`"order" asc, id asc`).Find(&list).Error
	return list, err
}

// NextOrder returns the order value a newly created row should get so it's
// appended after every existing row.
func (r *Repository) NextOrder() (int, error) {
	var maxOrder *int
	if err := r.DB.Model(&BlazerSize{}).Select(`MAX("order")`).Scan(&maxOrder).Error; err != nil {
		return 0, err
	}
	if maxOrder == nil {
		return 0, nil
	}
	return *maxOrder + 1, nil
}

func (r *Repository) UpdateOrder(uuidStr string, order int) error {
	return r.DB.Model(&BlazerSize{}).Where("uuid = ?", uuidStr).Update("order", order).Error
}

func (r *Repository) FindByUUID(uuidStr string) (*BlazerSize, error) {
	var blazerSize BlazerSize
	if err := r.DB.Where("uuid = ?", uuidStr).First(&blazerSize).Error; err != nil {
		return nil, err
	}
	return &blazerSize, nil
}

func (r *Repository) Create(blazerSize *BlazerSize) error {
	return r.DB.Create(blazerSize).Error
}

func (r *Repository) Save(blazerSize *BlazerSize) error {
	return r.DB.Save(blazerSize).Error
}

func (r *Repository) Delete(blazerSize *BlazerSize, actorID *uint64) error {
	if actorID != nil {
		r.DB.Model(blazerSize).Update("deleted_by", actorID)
	}
	return r.DB.Delete(blazerSize).Error
}
