package sliders

import "baseadmin/backend/utils"

// Type values a slider can be tagged with — the website shows Mobile
// sliders in its mobile-layout carousel slot and Desktop sliders in its
// desktop-layout slot, never mixed, since the two slots sit in different
// places in the page layout at different breakpoints.
const (
	TypeMobile  = "mobile"
	TypeDesktop = "desktop"
)

// ValidTypes gates Create/Update against typos — a slider with a stray type
// would never appear in either website slot.
var ValidTypes = map[string]bool{
	TypeMobile:  true,
	TypeDesktop: true,
}

// Slider is a homepage banner image shown to logged-in peserta on the
// website — image plus which layout slot it belongs to, and the same
// Order/IsActive controls as blazer_sizes so an admin can sequence and
// hide/show entries without deleting them.
type Slider struct {
	utils.AuditModel
	Image string `gorm:"type:varchar(255);not null" json:"image"`
	Type  string `gorm:"type:varchar(20);not null;default:desktop" json:"type"`
	// "order" is a reserved SQL keyword, hence the explicit column name and
	// the quoting wherever it appears in a raw ORDER BY — same as
	// blazer_sizes.BlazerSize.Order. Scoped per Type (see Repository.NextOrder)
	// since the mobile and desktop carousels are sequenced independently.
	Order    int  `gorm:"column:order;not null;default:0" json:"order"`
	IsActive bool `gorm:"not null;default:true" json:"is_active"`
}

func (Slider) TableName() string { return "sliders" }
