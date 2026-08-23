package qris_cross_border

import (
	"baseadmin/backend/utils"

	"gorm.io/gorm"
)

// Repository isolates all GORM/SQL access for the qris_cross_border module.
type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

// List always qualifies its own table's columns (qris_cross_borders.*) —
// search/pesertaUUID join against users, which also has a created_at/status
// column, so an unqualified ORDER BY/WHERE would be ambiguous.
func (r *Repository) List(p utils.Pagination, status, pesertaUUID string) ([]QrisCrossBorder, int64, error) {
	q := r.DB.Model(&QrisCrossBorder{}).Preload("Peserta")

	if status != "" {
		q = q.Where("qris_cross_borders.status = ?", status)
	}
	if p.Search != "" || pesertaUUID != "" {
		q = q.Joins("JOIN users ON users.id = qris_cross_borders.peserta_id")
	}
	if p.Search != "" {
		like := "%" + p.Search + "%"
		q = q.Where("users.name ILIKE ? OR users.ktp_number ILIKE ?", like, like)
	}
	if pesertaUUID != "" {
		q = q.Where("users.uuid = ?", pesertaUUID)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []QrisCrossBorder
	err := q.Order("qris_cross_borders." + p.SortBy + " " + p.SortDir).
		Offset(p.Offset).Limit(p.Limit).Find(&list).Error
	return list, total, err
}

// ListByPeserta returns a page of a peserta's own submissions, newest
// first — backs the website's self-service history list.
func (r *Repository) ListByPeserta(pesertaID uint64, p utils.Pagination) ([]QrisCrossBorder, int64, error) {
	q := r.DB.Model(&QrisCrossBorder{}).Where("peserta_id = ?", pesertaID)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []QrisCrossBorder
	err := q.Preload("Peserta").Order("created_at DESC").Offset(p.Offset).Limit(p.Limit).Find(&list).Error
	return list, total, err
}

func (r *Repository) FindByUUID(uuidStr string) (*QrisCrossBorder, error) {
	var row QrisCrossBorder
	if err := r.DB.Preload("Peserta").Where("uuid = ?", uuidStr).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *Repository) Create(row *QrisCrossBorder) error {
	return r.DB.Create(row).Error
}

// Save persists every column, including zero values (NominalAsing/Rupiah
// can legitimately be 0). Plain db.Save(struct) skips zero-value fields on
// UPDATE — see peserta.Repository.Save for the same precedent/explanation.
func (r *Repository) Save(row *QrisCrossBorder) error {
	return r.DB.Model(row).Select("*").Updates(row).Error
}

// ReferenceNumberExists backs duplicate detection — both the OCR job's
// automatic check and Service.Update's manual-edit guard. excludeUUID lets
// a record's own (already-saved) reference number not count as a collision
// against itself.
func (r *Repository) ReferenceNumberExists(refNumber, excludeUUID string) (bool, error) {
	var count int64
	err := r.DB.Model(&QrisCrossBorder{}).
		Where("reference_number = ? AND uuid != ?", refNumber, excludeUUID).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) Delete(row *QrisCrossBorder, actorID *uint64) error {
	if actorID != nil {
		r.DB.Model(row).Update("deleted_by", actorID)
	}
	return r.DB.Delete(row).Error
}

// ResolvePesertaID resolves a peserta UUID to its numeric user id, scoped to
// the fixed "Peserta" role only — mirrors peserta.Repository's role scoping
// so this module can never attach a record to a non-Peserta user.
func (r *Repository) ResolvePesertaID(uuidStr string) (uint64, error) {
	var id uint64
	err := r.DB.Table("users").
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.name = ? AND users.uuid = ?", "Peserta", uuidStr).
		Select("users.id").Scan(&id).Error
	if err != nil {
		return 0, err
	}
	if id == 0 {
		return 0, gorm.ErrRecordNotFound
	}
	return id, nil
}
