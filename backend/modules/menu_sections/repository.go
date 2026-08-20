package menu_sections

import "gorm.io/gorm"

// Repository isolates all GORM/SQL access for the menu_sections module.
type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) List() ([]MenuSection, error) {
	var list []MenuSection
	err := r.DB.Order("\"order\" asc").Find(&list).Error
	return list, err
}

func (r *Repository) FindByUUID(uuidStr string) (*MenuSection, error) {
	var section MenuSection
	if err := r.DB.Where("uuid = ?", uuidStr).First(&section).Error; err != nil {
		return nil, err
	}
	return &section, nil
}

func (r *Repository) Create(section *MenuSection) error {
	return r.DB.Create(section).Error
}

func (r *Repository) Save(section *MenuSection) error {
	return r.DB.Save(section).Error
}

func (r *Repository) Delete(section *MenuSection, actorID *uint64) error {
	if actorID != nil {
		r.DB.Model(section).Update("deleted_by", actorID)
	}
	return r.DB.Delete(section).Error
}

func (r *Repository) UpdateOrder(uuidStr string, order int) error {
	return r.DB.Model(&MenuSection{}).Where("uuid = ?", uuidStr).Update("order", order).Error
}
