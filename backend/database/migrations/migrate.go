package migrations

import (
	"log"

	"baseadmin/backend/modules/activity_logs"
	"baseadmin/backend/modules/bandara"
	"baseadmin/backend/modules/blazer_sizes"
	"baseadmin/backend/modules/kota_asal"
	"baseadmin/backend/modules/menu_sections"
	"baseadmin/backend/modules/menus"
	"baseadmin/backend/modules/permissions"
	"baseadmin/backend/modules/qr_gate"
	"baseadmin/backend/modules/qris_cross_border"
	"baseadmin/backend/modules/roles"
	"baseadmin/backend/modules/settings"
	"baseadmin/backend/modules/users"

	"gorm.io/gorm"
)

// Run applies GORM AutoMigrate across every model in the project.
func Run(db *gorm.DB) {
	err := db.AutoMigrate(
		&kota_asal.KotaAsal{},
		&bandara.Bandara{},
		&blazer_sizes.BlazerSize{},
		&users.User{},
		&roles.Role{},
		&users.Session{},
		&users.PasswordResetToken{},
		&users.PasswordHistory{},
		&menu_sections.MenuSection{},
		&menus.Menu{},
		&permissions.Permission{},
		&settings.Setting{},
		&activity_logs.ActivityLog{},
		&qr_gate.QrGate{},
		&qr_gate.QrGateScan{},
		&qris_cross_border.QrisCrossBorder{},
	)
	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	applyPartialUniqueIndexes(db)
	dropLegacyColumns(db)

	log.Println("migrations applied successfully")
}

// dropLegacyColumns removes columns superseded by a later, consolidated
// field — kept separate from AutoMigrate (which only ever adds/alters,
// never drops, for safety) and from applyPartialUniqueIndexes (a different
// concern). Idempotent: IF EXISTS makes every statement safe to re-run.
func dropLegacyColumns(db *gorm.DB) {
	statements := []string{
		// Replaced by users.attendance_status, which folds the attendance
		// answer and form/shirt completeness into one field instead of two
		// separate things a reader had to combine themselves.
		`ALTER TABLE users DROP COLUMN IF EXISTS attendance_confirmed`,
	}
	for _, stmt := range statements {
		if err := db.Exec(stmt).Error; err != nil {
			log.Fatalf("failed to drop legacy column (%s): %v", stmt, err)
		}
	}
}

// applyPartialUniqueIndexes enforces uniqueness only among non-deleted rows.
// GORM's `uniqueIndex` tag creates a plain unique index that also covers
// soft-deleted rows, which would permanently block re-using an email/name
// after the row holding it is deleted. Idempotent: safe to run on every
// migrate.
func applyPartialUniqueIndexes(db *gorm.DB) {
	statements := []string{
		`DROP INDEX IF EXISTS idx_users_email`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_active ON users (email) WHERE deleted_at IS NULL`,
		`DROP INDEX IF EXISTS idx_roles_name`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_name_active ON roles (name) WHERE deleted_at IS NULL`,
		// ktp_number (NIK) will be the peserta login identifier once that
		// flow ships — enforce uniqueness now. Postgres partial unique
		// indexes already treat NULL as distinct-from-NULL, so non-peserta
		// users (whose ktp_number is always NULL) never collide here.
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_ktp_number_active ON users (ktp_number) WHERE deleted_at IS NULL`,
		// The scanned code must resolve to exactly one active gate.
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_qr_gates_code_active ON qr_gates (code) WHERE deleted_at IS NULL`,
		// Second line of defense against duplicate QRIS submissions, behind
		// the application-level check in qris_cross_border's OCR job.
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_qris_cross_borders_reference_number_active ON qris_cross_borders (reference_number) WHERE deleted_at IS NULL AND reference_number IS NOT NULL`,
		// Two rows for the same size would split its stock count across
		// records, silently under-reporting availability.
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_blazer_sizes_size_active ON blazer_sizes (size) WHERE deleted_at IS NULL`,
	}
	for _, stmt := range statements {
		if err := db.Exec(stmt).Error; err != nil {
			log.Fatalf("failed to apply partial unique index (%s): %v", stmt, err)
		}
	}
}
