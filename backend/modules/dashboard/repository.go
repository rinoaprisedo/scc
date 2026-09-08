package dashboard

import (
	"baseadmin/backend/modules/users"

	"gorm.io/gorm"
)

// pesertaRoleName mirrors peserta.Repository's scoping — kept as a local
// constant rather than importing the peserta package, since this module only
// needs the read-only role filter, not any of peserta's business logic.
const pesertaRoleName = "Peserta"

// Repository isolates all GORM/SQL access for the dashboard module. Every
// query here is read-only aggregate reporting — no writes.
type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) scope() *gorm.DB {
	return r.DB.Model(&users.User{}).
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.name = ?", pesertaRoleName)
}

func (r *Repository) TotalPeserta() (int64, error) {
	var total int64
	err := r.scope().Count(&total).Error
	return total, err
}

func (r *Repository) LoggedInCount() (int64, error) {
	var count int64
	err := r.scope().Where("users.last_login_at IS NOT NULL").Count(&count).Error
	return count, err
}

// GroupCount is the shape for any "GROUP BY nullable varchar column" tally —
// reused for both attendance status and dietary restriction breakdowns.
type GroupCount struct {
	Value *string
	Count int64
}

func (r *Repository) AttendanceCounts() ([]GroupCount, error) {
	var rows []GroupCount
	err := r.scope().Select("users.attendance_status AS value, COUNT(*) AS count").
		Group("users.attendance_status").Scan(&rows).Error
	return rows, err
}

func (r *Repository) DietaryCounts() ([]GroupCount, error) {
	var rows []GroupCount
	err := r.scope().Select("users.dietary_restriction AS value, COUNT(*) AS count").
		Group("users.dietary_restriction").Scan(&rows).Error
	return rows, err
}

// GateCount is one QR gate plus how many distinct peserta have scanned it.
// The composite unique index on qr_gate_scans(qr_gate_id, user_id) already
// guarantees at most one scan row per peserta per gate, so a plain row count
// doubles as a distinct-participant count (same assumption qr_gate's own
// QrGateWithStats relies on).
type GateCount struct {
	GateName string
	Count    int64
}

func (r *Repository) GateScanCounts() ([]GateCount, error) {
	var rows []GateCount
	err := r.DB.Table("qr_gates").
		Select("qr_gates.name AS gate_name, COUNT(qr_gate_scans.id) AS count").
		Joins("LEFT JOIN qr_gate_scans ON qr_gate_scans.qr_gate_id = qr_gates.id").
		Where("qr_gates.deleted_at IS NULL").
		Group("qr_gates.id, qr_gates.name").
		Order("qr_gates.name ASC").
		Scan(&rows).Error
	return rows, err
}

// BlazerSizeStockRow is one blazer_sizes row's label, its admin-set initial
// quota (Stock), and how many peserta currently have that size assigned.
// Stock is never auto-decremented anywhere in the app (see
// peserta.Service.checkBlazerSizeStock) — it's the initial quota an admin
// typed in, so "remaining" always has to be computed as Stock minus Assigned
// rather than read off a live counter column.
type BlazerSizeStockRow struct {
	Size     string
	Stock    int
	Assigned int64
}

// BlazerSizeStockRows lists every active blazer size in the same admin-controlled
// display order the peserta form's dropdown uses, so the dashboard reads
// top-to-bottom in the order admins expect. The left join counts only
// non-deleted Peserta-role users, matching peserta.Repository's own scope.
func (r *Repository) BlazerSizeStockRows() ([]BlazerSizeStockRow, error) {
	var rows []BlazerSizeStockRow
	err := r.DB.Table("blazer_sizes AS bs").
		Select(`bs.size AS size, bs.stock AS stock, COUNT(u.id) AS assigned`).
		Joins(`LEFT JOIN users u ON u.blazer_size = bs.size AND u.deleted_at IS NULL AND u.role_id = (SELECT id FROM roles WHERE name = ?)`, pesertaRoleName).
		Where("bs.deleted_at IS NULL").
		Group(`bs.id, bs.size, bs.stock, bs."order"`).
		Order(`bs."order" ASC, bs.id ASC`).
		Scan(&rows).Error
	return rows, err
}
