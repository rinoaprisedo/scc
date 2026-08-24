package blazer_sizes

import "baseadmin/backend/utils"

type BlazerSize struct {
	utils.AuditModel
	Size  string `gorm:"type:varchar(10);not null" json:"size"`
	Stock int    `gorm:"not null;default:0" json:"stock"`
	// Order drives display order in both the admin list and the website's
	// size picker — set automatically (append-to-end) on create, and only
	// ever mutated in bulk via the drag-reorder endpoint. "order" is a
	// reserved SQL keyword, hence the explicit column name and the quoting
	// wherever it appears in a raw ORDER BY.
	Order int `gorm:"column:order;not null;default:0" json:"order"`
}

func (BlazerSize) TableName() string { return "blazer_sizes" }
