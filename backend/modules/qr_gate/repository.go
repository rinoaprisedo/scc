package qr_gate

import (
	"baseadmin/backend/utils"

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

// ListScansForGate backs the admin "who scanned this gate" view.
func (r *Repository) ListScansForGate(gateID uint64) ([]QrGateScan, error) {
	var list []QrGateScan
	err := r.DB.Preload("User").Where("qr_gate_id = ?", gateID).Order("created_at desc").Find(&list).Error
	return list, err
}

// ListScansForUser backs the participant's own scan history.
func (r *Repository) ListScansForUser(userID uint64) ([]QrGateScan, error) {
	var list []QrGateScan
	err := r.DB.Preload("QrGate").Where("user_id = ?", userID).Order("created_at desc").Find(&list).Error
	return list, err
}
