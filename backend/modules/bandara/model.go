package bandara

import "baseadmin/backend/utils"

type Bandara struct {
	utils.AuditModel
	Name string `gorm:"type:varchar(255);not null" json:"name"`
}

func (Bandara) TableName() string { return "bandara" }
