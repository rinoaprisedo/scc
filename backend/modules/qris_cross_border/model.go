package qris_cross_border

import (
	"baseadmin/backend/modules/users"
	"baseadmin/backend/utils"
)

// Status values a submission moves through. Pending is the only status a
// new row is ever created with — an async job (see ocr.go/worker.go) reads
// the proof image, then transitions the row to WaitingApproval or Rejected.
// Pending is deliberately excluded from validStatuses in service.go: it's a
// system-managed transient state, not something an admin picks manually.
const (
	StatusPending         = "pending"
	StatusWaitingApproval = "waiting_approval"
	StatusApproved        = "approved"
	StatusRejected        = "rejected"
)

// QrisCrossBorder is an admin-recorded cross-border QRIS top-up: which
// peserta it's for, proof-of-payment image, and the amount in both
// currencies (foreign currency the participant paid, IDR the equivalent
// credited). MerchantName/ReferenceNumber/RejectReason are filled in by the
// async AI-extraction job, not at creation time.
type QrisCrossBorder struct {
	utils.AuditModel
	PesertaID uint64 `gorm:"not null;index" json:"-"`
	// Never serialized directly — same precedent as qr_gate.QrGateScan.User:
	// a struct field's zero value still marshals regardless of `omitempty`.
	// Service.toResponse flattens the fields the frontend actually needs.
	Peserta *users.User `gorm:"foreignKey:PesertaID" json:"-"`
	Image   string      `gorm:"type:varchar(255);not null" json:"image"`
	// NominalAsing is generic (any foreign currency the screenshot shows,
	// e.g. THB or MYR) — no separate currency-code column, matching the
	// scope of what was asked. NominalRupiah is always the IDR side.
	NominalAsing    float64 `gorm:"type:decimal(15,2);not null;default:0" json:"nominal_asing"`
	NominalRupiah   float64 `gorm:"type:decimal(15,2);not null;default:0" json:"nominal_rupiah"`
	MerchantName    *string `gorm:"type:varchar(255)" json:"merchant_name"`
	ReferenceNumber *string `gorm:"type:varchar(100)" json:"reference_number"`
	RejectReason    *string `gorm:"type:varchar(255)" json:"reject_reason"`
	Status          string  `gorm:"type:varchar(20);not null;default:pending" json:"status"`
}

func (QrisCrossBorder) TableName() string { return "qris_cross_borders" }
