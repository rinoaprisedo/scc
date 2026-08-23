package seeders

import (
	"log"

	"baseadmin/backend/modules/bandara"
	"baseadmin/backend/modules/kota_asal"
	"baseadmin/backend/modules/menu_sections"
	"baseadmin/backend/modules/menus"
	"baseadmin/backend/modules/permissions"
	"baseadmin/backend/modules/roles"
	"baseadmin/backend/modules/settings"
	"baseadmin/backend/modules/users"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Run seeds the superadmin role/user, default menu structure, permissions,
// and default settings. Safe to re-run (idempotent via FirstOrCreate).
func Run(db *gorm.DB) {
	superadminRole := seedSuperadminRole(db)
	seedSuperadminUser(db, superadminRole)
	seedUser(db, superadminRole, "Rino Aprisedo", "rinoaprisedo@gmail.com", "password!!")
	seedPesertaRole(db)
	menuIDs := seedMenus(db)
	seedPermissions(db, superadminRole.ID, menuIDs)
	seedSettings(db)
	seedKotaAsal(db)
	seedBandara(db)
	log.Println("seed completed successfully")
}

func seedSuperadminRole(db *gorm.DB) roles.Role {
	var role roles.Role
	// .Attrs (not a positional FirstOrCreate arg) so these fields only apply
	// on create and never leak into the lookup's WHERE clause — GORM merges
	// non-zero struct fields passed positionally into the query too, which
	// would silently create a duplicate row if a seeded value ever changes.
	db.Where("name = ?", "Superadmin").Attrs(roles.Role{
		Name:         "Superadmin",
		Description:  "Full system access. Cannot be deleted or demoted.",
		IsSuperadmin: true,
	}).FirstOrCreate(&role)
	return role
}

// seedPesertaRole creates the role every row managed by the peserta module
// is scoped to (see backend/modules/peserta). Not granted any admin-panel
// menu permissions — participants aren't expected to log into the admin
// panel themselves.
func seedPesertaRole(db *gorm.DB) roles.Role {
	var role roles.Role
	db.Where("name = ?", "Peserta").Attrs(roles.Role{
		Name:        "Peserta",
		Description: "Event participant.",
	}).FirstOrCreate(&role)
	return role
}

func seedSuperadminUser(db *gorm.DB, role roles.Role) {
	var user users.User
	result := db.Where("email = ?", "admin@example.com").First(&user)
	if result.Error == nil {
		return
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte("Admin@12345"), 12)
	if err != nil {
		log.Fatalf("failed to hash superadmin password: %v", err)
	}
	user = users.User{
		Name:     "Super Admin",
		Email:    "admin@example.com",
		Password: string(hashed),
		Status:   users.StatusActive,
		RoleID:   &role.ID,
	}
	if err := db.Create(&user).Error; err != nil {
		log.Fatalf("failed to seed superadmin user: %v", err)
	}
}

func seedUser(db *gorm.DB, role roles.Role, name, email, password string) {
	var user users.User
	result := db.Where("email = ?", email).First(&user)
	if result.Error == nil {
		return
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		log.Fatalf("failed to hash user password: %v", err)
	}
	user = users.User{
		Name:     name,
		Email:    email,
		Password: string(hashed),
		Status:   users.StatusActive,
		RoleID:   &role.ID,
	}
	if err := db.Create(&user).Error; err != nil {
		log.Fatalf("failed to seed user: %v", err)
	}
}

type menuDef struct {
	Name string
	Icon string
	Path string
}

func seedMenus(db *gorm.DB) []uint64 {
	sections := []struct {
		Section menu_sections.MenuSection
		Menus   []menuDef
	}{
		{
			Section: menu_sections.MenuSection{Name: "Dashboard", Icon: "LayoutDashboard", Order: 1},
			Menus: []menuDef{
				{Name: "Dashboard", Icon: "LayoutDashboard", Path: "/"},
			},
		},
		{
			Section: menu_sections.MenuSection{Name: "User Management", Icon: "Users", Order: 2},
			Menus: []menuDef{
				{Name: "Users", Icon: "User", Path: "users"},
				{Name: "Roles", Icon: "Shield", Path: "roles"},
				{Name: "Peserta", Icon: "UserCheck", Path: "peserta"},
			},
		},
		{
			Section: menu_sections.MenuSection{Name: "Master Data", Icon: "Database", Order: 3},
			Menus: []menuDef{
				{Name: "Kota Asal", Icon: "MapPin", Path: "kota-asal"},
				{Name: "Bandara", Icon: "Plane", Path: "bandara"},
				{Name: "QR Gate", Icon: "QrCode", Path: "qr-gate"},
				{Name: "Qris Cross Border", Icon: "Landmark", Path: "qris-cross-border"},
			},
		},
		{
			Section: menu_sections.MenuSection{Name: "System", Icon: "Settings", Order: 4},
			Menus: []menuDef{
				{Name: "Menus", Icon: "Menu", Path: "menus"},
				{Name: "Menu Sections", Icon: "FolderTree", Path: "menu-sections"},
				{Name: "Settings", Icon: "Settings", Path: "settings"},
				{Name: "Activity Logs", Icon: "History", Path: "activity-logs"},
			},
		},
	}

	var menuIDs []uint64
	for _, s := range sections {
		var section menu_sections.MenuSection
		// See seedSuperadminRole for why .Attrs is used instead of a
		// positional FirstOrCreate arg: Order/Icon are non-zero and would
		// otherwise be merged into the lookup WHERE, so re-running the
		// seeder after reordering sections would create a duplicate row
		// instead of finding the existing one.
		db.Where("name = ?", s.Section.Name).Attrs(s.Section).FirstOrCreate(&section)

		for i, m := range s.Menus {
			var menu menus.Menu
			db.Where("path = ?", m.Path).Attrs(menus.Menu{
				MenuSectionID: section.ID,
				Name:          m.Name,
				Icon:          m.Icon,
				Path:          m.Path,
				Order:         i + 1,
				IsActive:      true,
			}).FirstOrCreate(&menu)
			menuIDs = append(menuIDs, menu.ID)
		}
	}
	return menuIDs
}

func seedPermissions(db *gorm.DB, roleID uint64, menuIDs []uint64) {
	for _, menuID := range menuIDs {
		var perm permissions.Permission
		db.Where("role_id = ? AND menu_id = ?", roleID, menuID).FirstOrCreate(&perm, permissions.Permission{
			RoleID:    roleID,
			MenuID:    menuID,
			CanView:   true,
			CanCreate: true,
			CanEdit:   true,
			CanDelete: true,
		})
	}
}

func seedSettings(db *gorm.DB) {
	defaults := map[string]settings.SettingType{
		"app_name":                       settings.TypeString,
		"app_logo":                       settings.TypeFile,
		"app_favicon":                    settings.TypeFile,
		"maintenance_mode":               settings.TypeBoolean,
		"menu_event_agenda_enabled":      settings.TypeBoolean,
		"menu_event_gallery_enabled":     settings.TypeBoolean,
		"menu_dress_code_enabled":        settings.TypeBoolean,
		"menu_qris_cross_border_enabled": settings.TypeBoolean,
		"menu_about_malaysia_enabled":    settings.TypeBoolean,
		"menu_scanner_qr_enabled":        settings.TypeBoolean,
		"menu_history_scanner_enabled":   settings.TypeBoolean,
		"menu_event_information_enabled": settings.TypeBoolean,
	}
	values := map[string]string{
		"app_name":                       "scc",
		"app_logo":                       "",
		"app_favicon":                    "",
		"maintenance_mode":               "false",
		"menu_event_agenda_enabled":      "true",
		"menu_event_gallery_enabled":     "true",
		"menu_dress_code_enabled":        "true",
		"menu_qris_cross_border_enabled": "true",
		"menu_about_malaysia_enabled":    "true",
		"menu_scanner_qr_enabled":        "true",
		"menu_history_scanner_enabled":   "true",
		"menu_event_information_enabled": "true",
	}
	for key, t := range defaults {
		var s settings.Setting
		db.Where("key = ?", key).FirstOrCreate(&s, settings.Setting{Key: key, Value: values[key], Type: t})
	}
}

// seedKotaAsal seeds the Kota Asal ("origin city") master data every
// provincial capital of Indonesia (all 38 provinces) plus a handful of other
// major metro areas commonly used as an event's departure city — this is
// the option list the peserta Excel import's Kota Asal dropdown is built
// from (see modules/peserta/import.go), so a thin seed list there would
// force most imports to fail Kota Asal validation with nothing to match.
func seedKotaAsal(db *gorm.DB) {
	cities := []kota_asal.KotaAsal{
		{Name: "Jakarta", Province: "DKI Jakarta"},
		{Name: "Bandung", Province: "Jawa Barat"},
		{Name: "Semarang", Province: "Jawa Tengah"},
		{Name: "Surabaya", Province: "Jawa Timur"},
		{Name: "Yogyakarta", Province: "DI Yogyakarta"},
		{Name: "Serang", Province: "Banten"},
		{Name: "Denpasar", Province: "Bali"},
		{Name: "Mataram", Province: "Nusa Tenggara Barat"},
		{Name: "Kupang", Province: "Nusa Tenggara Timur"},
		{Name: "Pontianak", Province: "Kalimantan Barat"},
		{Name: "Palangka Raya", Province: "Kalimantan Tengah"},
		{Name: "Banjarmasin", Province: "Kalimantan Selatan"},
		{Name: "Samarinda", Province: "Kalimantan Timur"},
		{Name: "Tanjung Selor", Province: "Kalimantan Utara"},
		{Name: "Manado", Province: "Sulawesi Utara"},
		{Name: "Palu", Province: "Sulawesi Tengah"},
		{Name: "Makassar", Province: "Sulawesi Selatan"},
		{Name: "Kendari", Province: "Sulawesi Tenggara"},
		{Name: "Gorontalo", Province: "Gorontalo"},
		{Name: "Mamuju", Province: "Sulawesi Barat"},
		{Name: "Ambon", Province: "Maluku"},
		{Name: "Sofifi", Province: "Maluku Utara"},
		{Name: "Jayapura", Province: "Papua"},
		{Name: "Manokwari", Province: "Papua Barat"},
		{Name: "Sorong", Province: "Papua Barat Daya"},
		{Name: "Nabire", Province: "Papua Tengah"},
		{Name: "Wamena", Province: "Papua Pegunungan"},
		{Name: "Merauke", Province: "Papua Selatan"},
		{Name: "Banda Aceh", Province: "Aceh"},
		{Name: "Medan", Province: "Sumatera Utara"},
		{Name: "Padang", Province: "Sumatera Barat"},
		{Name: "Pekanbaru", Province: "Riau"},
		{Name: "Tanjung Pinang", Province: "Kepulauan Riau"},
		{Name: "Jambi", Province: "Jambi"},
		{Name: "Palembang", Province: "Sumatera Selatan"},
		{Name: "Pangkal Pinang", Province: "Kepulauan Bangka Belitung"},
		{Name: "Bengkulu", Province: "Bengkulu"},
		{Name: "Bandar Lampung", Province: "Lampung"},
		// Other major metro areas beyond the provincial capitals above.
		{Name: "Surakarta", Province: "Jawa Tengah"},
		{Name: "Malang", Province: "Jawa Timur"},
		{Name: "Bogor", Province: "Jawa Barat"},
		{Name: "Bekasi", Province: "Jawa Barat"},
		{Name: "Tangerang", Province: "Banten"},
		{Name: "Depok", Province: "Jawa Barat"},
		{Name: "Batam", Province: "Kepulauan Riau"},
		{Name: "Balikpapan", Province: "Kalimantan Timur"},
		{Name: "Cirebon", Province: "Jawa Barat"},
	}
	for _, c := range cities {
		var existing kota_asal.KotaAsal
		db.Where("name = ?", c.Name).Attrs(c).FirstOrCreate(&existing)
	}
}

// seedBandara seeds the Bandara Terdekat ("nearest airport") master data
// with major commercial airports across Indonesia, roughly one per metro
// area seeded in seedKotaAsal — same reasoning as there: this backs the
// Excel import's Bandara Terdekat dropdown.
func seedBandara(db *gorm.DB) {
	airports := []string{
		"Soekarno-Hatta International Airport",
		"Halim Perdanakusuma Airport",
		"Husein Sastranegara International Airport",
		"Kertajati International Airport",
		"Achmad Yani International Airport",
		"Adisumarmo International Airport",
		"Adisutjipto International Airport",
		"Yogyakarta International Airport",
		"Juanda International Airport",
		"Abdul Rachman Saleh Airport",
		"Ngurah Rai International Airport",
		"Zainuddin Abdul Madjid International Airport",
		"El Tari Airport",
		"Supadio International Airport",
		"Tjilik Riwut Airport",
		"Syamsudin Noor International Airport",
		"Sultan Aji Muhammad Sulaiman Sepinggan International Airport",
		"Aji Pangeran Tumenggung Pranoto International Airport",
		"Tanjung Harapan Airport",
		"Sam Ratulangi International Airport",
		"Mutiara SIS Al-Jufrie Airport",
		"Sultan Hasanuddin International Airport",
		"Haluoleo Airport",
		"Djalaluddin Airport",
		"Tampa Padang Airport",
		"Pattimura Airport",
		"Sultan Babullah Airport",
		"Sentani International Airport",
		"Rendani Airport",
		"Domine Eduard Osok Airport",
		"Douw Aturure Airport",
		"Wamena Airport",
		"Mopah International Airport",
		"Sultan Iskandar Muda International Airport",
		"Kualanamu International Airport",
		"Minangkabau International Airport",
		"Sultan Syarif Kasim II International Airport",
		"Raja Haji Fisabilillah Airport",
		"Hang Nadim International Airport",
		"Sultan Thaha Airport",
		"Sultan Mahmud Badaruddin II International Airport",
		"Depati Amir Airport",
		"Fatmawati Soekarno Airport",
		"Radin Inten II Airport",
	}
	for _, name := range airports {
		var existing bandara.Bandara
		db.Where("name = ?", name).Attrs(bandara.Bandara{Name: name}).FirstOrCreate(&existing)
	}
}
