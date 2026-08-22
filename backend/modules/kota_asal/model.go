package kota_asal

import "baseadmin/backend/utils"

type KotaAsal struct {
	utils.AuditModel
	Name     string `gorm:"type:varchar(255);not null" json:"name"`
	Province string `gorm:"type:varchar(255)" json:"province"`
}

func (KotaAsal) TableName() string { return "kota_asal" }
