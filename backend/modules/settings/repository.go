package settings

import "gorm.io/gorm"

// Repository isolates all GORM/SQL access for the settings module.
type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) List() ([]Setting, error) {
	var list []Setting
	err := r.DB.Find(&list).Error
	return list, err
}

// FindByKey backs the multi-image endpoints (UploadImages/RemoveImage),
// which need to read-modify-write a single row's Value rather than upsert
// blind, unlike every other write in this module.
func (r *Repository) FindByKey(key string) (*Setting, error) {
	var setting Setting
	if err := r.DB.Where("key = ?", key).First(&setting).Error; err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *Repository) Upsert(key, value string, settingType SettingType) error {
	var setting Setting
	result := r.DB.Where("key = ?", key).First(&setting)
	if result.Error != nil {
		setting = Setting{Key: key, Value: value, Type: settingType}
		return r.DB.Create(&setting).Error
	}
	setting.Value = value
	return r.DB.Save(&setting).Error
}
