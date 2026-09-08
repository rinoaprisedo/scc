package blazer_sizes

import (
	"baseadmin/backend/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// pesertaRoleName mirrors the same constant independently defined in the
// peserta and dashboard packages — kept local rather than a shared import,
// since this is just a read-only string comparison, not a dependency worth
// coupling modules over.
const pesertaRoleName = "Peserta"

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

// OptionRow is one blazer_sizes row plus how many peserta currently have it
// assigned — Stock is an admin-set initial quota, not a live remaining
// counter (it's never decremented anywhere in the app), so any caller
// needing "how many are actually left" must compute Stock minus Assigned
// themselves, same as peserta.Repository.CountByBlazerSize and
// dashboard.Repository.BlazerSizeStockRows do independently.
type OptionRow struct {
	UUID     uuid.UUID
	Size     string
	Stock    int
	Assigned int64
}

// OptionsWithRemaining returns every row ordered by the admin-controlled
// display order — backs the /blazer-sizes/options self-service lookup the
// website's and admin's size pickers render from, so dragging a size in the
// admin list reorders both. id asc breaks ties among rows that share the
// same order (e.g. freshly seeded rows before anyone has dragged yet). The
// left join counts only non-deleted Peserta-role users, matching
// peserta.Repository's own scope.
func (r *Repository) OptionsWithRemaining() ([]OptionRow, error) {
	var rows []OptionRow
	err := r.DB.Table("blazer_sizes AS bs").
		Select(`bs.uuid AS uuid, bs.size AS size, bs.stock AS stock, COUNT(u.id) AS assigned`).
		Joins(`LEFT JOIN users u ON u.blazer_size = bs.size AND u.deleted_at IS NULL AND u.role_id = (SELECT id FROM roles WHERE name = ?)`, pesertaRoleName).
		Where("bs.deleted_at IS NULL").
		Group("bs.id").
		Order(`bs."order" ASC, bs.id ASC`).
		Scan(&rows).Error
	return rows, err
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

// FindBySize looks up a row by its exact size label — backs the peserta
// module's stock-availability check, which receives a size string (from
// the admin form or website) rather than a blazer_sizes UUID.
func (r *Repository) FindBySize(size string) (*BlazerSize, error) {
	var blazerSize BlazerSize
	if err := r.DB.Where("size = ?", size).First(&blazerSize).Error; err != nil {
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
