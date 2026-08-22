package qr_gate

import (
	"time"

	"baseadmin/backend/modules/users"
	"baseadmin/backend/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// QrGate is an admin-defined checkpoint: a custom text code (printed as a QR
// at an event location), a point value, and whether it can be claimed by
// many participants or only the very first one to scan it.
type QrGate struct {
	utils.AuditModel
	Name       string `gorm:"type:varchar(255);not null" json:"name"`
	Code       string `gorm:"type:varchar(100);not null" json:"code"`
	Points     int    `gorm:"not null;default:0" json:"points"`
	// No `default:true` tag — GORM's struct Create omits zero-value fields
	// that carry a `default` tag, so an explicit `false` here would silently
	// fall back to the DB default. The service always sets a real value from
	// admin input, so no DB-level default is needed.
	IsReusable bool `gorm:"not null" json:"is_reusable"`
}

func (QrGate) TableName() string { return "qr_gates" }

// QrGateScan is an append-only redemption ledger, same precedent as
// activity_logs.ActivityLog (no soft-delete — history shouldn't disappear).
// The composite unique index on (qr_gate_id, user_id) is what stops a single
// participant from claiming the same gate's points twice, regardless of
// whether the gate itself is reusable.
type QrGateScan struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"-"`
	UUID          uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"uuid"`
	QrGateID      uint64    `gorm:"not null;uniqueIndex:idx_qr_gate_scans_gate_user" json:"-"`
	UserID        uint64    `gorm:"not null;uniqueIndex:idx_qr_gate_scans_gate_user" json:"-"`
	PointsAwarded int       `gorm:"not null" json:"points_awarded"`
	CreatedAt     time.Time `json:"scanned_at"`

	// Never serialized directly — a struct field's zero value still gets
	// marshaled by encoding/json regardless of `omitempty`, so an
	// un-preloaded relation would otherwise leak an empty object. Callers
	// read these in Go and shape an explicit response DTO instead (see
	// Service.MyScans / Service.GateScans).
	QrGate QrGate     `gorm:"foreignKey:QrGateID" json:"-"`
	User   users.User `gorm:"foreignKey:UserID" json:"-"`
}

func (QrGateScan) TableName() string { return "qr_gate_scans" }

func (s *QrGateScan) BeforeCreate(tx *gorm.DB) error {
	if s.UUID == uuid.Nil {
		s.UUID = uuid.New()
	}
	return nil
}
