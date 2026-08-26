package kota_asal

import "baseadmin/backend/utils"

type KotaAsal struct {
	utils.AuditModel
	Name     string `gorm:"type:varchar(255);not null" json:"name"`
	Province string `gorm:"type:varchar(255)" json:"province"`
	// Description holds search keywords for the website's kota-asal
	// combobox — e.g. a "Jabodetabek" entry can list "Jakarta, Bogor,
	// Depok, Tangerang, Bekasi" here so participants searching any of
	// those city names still find it, without those names cluttering
	// the displayed option label itself.
	Description string `gorm:"type:varchar(500)" json:"description"`
}

func (KotaAsal) TableName() string { return "kota_asal" }
