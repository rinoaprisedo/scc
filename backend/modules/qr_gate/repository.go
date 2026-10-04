package qr_gate

import (
	"baseadmin/backend/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository isolates all GORM/SQL access for the qr_gate module.
type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

// QrGateWithStats is what List returns — a gate plus how many distinct
// participants have scanned it (each user can only redeem a given gate
// once, per the composite unique index on qr_gate_scans, so a plain row
// count is already a distinct-user count).
type QrGateWithStats struct {
	QrGate
	ScanCount int64 `json:"scan_count"`
}

func (r *Repository) List(p utils.Pagination) ([]QrGateWithStats, int64, error) {
	q := r.DB.Model(&QrGate{})

	if p.Search != "" {
		like := "%" + p.Search + "%"
		q = q.Where("name ILIKE ? OR code ILIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []QrGateWithStats
	err := q.Select("qr_gates.*, (SELECT COUNT(*) FROM qr_gate_scans WHERE qr_gate_scans.qr_gate_id = qr_gates.id) AS scan_count").
		Order(p.SortBy + " " + p.SortDir).Offset(p.Offset).Limit(p.Limit).Find(&list).Error
	return list, total, err
}

func (r *Repository) FindByUUID(uuidStr string) (*QrGate, error) {
	var gate QrGate
	if err := r.DB.Where("uuid = ?", uuidStr).First(&gate).Error; err != nil {
		return nil, err
	}
	return &gate, nil
}

func (r *Repository) FindByCodeTx(tx *gorm.DB, code string) (*QrGate, error) {
	var gate QrGate
	// Row lock so two concurrent scans of the same one-time-only gate can't
	// both pass the "already claimed" check before either commits.
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("code = ?", code).First(&gate).Error; err != nil {
		return nil, err
	}
	return &gate, nil
}

func (r *Repository) FindByUUIDTx(tx *gorm.DB, uuidStr string) (*QrGate, error) {
	var gate QrGate
	// Same row lock as FindByCodeTx — see there.
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("uuid = ?", uuidStr).First(&gate).Error; err != nil {
		return nil, err
	}
	return &gate, nil
}

type gateStatusRow struct {
	UUID       uuid.UUID
	Name       string
	Points     int
	IsReusable bool
	Scanned    bool
	ScanCount  int64
}

func (r *Repository) GateStatusForUser(userID uint64) ([]gateStatusRow, error) {
	var rows []gateStatusRow
	err := r.DB.Model(&QrGate{}).
		Select(`qr_gates.uuid, qr_gates.name, qr_gates.points, qr_gates.is_reusable,
			EXISTS (SELECT 1 FROM qr_gate_scans s WHERE s.qr_gate_id = qr_gates.id AND s.user_id = ?) AS scanned,
			(SELECT COUNT(*) FROM qr_gate_scans s WHERE s.qr_gate_id = qr_gates.id) AS scan_count`, userID).
		Order("qr_gates.name ASC").
		Scan(&rows).Error
	return rows, err
}

func (r *Repository) Create(gate *QrGate) error {
	return r.DB.Create(gate).Error
}

func (r *Repository) Save(gate *QrGate) error {
	return r.DB.Save(gate).Error
}

func (r *Repository) Delete(gate *QrGate, actorID *uint64) error {
	if actorID != nil {
		r.DB.Model(gate).Update("deleted_by", actorID)
	}
	return r.DB.Delete(gate).Error
}

func (r *Repository) CountScansForGateTx(tx *gorm.DB, gateID uint64) (int64, error) {
	var count int64
	err := tx.Model(&QrGateScan{}).Where("qr_gate_id = ?", gateID).Count(&count).Error
	return count, err
}

func (r *Repository) HasUserScannedTx(tx *gorm.DB, gateID, userID uint64) (bool, error) {
	var count int64
	err := tx.Model(&QrGateScan{}).Where("qr_gate_id = ? AND user_id = ?", gateID, userID).Count(&count).Error
	return count > 0, err
}

func (r *Repository) CreateScanTx(tx *gorm.DB, scan *QrGateScan) error {
	return tx.Create(scan).Error
}

// unscopedPreload drops GORM's default soft-delete scope on a preloaded
// association — QrGateScan is an append-only ledger (see its doc comment),
// so a scan's attribution shouldn't go blank just because the QrGate or User
// it points to was later soft-deleted; without this, Preload silently loads
// a zero-value struct (empty name) for a deleted row instead of erroring.
func unscopedPreload(db *gorm.DB) *gorm.DB { return db.Unscoped() }

// ListScansForGate backs the admin "who scanned this gate" view.
func (r *Repository) ListScansForGate(gateID uint64) ([]QrGateScan, error) {
	var list []QrGateScan
	err := r.DB.Preload("User", unscopedPreload).Where("qr_gate_id = ?", gateID).Order("created_at desc").Find(&list).Error
	return list, err
}

// ListScansForUser backs the participant's own scan history.
func (r *Repository) ListScansForUser(userID uint64) ([]QrGateScan, error) {
	var list []QrGateScan
	err := r.DB.Preload("QrGate", unscopedPreload).Where("user_id = ?", userID).Order("created_at desc").Find(&list).Error
	return list, err
}

// pesertaRoleName mirrors the same constant independently defined in
// blazer_sizes and peserta — kept local rather than a shared import, same
// precedent as blazer_sizes.Repository.
const pesertaRoleName = "Peserta"

// LeaderboardRow is one ranked participant — Points is a SUM across every
// qr_gate_scans and manual_points row for that user (0 via COALESCE for a peserta who hasn't
// scanned anything yet), not a stored counter.
type LeaderboardRow struct {
	ID        uint64
	UUID      uuid.UUID
	Name      string
	KtpNumber *string
	Points    int64
	ScanCount int64
	RankNo    int
}

// rankedLeaderboardSQL ranks every Peserta-role user by total QR gate points,
// highest first, name ascending to break ties. The LEFT JOIN (not INNER)
// keeps a peserta with zero points on the board at 0 points instead of
// omitting them. Scans and manual points are pre-aggregated in separate
// subqueries so joining both can't multiply each other's rows. ROW_NUMBER is computed over the whole set (not after
// LIMIT) so a single participant's position can be looked up even when
// they're far outside the top N. Aliased rank_no because RANK is reserved
// in MySQL 8.
const rankedLeaderboardSQL = `
SELECT * FROM (
	SELECT u.id, u.uuid, u.name, u.ktp_number,
		COALESCE(sc.points, 0) + COALESCE(mp.points, 0) AS points,
		COALESCE(sc.cnt, 0) AS scan_count,
		ROW_NUMBER() OVER (ORDER BY COALESCE(sc.points, 0) + COALESCE(mp.points, 0) DESC, u.name ASC, u.id ASC) AS rank_no
	FROM users u
	LEFT JOIN (SELECT user_id, SUM(points_awarded) AS points, COUNT(*) AS cnt FROM qr_gate_scans GROUP BY user_id) sc ON sc.user_id = u.id
	LEFT JOIN (SELECT user_id, SUM(points) AS points FROM manual_points WHERE deleted_at IS NULL GROUP BY user_id) mp ON mp.user_id = u.id
	WHERE u.deleted_at IS NULL AND u.role_id = (SELECT id FROM roles WHERE name = ?)
) ranked`

// Leaderboard returns the top `limit` ranked participants; limit <= 0 means
// the whole board (admin view).
func (r *Repository) Leaderboard(limit int) ([]LeaderboardRow, error) {
	var rows []LeaderboardRow
	q := rankedLeaderboardSQL + " ORDER BY rank_no"
	args := []interface{}{pesertaRoleName}
	if limit > 0 {
		q += " LIMIT ?"
		args = append(args, limit)
	}
	err := r.DB.Raw(q, args...).Scan(&rows).Error
	return rows, err
}

// LeaderboardPosition returns a single participant's ranked row, or nil if
// they aren't on the board (e.g. a non-Peserta account).
func (r *Repository) LeaderboardPosition(userID uint64) (*LeaderboardRow, error) {
	var rows []LeaderboardRow
	if err := r.DB.Raw(rankedLeaderboardSQL+" WHERE id = ?", pesertaRoleName, userID).Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

func (r *Repository) ListManualPointsForUser(userID uint64) ([]ManualPoint, error) {
	var list []ManualPoint
	err := r.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&list).Error
	return list, err
}

func (r *Repository) CreateManualPoint(mp *ManualPoint) error {
	return r.DB.Create(mp).Error
}

func (r *Repository) FindManualPoint(userID uint64, uuidStr string) (*ManualPoint, error) {
	var mp ManualPoint
	if err := r.DB.Where("user_id = ? AND uuid = ?", userID, uuidStr).First(&mp).Error; err != nil {
		return nil, err
	}
	return &mp, nil
}

func (r *Repository) DeleteManualPoint(mp *ManualPoint, actorID *uint64) error {
	if actorID != nil {
		r.DB.Model(mp).Update("deleted_by", actorID)
	}
	return r.DB.Delete(mp).Error
}

func (r *Repository) FindScanForUser(userID uint64, uuidStr string) (*QrGateScan, error) {
	var scan QrGateScan
	if err := r.DB.Preload("QrGate", unscopedPreload).Where("user_id = ? AND uuid = ?", userID, uuidStr).First(&scan).Error; err != nil {
		return nil, err
	}
	return &scan, nil
}

// DeleteScan hard-deletes — the ledger has no soft-delete column, and the
// row must actually be gone for the (qr_gate_id, user_id) unique index to
// let the participant check in to that gate again.
func (r *Repository) DeleteScan(scan *QrGateScan) error {
	return r.DB.Delete(scan).Error
}
