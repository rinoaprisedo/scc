package migrations

import (
	"log"

	"baseadmin/backend/modules/activity_logs"
	"baseadmin/backend/modules/menu_sections"
	"baseadmin/backend/modules/menus"
	"baseadmin/backend/modules/permissions"
	"baseadmin/backend/modules/roles"
	"baseadmin/backend/modules/settings"
	"baseadmin/backend/modules/users"

	"gorm.io/gorm"
)

// Run applies GORM AutoMigrate across every model in the project.
func Run(db *gorm.DB) {
	err := db.AutoMigrate(
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
	)
	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	applyPartialUniqueIndexes(db)

	log.Println("migrations applied successfully")
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
	}
	for _, stmt := range statements {
		if err := db.Exec(stmt).Error; err != nil {
			log.Fatalf("failed to apply partial unique index (%s): %v", stmt, err)
		}
	}
}
