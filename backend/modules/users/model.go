package users

import (
	"time"

	"baseadmin/backend/modules/bandara"
	"baseadmin/backend/modules/kota_asal"
	"baseadmin/backend/modules/roles"
	"baseadmin/backend/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserStatus string

const (
	StatusActive    UserStatus = "active"
	StatusInactive  UserStatus = "inactive"
	StatusSuspended UserStatus = "suspended"
)

type User struct {
	utils.AuditModel
	Name string `gorm:"type:varchar(255);not null" json:"name"`
	// Uniqueness on email is enforced by a partial index (WHERE deleted_at IS
	// NULL) created in migrations, not by this tag — a plain uniqueIndex would
	// block re-using an email after the row holding it is soft-deleted.
	Email       string     `gorm:"type:varchar(255);index;not null" json:"email"`
	Password    string     `gorm:"type:varchar(255);not null" json:"-"`
	Avatar      *string    `gorm:"type:varchar(255)" json:"avatar"`
	Status      UserStatus `gorm:"type:varchar(20);default:active" json:"status"`
	LastLoginAt *time.Time `json:"last_login_at"`

	RoleID *uint64     `gorm:"index" json:"-"`
	Role   *roles.Role `gorm:"foreignKey:RoleID" json:"role,omitempty"`

	// Peserta (event participant) profile fields — nullable, unused by
	// non-Peserta users. Kept on this table rather than a separate profile
	// table by deliberate choice; see the peserta module for the CRUD that
	// manages them.
	Title              *string             `gorm:"type:varchar(10)" json:"title"`
	FirstName          *string             `gorm:"type:varchar(255)" json:"first_name"`
	MiddleName         *string             `gorm:"type:varchar(255)" json:"middle_name"`
	LastName           *string             `gorm:"type:varchar(255)" json:"last_name"`
	BirthDate          *time.Time          `json:"birth_date"`
	OriginCityID       *uint64             `gorm:"index" json:"-"`
	OriginCity         *kota_asal.KotaAsal `gorm:"foreignKey:OriginCityID" json:"origin_city,omitempty"`
	OriginCityOther    *string             `gorm:"type:varchar(255)" json:"origin_city_other"`
	NearestAirportID   *uint64             `gorm:"index" json:"-"`
	NearestAirport     *bandara.Bandara    `gorm:"foreignKey:NearestAirportID" json:"nearest_airport,omitempty"`
	DietaryRestriction *string             `gorm:"type:varchar(50)" json:"dietary_restriction"`
	PhoneNumber        *string             `gorm:"type:varchar(20)" json:"phone_number"`
	// Region/Cabang/Position are free-text organizational fields (the
	// participant's work region/branch office/job title) — admin-managed
	// like NomorMeja/Description below (manual entry or Excel import),
	// deliberately absent from SelfProfileInput/selfProfileRequest so a
	// participant can't set these via the website's self-service form.
	Region   *string `gorm:"type:varchar(255)" json:"region"`
	Cabang   *string `gorm:"type:varchar(255)" json:"cabang"`
	Position *string `gorm:"type:varchar(255)" json:"position"`
	// KtpNumber is the login identifier ("NIK") — required, unique
	// (idx_users_ktp_number_active), set at account creation/import and not
	// user-editable via the website's own onboarding form. NomorKtp is a
	// separate, optional, non-unique field for the KTP card number as
	// displayed profile data — the two are legally the same number in
	// Indonesia, but this app deliberately keeps them as distinct fields.
	KtpNumber      *string    `gorm:"type:varchar(30)" json:"ktp_number"`
	NomorKtp       *string    `gorm:"type:varchar(30)" json:"nomor_ktp"`
	KtpFile        *string    `gorm:"type:varchar(255)" json:"ktp_file"`
	PassportNumber *string    `gorm:"type:varchar(30)" json:"passport_number"`
	PassportExpiry *time.Time `json:"passport_expiry"`
	PassportFile   *string    `gorm:"type:varchar(255)" json:"passport_file"`
	BlazerSize     *string    `gorm:"type:varchar(10)" json:"blazer_size"`
	// NomorMeja (table number) is assigned by an admin (manual entry or
	// Excel import) for event seating — deliberately not part of
	// SelfProfileInput/selfProfileRequest, so a participant can't set their
	// own table number via the website's self-service profile form.
	NomorMeja *string `gorm:"type:varchar(20)" json:"nomor_meja"`
	// Description is an admin-only free-text note (manual entry or Excel
	// import) — same treatment as NomorMeja: deliberately absent from
	// SelfProfileInput/selfProfileRequest so it never surfaces on the
	// website's self-service form, only in the admin panel.
	Description *string `gorm:"type:text" json:"description"`
	// AttendanceStatus consolidates the attendance answer and form/shirt
	// completeness into a single field (see modules/peserta status
	// constants) rather than deriving a display status from
	// attendance-answer + completeness-of-several-other-fields on every
	// read. nil for non-Peserta users; nil is also treated as
	// "belum_konfirmasi" by readers, matching a freshly seeded/created
	// peserta that hasn't logged into the website yet.
	AttendanceStatus *string `gorm:"type:varchar(30)" json:"attendance_status"`
	// DataConsentAt records when the peserta agreed to the website's
	// Persetujuan Data Pribadi popup — nil means they haven't agreed yet, so
	// Landing.jsx shows the popup; set once via peserta.Service.AgreeDataConsent
	// and never cleared automatically, so the popup only ever appears once.
	DataConsentAt *time.Time `json:"data_consent_at"`
	// PassportSingleName means this peserta's passport has no middle/last
	// name (some Indonesian passports are issued with only one name) — set
	// via the website onboarding form's "nama saya di paspor hanya terdiri 1
	// nama" checkbox. isFormComplete (modules/peserta/service.go) consults
	// this so LastName isn't wrongly required for a peserta who genuinely
	// doesn't have one.
	PassportSingleName bool `gorm:"default:false" json:"passport_single_name"`
}

func (User) TableName() string { return "users" }

type Session struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"-"`
	UUID      string    `gorm:"type:uuid;uniqueIndex" json:"uuid"`
	UserID    uint64    `json:"-"`
	Token     string    `gorm:"type:varchar(255);index" json:"-"`
	IPAddress string    `gorm:"type:varchar(64)" json:"ip_address"`
	UserAgent string    `gorm:"type:text" json:"user_agent"`
	ExpiredAt time.Time `json:"expired_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (Session) TableName() string { return "sessions" }

func (s *Session) BeforeCreate(tx *gorm.DB) error {
	if s.UUID == "" {
		s.UUID = uuid.New().String()
	}
	return nil
}

type PasswordResetToken struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement"`
	Email     string `gorm:"type:varchar(255);index"`
	Token     string `gorm:"type:varchar(255);uniqueIndex"`
	ExpiredAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

func (PasswordResetToken) TableName() string { return "password_reset_tokens" }

type PasswordHistory struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement"`
	UserID    uint64 `gorm:"index"`
	Password  string `gorm:"type:varchar(255)"`
	CreatedAt time.Time
}

func (PasswordHistory) TableName() string { return "password_histories" }
