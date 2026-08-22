package peserta

import (
	"baseadmin/backend/modules/bandara"
	"baseadmin/backend/modules/users"
	"baseadmin/backend/utils"

	"gorm.io/gorm"
)

// pesertaRoleName is the fixed role every row managed by this module must
// carry. Every query is scoped by joining on it so the module can never
// see/edit/delete a user that isn't a Peserta, even if called with an
// arbitrary uuid.
const pesertaRoleName = "Peserta"

// Repository isolates all GORM/SQL access for the peserta module. It reuses
// the users.User model/table rather than a separate one — see the module
// doc for why.
type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) scope() *gorm.DB {
	return r.DB.Model(&users.User{}).
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.name = ?", pesertaRoleName)
}

// UserWithPoints is what List returns — a peserta plus their total QR gate
// points, summed from qr_gate_scans. Referenced by table name rather than
// importing the qr_gate package, to avoid a cross-module Go dependency for
// what's just a read-only reporting join.
type UserWithPoints struct {
	users.User
	TotalPoints int64 `json:"total_points"`
}

func (r *Repository) List(p utils.Pagination) ([]UserWithPoints, int64, error) {
	q := r.scope().Preload("Role").Preload("OriginCity").Preload("NearestAirport")

	if p.Search != "" {
		like := "%" + p.Search + "%"
		q = q.Where(
			"users.name ILIKE ? OR users.email ILIKE ? OR users.phone_number ILIKE ? OR users.ktp_number ILIKE ? OR users.nomor_ktp ILIKE ? OR users.passport_number ILIKE ?",
			like, like, like, like, like, like,
		)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []UserWithPoints
	err := q.Select("users.*, (SELECT COALESCE(SUM(points_awarded), 0) FROM qr_gate_scans WHERE qr_gate_scans.user_id = users.id) AS total_points").
		Order("users." + p.SortBy + " " + p.SortDir).Offset(p.Offset).Limit(p.Limit).Find(&list).Error
	return list, total, err
}

// ListAll returns every matching row unpaginated — backs CSV/Excel export,
// which needs the full filtered result set rather than one page of it.
func (r *Repository) ListAll(search string) ([]users.User, error) {
	q := r.scope().Preload("Role").Preload("OriginCity").Preload("NearestAirport")

	if search != "" {
		like := "%" + search + "%"
		q = q.Where(
			"users.name ILIKE ? OR users.email ILIKE ? OR users.phone_number ILIKE ? OR users.ktp_number ILIKE ? OR users.nomor_ktp ILIKE ? OR users.passport_number ILIKE ?",
			like, like, like, like, like, like,
		)
	}

	var list []users.User
	err := q.Order("users.created_at desc").Find(&list).Error
	return list, err
}

func (r *Repository) FindByUUID(uuidStr string) (*users.User, error) {
	var user users.User
	err := r.scope().Preload("Role").Preload("OriginCity").Preload("NearestAirport").
		Where("users.uuid = ?", uuidStr).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByUserID is FindByUUID keyed by numeric ID instead — used by the
// self-service /peserta/me endpoints, which know the caller's ID from the
// session rather than a path param. Still routed through r.scope(), so a
// non-Peserta session can never reach a row here.
func (r *Repository) FindByUserID(id uint64) (*users.User, error) {
	var user users.User
	err := r.scope().Preload("Role").Preload("OriginCity").Preload("NearestAirport").
		Where("users.id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) Create(user *users.User) error {
	return r.DB.Create(user).Error
}

// Save persists every column on user, including zero values. Plain
// db.Save(user) silently SKIPS zero-value fields (nil pointers, empty
// strings) in its generated UPDATE — so switching origin_city_uuid back to
// "" to clear it (e.g. picking "Other" after a real city was set) would
// never actually null out the old origin_city_id in the database, even
// though the in-memory struct and every API response looked correct.
// Select("*") forces every column to be included regardless of zero-ness.
func (r *Repository) Save(user *users.User) error {
	return r.DB.Model(user).Select("*").Updates(user).Error
}

func (r *Repository) Delete(user *users.User, actorID *uint64) error {
	if actorID != nil {
		r.DB.Model(user).Update("deleted_by", actorID)
	}
	return r.DB.Delete(user).Error
}

// DeleteAllSessions mirrors users.Repository.DeleteAllSessions — removes
// every DB session-mirror row for a user, used alongside a Redis-side
// force-logout-all after a status change.
func (r *Repository) DeleteAllSessions(userID uint64) error {
	return r.DB.Where("user_id = ?", userID).Delete(&users.Session{}).Error
}

// PesertaRoleID resolves the fixed "Peserta" role's numeric ID, used to force
// every row created/kept by this module onto that role regardless of what a
// caller might otherwise pass.
func (r *Repository) PesertaRoleID() (uint64, error) {
	var id uint64
	err := r.DB.Table("roles").Select("id").Where("name = ?", pesertaRoleName).Scan(&id).Error
	return id, err
}

// ResolveCityID resolves a kota_asal UUID to its numeric ID. An empty uuid
// clears the relation (returns nil, nil), matching the nil-clears-relation
// convention already used by users.Repository.SetRole.
func (r *Repository) ResolveCityID(uuidStr string) (*uint64, error) {
	if uuidStr == "" {
		return nil, nil
	}
	var id uint64
	if err := r.DB.Table("kota_asal").Select("id").Where("uuid = ?", uuidStr).Scan(&id).Error; err != nil {
		return nil, err
	}
	if id == 0 {
		return nil, nil
	}
	return &id, nil
}

// ResolveAirportID resolves a bandara UUID to its numeric ID. Same
// empty-clears-relation convention as ResolveCityID.
func (r *Repository) ResolveAirportID(uuidStr string) (*uint64, error) {
	if uuidStr == "" {
		return nil, nil
	}
	var id uint64
	if err := r.DB.Table("bandara").Select("id").Where("uuid = ?", uuidStr).Scan(&id).Error; err != nil {
		return nil, err
	}
	if id == 0 {
		return nil, nil
	}
	return &id, nil
}

// FindCityIDByName looks up a kota_asal row by case-insensitive name — used
// by Excel import, which gets a plain city name in a spreadsheet cell
// rather than a uuid a dropdown would supply. Returns (nil, nil), not an
// error, when nothing matches: the caller falls back to storing the text as
// origin_city_other, the same "Other" path the website's own form uses.
func (r *Repository) FindCityIDByName(name string) (*uint64, error) {
	var id uint64
	if err := r.DB.Table("kota_asal").Select("id").Where("LOWER(name) = LOWER(?)", name).Scan(&id).Error; err != nil {
		return nil, err
	}
	if id == 0 {
		return nil, nil
	}
	return &id, nil
}

// FindAirportIDByName looks up a bandara row by case-insensitive name,
// read-only — mirrors FindCityIDByName's nil-means-not-found convention.
// Used by Excel import validation (a dry run must never write to the DB)
// and by the real import's prepare phase; actually creating a missing
// airport is deferred to CreateAirportTx, run only once every row in the
// file has passed validation.
func (r *Repository) FindAirportIDByName(name string) (*uint64, error) {
	var id uint64
	if err := r.DB.Table("bandara").Select("id").Where("LOWER(name) = LOWER(?)", name).Scan(&id).Error; err != nil {
		return nil, err
	}
	if id == 0 {
		return nil, nil
	}
	return &id, nil
}

// CreateAirportTx creates a new bandara row for an import row whose airport
// name didn't match existing master data, using the given transaction so it
// rolls back along with the rest of the import if a later row fails.
func (r *Repository) CreateAirportTx(tx *gorm.DB, name string, actorID *uint64) (*uint64, error) {
	newAirport := bandara.Bandara{Name: name}
	newAirport.CreatedBy = actorID
	if err := tx.Create(&newAirport).Error; err != nil {
		return nil, err
	}
	return &newAirport.ID, nil
}

// NIKExists/EmailExists back Excel import validation's duplicate check
// against already-stored peserta — GORM's default soft-delete scope means
// these only ever see non-deleted rows, matching the partial unique indexes
// on ktp_number/email (see migrate.go).
func (r *Repository) NIKExists(nik string) (bool, error) {
	var count int64
	err := r.DB.Model(&users.User{}).Where("ktp_number = ?", nik).Count(&count).Error
	return count > 0, err
}

func (r *Repository) EmailExists(email string) (bool, error) {
	var count int64
	err := r.DB.Model(&users.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// AllCityNames/AllAirportNames back the Excel import template's dropdown
// columns for Kota Asal/Bandara Terdekat — the full unfiltered master-data
// name list, written to a hidden helper sheet the visible columns' data
// validation references (see import.go).
func (r *Repository) AllCityNames() ([]string, error) {
	var names []string
	err := r.DB.Table("kota_asal").Order("name").Pluck("name", &names).Error
	return names, err
}

func (r *Repository) AllAirportNames() ([]string, error) {
	var names []string
	err := r.DB.Table("bandara").Order("name").Pluck("name", &names).Error
	return names, err
}
