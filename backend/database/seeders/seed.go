package seeders

import (
	"log"

	"baseadmin/backend/modules/bandara"
	"baseadmin/backend/modules/blazer_sizes"
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
	seedBlazerSizes(db)
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
				{Name: "Blazer Size", Icon: "Shirt", Path: "blazer-sizes"},
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

// seedKotaAsal seeds the Kota Asal ("origin city") master data with every
// kabupaten/kota in Indonesia (38 provinces, 514 regencies as of Kepmendagri
// No 300.2.2-2138 Tahun 2025) — this is the option list the peserta Excel
// import's Kota Asal dropdown is built from (see modules/peserta/import.go),
// so a thin seed list there would force most imports to fail Kota Asal
// validation with nothing to match.
func seedKotaAsal(db *gorm.DB) {
	cities := []kota_asal.KotaAsal{
		{Name: "Kabupaten Aceh Selatan", Province: "Aceh"},
		{Name: "Kabupaten Aceh Tenggara", Province: "Aceh"},
		{Name: "Kabupaten Aceh Timur", Province: "Aceh"},
		{Name: "Kabupaten Aceh Tengah", Province: "Aceh"},
		{Name: "Kabupaten Aceh Barat", Province: "Aceh"},
		{Name: "Kabupaten Aceh Besar", Province: "Aceh"},
		{Name: "Kabupaten Pidie", Province: "Aceh"},
		{Name: "Kabupaten Aceh Utara", Province: "Aceh"},
		{Name: "Kabupaten Simeulue", Province: "Aceh"},
		{Name: "Kabupaten Aceh Singkil", Province: "Aceh"},
		{Name: "Kabupaten Bireuen", Province: "Aceh"},
		{Name: "Kabupaten Aceh Barat Daya", Province: "Aceh"},
		{Name: "Kabupaten Gayo Lues", Province: "Aceh"},
		{Name: "Kabupaten Aceh Jaya", Province: "Aceh"},
		{Name: "Kabupaten Nagan Raya", Province: "Aceh"},
		{Name: "Kabupaten Aceh Tamiang", Province: "Aceh"},
		{Name: "Kabupaten Bener Meriah", Province: "Aceh"},
		{Name: "Kabupaten Pidie Jaya", Province: "Aceh"},
		{Name: "Kota Banda Aceh", Province: "Aceh"},
		{Name: "Kota Sabang", Province: "Aceh"},
		{Name: "Kota Lhokseumawe", Province: "Aceh"},
		{Name: "Kota Langsa", Province: "Aceh"},
		{Name: "Kota Subulussalam", Province: "Aceh"},
		{Name: "Kabupaten Tapanuli Tengah", Province: "Sumatera Utara"},
		{Name: "Kabupaten Tapanuli Utara", Province: "Sumatera Utara"},
		{Name: "Kabupaten Tapanuli Selatan", Province: "Sumatera Utara"},
		{Name: "Kabupaten Nias", Province: "Sumatera Utara"},
		{Name: "Kabupaten Langkat", Province: "Sumatera Utara"},
		{Name: "Kabupaten Karo", Province: "Sumatera Utara"},
		{Name: "Kabupaten Deli Serdang", Province: "Sumatera Utara"},
		{Name: "Kabupaten Simalungun", Province: "Sumatera Utara"},
		{Name: "Kabupaten Asahan", Province: "Sumatera Utara"},
		{Name: "Kabupaten Labuhanbatu", Province: "Sumatera Utara"},
		{Name: "Kabupaten Dairi", Province: "Sumatera Utara"},
		{Name: "Kabupaten Toba", Province: "Sumatera Utara"},
		{Name: "Kabupaten Mandailing Natal", Province: "Sumatera Utara"},
		{Name: "Kabupaten Nias Selatan", Province: "Sumatera Utara"},
		{Name: "Kabupaten Pakpak Bharat", Province: "Sumatera Utara"},
		{Name: "Kabupaten Humbang Hasundutan", Province: "Sumatera Utara"},
		{Name: "Kabupaten Samosir", Province: "Sumatera Utara"},
		{Name: "Kabupaten Serdang Bedagai", Province: "Sumatera Utara"},
		{Name: "Kabupaten Batu Bara", Province: "Sumatera Utara"},
		{Name: "Kabupaten Padang Lawas Utara", Province: "Sumatera Utara"},
		{Name: "Kabupaten Padang Lawas", Province: "Sumatera Utara"},
		{Name: "Kabupaten Labuhanbatu Selatan", Province: "Sumatera Utara"},
		{Name: "Kabupaten Labuhanbatu Utara", Province: "Sumatera Utara"},
		{Name: "Kabupaten Nias Utara", Province: "Sumatera Utara"},
		{Name: "Kabupaten Nias Barat", Province: "Sumatera Utara"},
		{Name: "Kota Medan", Province: "Sumatera Utara"},
		{Name: "Kota Pematangsiantar", Province: "Sumatera Utara"},
		{Name: "Kota Sibolga", Province: "Sumatera Utara"},
		{Name: "Kota Tanjungbalai", Province: "Sumatera Utara"},
		{Name: "Kota Binjai", Province: "Sumatera Utara"},
		{Name: "Kota Tebing Tinggi", Province: "Sumatera Utara"},
		{Name: "Kota Padangsidimpuan", Province: "Sumatera Utara"},
		{Name: "Kota Gunungsitoli", Province: "Sumatera Utara"},
		{Name: "Kabupaten Pesisir Selatan", Province: "Sumatera Barat"},
		{Name: "Kabupaten Solok", Province: "Sumatera Barat"},
		{Name: "Kabupaten Sijunjung", Province: "Sumatera Barat"},
		{Name: "Kabupaten Tanah Datar", Province: "Sumatera Barat"},
		{Name: "Kabupaten Padang Pariaman", Province: "Sumatera Barat"},
		{Name: "Kabupaten Agam", Province: "Sumatera Barat"},
		{Name: "Kabupaten Lima Puluh Kota", Province: "Sumatera Barat"},
		{Name: "Kabupaten Pasaman", Province: "Sumatera Barat"},
		{Name: "Kabupaten Kepulauan Mentawai", Province: "Sumatera Barat"},
		{Name: "Kabupaten Dharmasraya", Province: "Sumatera Barat"},
		{Name: "Kabupaten Solok Selatan", Province: "Sumatera Barat"},
		{Name: "Kabupaten Pasaman Barat", Province: "Sumatera Barat"},
		{Name: "Kota Padang", Province: "Sumatera Barat"},
		{Name: "Kota Solok", Province: "Sumatera Barat"},
		{Name: "Kota Sawahlunto", Province: "Sumatera Barat"},
		{Name: "Kota Padang Panjang", Province: "Sumatera Barat"},
		{Name: "Kota Bukittinggi", Province: "Sumatera Barat"},
		{Name: "Kota Payakumbuh", Province: "Sumatera Barat"},
		{Name: "Kota Pariaman", Province: "Sumatera Barat"},
		{Name: "Kabupaten Kampar", Province: "Riau"},
		{Name: "Kabupaten Indragiri Hulu", Province: "Riau"},
		{Name: "Kabupaten Bengkalis", Province: "Riau"},
		{Name: "Kabupaten Indragiri Hilir", Province: "Riau"},
		{Name: "Kabupaten Pelalawan", Province: "Riau"},
		{Name: "Kabupaten Rokan Hulu", Province: "Riau"},
		{Name: "Kabupaten Rokan Hilir", Province: "Riau"},
		{Name: "Kabupaten Siak", Province: "Riau"},
		{Name: "Kabupaten Kuantan Singingi", Province: "Riau"},
		{Name: "Kabupaten Kepulauan Meranti", Province: "Riau"},
		{Name: "Kota Pekanbaru", Province: "Riau"},
		{Name: "Kota Dumai", Province: "Riau"},
		{Name: "Kabupaten Kerinci", Province: "Jambi"},
		{Name: "Kabupaten Merangin", Province: "Jambi"},
		{Name: "Kabupaten Sarolangun", Province: "Jambi"},
		{Name: "Kabupaten Batanghari", Province: "Jambi"},
		{Name: "Kabupaten Muaro Jambi", Province: "Jambi"},
		{Name: "Kabupaten Tanjung Jabung Barat", Province: "Jambi"},
		{Name: "Kabupaten Tanjung Jabung Timur", Province: "Jambi"},
		{Name: "Kabupaten Bungo", Province: "Jambi"},
		{Name: "Kabupaten Tebo", Province: "Jambi"},
		{Name: "Kota Jambi", Province: "Jambi"},
		{Name: "Kota Sungai Penuh", Province: "Jambi"},
		{Name: "Kabupaten Ogan Komering Ulu", Province: "Sumatera Selatan"},
		{Name: "Kabupaten Ogan Komering Ilir", Province: "Sumatera Selatan"},
		{Name: "Kabupaten Muara Enim", Province: "Sumatera Selatan"},
		{Name: "Kabupaten Lahat", Province: "Sumatera Selatan"},
		{Name: "Kabupaten Musi Rawas", Province: "Sumatera Selatan"},
		{Name: "Kabupaten Musi Banyuasin", Province: "Sumatera Selatan"},
		{Name: "Kabupaten Banyuasin", Province: "Sumatera Selatan"},
		{Name: "Kabupaten Ogan Komering Ulu Timur", Province: "Sumatera Selatan"},
		{Name: "Kabupaten Ogan Komering Ulu Selatan", Province: "Sumatera Selatan"},
		{Name: "Kabupaten Ogan Ilir", Province: "Sumatera Selatan"},
		{Name: "Kabupaten Empat Lawang", Province: "Sumatera Selatan"},
		{Name: "Kabupaten Penukal Abab Lematang Ilir", Province: "Sumatera Selatan"},
		{Name: "Kabupaten Musi Rawas Utara", Province: "Sumatera Selatan"},
		{Name: "Kota Palembang", Province: "Sumatera Selatan"},
		{Name: "Kota Pagar Alam", Province: "Sumatera Selatan"},
		{Name: "Kota Lubuk Linggau", Province: "Sumatera Selatan"},
		{Name: "Kota Prabumulih", Province: "Sumatera Selatan"},
		{Name: "Kabupaten Bengkulu Selatan", Province: "Bengkulu"},
		{Name: "Kabupaten Rejang Lebong", Province: "Bengkulu"},
		{Name: "Kabupaten Bengkulu Utara", Province: "Bengkulu"},
		{Name: "Kabupaten Kaur", Province: "Bengkulu"},
		{Name: "Kabupaten Seluma", Province: "Bengkulu"},
		{Name: "Kabupaten Mukomuko", Province: "Bengkulu"},
		{Name: "Kabupaten Lebong", Province: "Bengkulu"},
		{Name: "Kabupaten Kepahiang", Province: "Bengkulu"},
		{Name: "Kabupaten Bengkulu Tengah", Province: "Bengkulu"},
		{Name: "Kota Bengkulu", Province: "Bengkulu"},
		{Name: "Kabupaten Lampung Selatan", Province: "Lampung"},
		{Name: "Kabupaten Lampung Tengah", Province: "Lampung"},
		{Name: "Kabupaten Lampung Utara", Province: "Lampung"},
		{Name: "Kabupaten Lampung Barat", Province: "Lampung"},
		{Name: "Kabupaten Tulang Bawang", Province: "Lampung"},
		{Name: "Kabupaten Tanggamus", Province: "Lampung"},
		{Name: "Kabupaten Lampung Timur", Province: "Lampung"},
		{Name: "Kabupaten Way Kanan", Province: "Lampung"},
		{Name: "Kabupaten Pesawaran", Province: "Lampung"},
		{Name: "Kabupaten Pringsewu", Province: "Lampung"},
		{Name: "Kabupaten Mesuji", Province: "Lampung"},
		{Name: "Kabupaten Tulang Bawang Barat", Province: "Lampung"},
		{Name: "Kabupaten Pesisir Barat", Province: "Lampung"},
		{Name: "Kota Bandar Lampung", Province: "Lampung"},
		{Name: "Kota Metro", Province: "Lampung"},
		{Name: "Kabupaten Bangka", Province: "Kepulauan Bangka Belitung"},
		{Name: "Kabupaten Belitung", Province: "Kepulauan Bangka Belitung"},
		{Name: "Kabupaten Bangka Selatan", Province: "Kepulauan Bangka Belitung"},
		{Name: "Kabupaten Bangka Tengah", Province: "Kepulauan Bangka Belitung"},
		{Name: "Kabupaten Bangka Barat", Province: "Kepulauan Bangka Belitung"},
		{Name: "Kabupaten Belitung Timur", Province: "Kepulauan Bangka Belitung"},
		{Name: "Kota Pangkal Pinang", Province: "Kepulauan Bangka Belitung"},
		{Name: "Kabupaten Bintan", Province: "Kepulauan Riau"},
		{Name: "Kabupaten Karimun", Province: "Kepulauan Riau"},
		{Name: "Kabupaten Natuna", Province: "Kepulauan Riau"},
		{Name: "Kabupaten Lingga", Province: "Kepulauan Riau"},
		{Name: "Kabupaten Kepulauan Anambas", Province: "Kepulauan Riau"},
		{Name: "Kota Batam", Province: "Kepulauan Riau"},
		{Name: "Kota Tanjung Pinang", Province: "Kepulauan Riau"},
		{Name: "Kabupaten Administrasi Kepulauan Seribu", Province: "DKI Jakarta"},
		{Name: "Kota Administrasi Jakarta Pusat", Province: "DKI Jakarta"},
		{Name: "Kota Administrasi Jakarta Utara", Province: "DKI Jakarta"},
		{Name: "Kota Administrasi Jakarta Barat", Province: "DKI Jakarta"},
		{Name: "Kota Administrasi Jakarta Selatan", Province: "DKI Jakarta"},
		{Name: "Kota Administrasi Jakarta Timur", Province: "DKI Jakarta"},
		{Name: "Kabupaten Bogor", Province: "Jawa Barat"},
		{Name: "Kabupaten Sukabumi", Province: "Jawa Barat"},
		{Name: "Kabupaten Cianjur", Province: "Jawa Barat"},
		{Name: "Kabupaten Bandung", Province: "Jawa Barat"},
		{Name: "Kabupaten Garut", Province: "Jawa Barat"},
		{Name: "Kabupaten Tasikmalaya", Province: "Jawa Barat"},
		{Name: "Kabupaten Ciamis", Province: "Jawa Barat"},
		{Name: "Kabupaten Kuningan", Province: "Jawa Barat"},
		{Name: "Kabupaten Cirebon", Province: "Jawa Barat"},
		{Name: "Kabupaten Majalengka", Province: "Jawa Barat"},
		{Name: "Kabupaten Sumedang", Province: "Jawa Barat"},
		{Name: "Kabupaten Indramayu", Province: "Jawa Barat"},
		{Name: "Kabupaten Subang", Province: "Jawa Barat"},
		{Name: "Kabupaten Purwakarta", Province: "Jawa Barat"},
		{Name: "Kabupaten Karawang", Province: "Jawa Barat"},
		{Name: "Kabupaten Bekasi", Province: "Jawa Barat"},
		{Name: "Kabupaten Bandung Barat", Province: "Jawa Barat"},
		{Name: "Kabupaten Pangandaran", Province: "Jawa Barat"},
		{Name: "Kota Bogor", Province: "Jawa Barat"},
		{Name: "Kota Sukabumi", Province: "Jawa Barat"},
		{Name: "Kota Bandung", Province: "Jawa Barat"},
		{Name: "Kota Cirebon", Province: "Jawa Barat"},
		{Name: "Kota Bekasi", Province: "Jawa Barat"},
		{Name: "Kota Depok", Province: "Jawa Barat"},
		{Name: "Kota Cimahi", Province: "Jawa Barat"},
		{Name: "Kota Tasikmalaya", Province: "Jawa Barat"},
		{Name: "Kota Banjar", Province: "Jawa Barat"},
		{Name: "Kabupaten Cilacap", Province: "Jawa Tengah"},
		{Name: "Kabupaten Banyumas", Province: "Jawa Tengah"},
		{Name: "Kabupaten Purbalingga", Province: "Jawa Tengah"},
		{Name: "Kabupaten Banjarnegara", Province: "Jawa Tengah"},
		{Name: "Kabupaten Kebumen", Province: "Jawa Tengah"},
		{Name: "Kabupaten Purworejo", Province: "Jawa Tengah"},
		{Name: "Kabupaten Wonosobo", Province: "Jawa Tengah"},
		{Name: "Kabupaten Magelang", Province: "Jawa Tengah"},
		{Name: "Kabupaten Boyolali", Province: "Jawa Tengah"},
		{Name: "Kabupaten Klaten", Province: "Jawa Tengah"},
		{Name: "Kabupaten Sukoharjo", Province: "Jawa Tengah"},
		{Name: "Kabupaten Wonogiri", Province: "Jawa Tengah"},
		{Name: "Kabupaten Karanganyar", Province: "Jawa Tengah"},
		{Name: "Kabupaten Sragen", Province: "Jawa Tengah"},
		{Name: "Kabupaten Grobogan", Province: "Jawa Tengah"},
		{Name: "Kabupaten Blora", Province: "Jawa Tengah"},
		{Name: "Kabupaten Rembang", Province: "Jawa Tengah"},
		{Name: "Kabupaten Pati", Province: "Jawa Tengah"},
		{Name: "Kabupaten Kudus", Province: "Jawa Tengah"},
		{Name: "Kabupaten Jepara", Province: "Jawa Tengah"},
		{Name: "Kabupaten Demak", Province: "Jawa Tengah"},
		{Name: "Kabupaten Semarang", Province: "Jawa Tengah"},
		{Name: "Kabupaten Temanggung", Province: "Jawa Tengah"},
		{Name: "Kabupaten Kendal", Province: "Jawa Tengah"},
		{Name: "Kabupaten Batang", Province: "Jawa Tengah"},
		{Name: "Kabupaten Pekalongan", Province: "Jawa Tengah"},
		{Name: "Kabupaten Pemalang", Province: "Jawa Tengah"},
		{Name: "Kabupaten Tegal", Province: "Jawa Tengah"},
		{Name: "Kabupaten Brebes", Province: "Jawa Tengah"},
		{Name: "Kota Magelang", Province: "Jawa Tengah"},
		{Name: "Kota Surakarta", Province: "Jawa Tengah"},
		{Name: "Kota Salatiga", Province: "Jawa Tengah"},
		{Name: "Kota Semarang", Province: "Jawa Tengah"},
		{Name: "Kota Pekalongan", Province: "Jawa Tengah"},
		{Name: "Kota Tegal", Province: "Jawa Tengah"},
		{Name: "Kabupaten Kulon Progo", Province: "DI Yogyakarta"},
		{Name: "Kabupaten Bantul", Province: "DI Yogyakarta"},
		{Name: "Kabupaten Gunungkidul", Province: "DI Yogyakarta"},
		{Name: "Kabupaten Sleman", Province: "DI Yogyakarta"},
		{Name: "Kota Yogyakarta", Province: "DI Yogyakarta"},
		{Name: "Kabupaten Pacitan", Province: "Jawa Timur"},
		{Name: "Kabupaten Ponorogo", Province: "Jawa Timur"},
		{Name: "Kabupaten Trenggalek", Province: "Jawa Timur"},
		{Name: "Kabupaten Tulungagung", Province: "Jawa Timur"},
		{Name: "Kabupaten Blitar", Province: "Jawa Timur"},
		{Name: "Kabupaten Kediri", Province: "Jawa Timur"},
		{Name: "Kabupaten Malang", Province: "Jawa Timur"},
		{Name: "Kabupaten Lumajang", Province: "Jawa Timur"},
		{Name: "Kabupaten Jember", Province: "Jawa Timur"},
		{Name: "Kabupaten Banyuwangi", Province: "Jawa Timur"},
		{Name: "Kabupaten Bondowoso", Province: "Jawa Timur"},
		{Name: "Kabupaten Situbondo", Province: "Jawa Timur"},
		{Name: "Kabupaten Probolinggo", Province: "Jawa Timur"},
		{Name: "Kabupaten Pasuruan", Province: "Jawa Timur"},
		{Name: "Kabupaten Sidoarjo", Province: "Jawa Timur"},
		{Name: "Kabupaten Mojokerto", Province: "Jawa Timur"},
		{Name: "Kabupaten Jombang", Province: "Jawa Timur"},
		{Name: "Kabupaten Nganjuk", Province: "Jawa Timur"},
		{Name: "Kabupaten Madiun", Province: "Jawa Timur"},
		{Name: "Kabupaten Magetan", Province: "Jawa Timur"},
		{Name: "Kabupaten Ngawi", Province: "Jawa Timur"},
		{Name: "Kabupaten Bojonegoro", Province: "Jawa Timur"},
		{Name: "Kabupaten Tuban", Province: "Jawa Timur"},
		{Name: "Kabupaten Lamongan", Province: "Jawa Timur"},
		{Name: "Kabupaten Gresik", Province: "Jawa Timur"},
		{Name: "Kabupaten Bangkalan", Province: "Jawa Timur"},
		{Name: "Kabupaten Sampang", Province: "Jawa Timur"},
		{Name: "Kabupaten Pamekasan", Province: "Jawa Timur"},
		{Name: "Kabupaten Sumenep", Province: "Jawa Timur"},
		{Name: "Kota Kediri", Province: "Jawa Timur"},
		{Name: "Kota Blitar", Province: "Jawa Timur"},
		{Name: "Kota Malang", Province: "Jawa Timur"},
		{Name: "Kota Probolinggo", Province: "Jawa Timur"},
		{Name: "Kota Pasuruan", Province: "Jawa Timur"},
		{Name: "Kota Mojokerto", Province: "Jawa Timur"},
		{Name: "Kota Madiun", Province: "Jawa Timur"},
		{Name: "Kota Surabaya", Province: "Jawa Timur"},
		{Name: "Kota Batu", Province: "Jawa Timur"},
		{Name: "Kabupaten Pandeglang", Province: "Banten"},
		{Name: "Kabupaten Lebak", Province: "Banten"},
		{Name: "Kabupaten Tangerang", Province: "Banten"},
		{Name: "Kabupaten Serang", Province: "Banten"},
		{Name: "Kota Tangerang", Province: "Banten"},
		{Name: "Kota Cilegon", Province: "Banten"},
		{Name: "Kota Serang", Province: "Banten"},
		{Name: "Kota Tangerang Selatan", Province: "Banten"},
		{Name: "Kabupaten Jembrana", Province: "Bali"},
		{Name: "Kabupaten Tabanan", Province: "Bali"},
		{Name: "Kabupaten Badung", Province: "Bali"},
		{Name: "Kabupaten Gianyar", Province: "Bali"},
		{Name: "Kabupaten Klungkung", Province: "Bali"},
		{Name: "Kabupaten Bangli", Province: "Bali"},
		{Name: "Kabupaten Karangasem", Province: "Bali"},
		{Name: "Kabupaten Buleleng", Province: "Bali"},
		{Name: "Kota Denpasar", Province: "Bali"},
		{Name: "Kabupaten Lombok Barat", Province: "Nusa Tenggara Barat"},
		{Name: "Kabupaten Lombok Tengah", Province: "Nusa Tenggara Barat"},
		{Name: "Kabupaten Lombok Timur", Province: "Nusa Tenggara Barat"},
		{Name: "Kabupaten Sumbawa", Province: "Nusa Tenggara Barat"},
		{Name: "Kabupaten Dompu", Province: "Nusa Tenggara Barat"},
		{Name: "Kabupaten Bima", Province: "Nusa Tenggara Barat"},
		{Name: "Kabupaten Sumbawa Barat", Province: "Nusa Tenggara Barat"},
		{Name: "Kabupaten Lombok Utara", Province: "Nusa Tenggara Barat"},
		{Name: "Kota Mataram", Province: "Nusa Tenggara Barat"},
		{Name: "Kota Bima", Province: "Nusa Tenggara Barat"},
		{Name: "Kabupaten Kupang", Province: "Nusa Tenggara Timur"},
		{Name: "Kabupaten Timor Tengah Selatan", Province: "Nusa Tenggara Timur"},
		{Name: "Kabupaten Timor Tengah Utara", Province: "Nusa Tenggara Timur"},
		{Name: "Kabupaten Belu", Province: "Nusa Tenggara Timur"},
		{Name: "Kabupaten Alor", Province: "Nusa Tenggara Timur"},
		{Name: "Kabupaten Flores Timur", Province: "Nusa Tenggara Timur"},
		{Name: "Kabupaten Sikka", Province: "Nusa Tenggara Timur"},
		{Name: "Kabupaten Ende", Province: "Nusa Tenggara Timur"},
		{Name: "Kabupaten Ngada", Province: "Nusa Tenggara Timur"},
		{Name: "Kabupaten Manggarai", Province: "Nusa Tenggara Timur"},
		{Name: "Kabupaten Sumba Timur", Province: "Nusa Tenggara Timur"},
		{Name: "Kabupaten Sumba Barat", Province: "Nusa Tenggara Timur"},
		{Name: "Kabupaten Lembata", Province: "Nusa Tenggara Timur"},
		{Name: "Kabupaten Rote Ndao", Province: "Nusa Tenggara Timur"},
		{Name: "Kabupaten Manggarai Barat", Province: "Nusa Tenggara Timur"},
		{Name: "Kabupaten Nagekeo", Province: "Nusa Tenggara Timur"},
		{Name: "Kabupaten Sumba Tengah", Province: "Nusa Tenggara Timur"},
		{Name: "Kabupaten Sumba Barat Daya", Province: "Nusa Tenggara Timur"},
		{Name: "Kabupaten Manggarai Timur", Province: "Nusa Tenggara Timur"},
		{Name: "Kabupaten Sabu Raijua", Province: "Nusa Tenggara Timur"},
		{Name: "Kabupaten Malaka", Province: "Nusa Tenggara Timur"},
		{Name: "Kota Kupang", Province: "Nusa Tenggara Timur"},
		{Name: "Kabupaten Sambas", Province: "Kalimantan Barat"},
		{Name: "Kabupaten Mempawah", Province: "Kalimantan Barat"},
		{Name: "Kabupaten Sanggau", Province: "Kalimantan Barat"},
		{Name: "Kabupaten Ketapang", Province: "Kalimantan Barat"},
		{Name: "Kabupaten Sintang", Province: "Kalimantan Barat"},
		{Name: "Kabupaten Kapuas Hulu", Province: "Kalimantan Barat"},
		{Name: "Kabupaten Bengkayang", Province: "Kalimantan Barat"},
		{Name: "Kabupaten Landak", Province: "Kalimantan Barat"},
		{Name: "Kabupaten Sekadau", Province: "Kalimantan Barat"},
		{Name: "Kabupaten Melawi", Province: "Kalimantan Barat"},
		{Name: "Kabupaten Kayong Utara", Province: "Kalimantan Barat"},
		{Name: "Kabupaten Kubu Raya", Province: "Kalimantan Barat"},
		{Name: "Kota Pontianak", Province: "Kalimantan Barat"},
		{Name: "Kota Singkawang", Province: "Kalimantan Barat"},
		{Name: "Kabupaten Kotawaringin Barat", Province: "Kalimantan Tengah"},
		{Name: "Kabupaten Kotawaringin Timur", Province: "Kalimantan Tengah"},
		{Name: "Kabupaten Kapuas", Province: "Kalimantan Tengah"},
		{Name: "Kabupaten Barito Selatan", Province: "Kalimantan Tengah"},
		{Name: "Kabupaten Barito Utara", Province: "Kalimantan Tengah"},
		{Name: "Kabupaten Katingan", Province: "Kalimantan Tengah"},
		{Name: "Kabupaten Seruyan", Province: "Kalimantan Tengah"},
		{Name: "Kabupaten Sukamara", Province: "Kalimantan Tengah"},
		{Name: "Kabupaten Lamandau", Province: "Kalimantan Tengah"},
		{Name: "Kabupaten Gunung Mas", Province: "Kalimantan Tengah"},
		{Name: "Kabupaten Pulang Pisau", Province: "Kalimantan Tengah"},
		{Name: "Kabupaten Murung Raya", Province: "Kalimantan Tengah"},
		{Name: "Kabupaten Barito Timur", Province: "Kalimantan Tengah"},
		{Name: "Kota Palangkaraya", Province: "Kalimantan Tengah"},
		{Name: "Kabupaten Tanah Laut", Province: "Kalimantan Selatan"},
		{Name: "Kabupaten Kotabaru", Province: "Kalimantan Selatan"},
		{Name: "Kabupaten Banjar", Province: "Kalimantan Selatan"},
		{Name: "Kabupaten Barito Kuala", Province: "Kalimantan Selatan"},
		{Name: "Kabupaten Tapin", Province: "Kalimantan Selatan"},
		{Name: "Kabupaten Hulu Sungai Selatan", Province: "Kalimantan Selatan"},
		{Name: "Kabupaten Hulu Sungai Tengah", Province: "Kalimantan Selatan"},
		{Name: "Kabupaten Hulu Sungai Utara", Province: "Kalimantan Selatan"},
		{Name: "Kabupaten Tabalong", Province: "Kalimantan Selatan"},
		{Name: "Kabupaten Tanah Bumbu", Province: "Kalimantan Selatan"},
		{Name: "Kabupaten Balangan", Province: "Kalimantan Selatan"},
		{Name: "Kota Banjarmasin", Province: "Kalimantan Selatan"},
		{Name: "Kota Banjarbaru", Province: "Kalimantan Selatan"},
		{Name: "Kabupaten Paser", Province: "Kalimantan Timur"},
		{Name: "Kabupaten Kutai Kartanegara", Province: "Kalimantan Timur"},
		{Name: "Kabupaten Berau", Province: "Kalimantan Timur"},
		{Name: "Kabupaten Kutai Barat", Province: "Kalimantan Timur"},
		{Name: "Kabupaten Kutai Timur", Province: "Kalimantan Timur"},
		{Name: "Kabupaten Penajam Paser Utara", Province: "Kalimantan Timur"},
		{Name: "Kabupaten Mahakam Ulu", Province: "Kalimantan Timur"},
		{Name: "Kota Balikpapan", Province: "Kalimantan Timur"},
		{Name: "Kota Samarinda", Province: "Kalimantan Timur"},
		{Name: "Kota Bontang", Province: "Kalimantan Timur"},
		{Name: "Kabupaten Bulungan", Province: "Kalimantan Utara"},
		{Name: "Kabupaten Malinau", Province: "Kalimantan Utara"},
		{Name: "Kabupaten Nunukan", Province: "Kalimantan Utara"},
		{Name: "Kabupaten Tana Tidung", Province: "Kalimantan Utara"},
		{Name: "Kota Tarakan", Province: "Kalimantan Utara"},
		{Name: "Kabupaten Bolaang Mongondow", Province: "Sulawesi Utara"},
		{Name: "Kabupaten Minahasa", Province: "Sulawesi Utara"},
		{Name: "Kabupaten Kepulauan Sangihe", Province: "Sulawesi Utara"},
		{Name: "Kabupaten Kepulauan Talaud", Province: "Sulawesi Utara"},
		{Name: "Kabupaten Minahasa Selatan", Province: "Sulawesi Utara"},
		{Name: "Kabupaten Minahasa Utara", Province: "Sulawesi Utara"},
		{Name: "Kabupaten Minahasa Tenggara", Province: "Sulawesi Utara"},
		{Name: "Kabupaten Bolaang Mongondow Utara", Province: "Sulawesi Utara"},
		{Name: "Kabupaten Kepulauan Siau Tagulandang Biaro", Province: "Sulawesi Utara"},
		{Name: "Kabupaten Bolaang Mongondow Timur", Province: "Sulawesi Utara"},
		{Name: "Kabupaten Bolaang Mongondow Selatan", Province: "Sulawesi Utara"},
		{Name: "Kota Manado", Province: "Sulawesi Utara"},
		{Name: "Kota Bitung", Province: "Sulawesi Utara"},
		{Name: "Kota Tomohon", Province: "Sulawesi Utara"},
		{Name: "Kota Kotamobagu", Province: "Sulawesi Utara"},
		{Name: "Kabupaten Banggai", Province: "Sulawesi Tengah"},
		{Name: "Kabupaten Poso", Province: "Sulawesi Tengah"},
		{Name: "Kabupaten Donggala", Province: "Sulawesi Tengah"},
		{Name: "Kabupaten Toli-Toli", Province: "Sulawesi Tengah"},
		{Name: "Kabupaten Buol", Province: "Sulawesi Tengah"},
		{Name: "Kabupaten Morowali", Province: "Sulawesi Tengah"},
		{Name: "Kabupaten Banggai Kepulauan", Province: "Sulawesi Tengah"},
		{Name: "Kabupaten Parigi Moutong", Province: "Sulawesi Tengah"},
		{Name: "Kabupaten Tojo Una Una", Province: "Sulawesi Tengah"},
		{Name: "Kabupaten Sigi", Province: "Sulawesi Tengah"},
		{Name: "Kabupaten Banggai Laut", Province: "Sulawesi Tengah"},
		{Name: "Kabupaten Morowali Utara", Province: "Sulawesi Tengah"},
		{Name: "Kota Palu", Province: "Sulawesi Tengah"},
		{Name: "Kabupaten Kepulauan Selayar", Province: "Sulawesi Selatan"},
		{Name: "Kabupaten Bulukumba", Province: "Sulawesi Selatan"},
		{Name: "Kabupaten Bantaeng", Province: "Sulawesi Selatan"},
		{Name: "Kabupaten Jeneponto", Province: "Sulawesi Selatan"},
		{Name: "Kabupaten Takalar", Province: "Sulawesi Selatan"},
		{Name: "Kabupaten Gowa", Province: "Sulawesi Selatan"},
		{Name: "Kabupaten Sinjai", Province: "Sulawesi Selatan"},
		{Name: "Kabupaten Bone", Province: "Sulawesi Selatan"},
		{Name: "Kabupaten Maros", Province: "Sulawesi Selatan"},
		{Name: "Kabupaten Pangkajene dan Kepulauan", Province: "Sulawesi Selatan"},
		{Name: "Kabupaten Barru", Province: "Sulawesi Selatan"},
		{Name: "Kabupaten Soppeng", Province: "Sulawesi Selatan"},
		{Name: "Kabupaten Wajo", Province: "Sulawesi Selatan"},
		{Name: "Kabupaten Sidenreng Rappang", Province: "Sulawesi Selatan"},
		{Name: "Kabupaten Pinrang", Province: "Sulawesi Selatan"},
		{Name: "Kabupaten Enrekang", Province: "Sulawesi Selatan"},
		{Name: "Kabupaten Luwu", Province: "Sulawesi Selatan"},
		{Name: "Kabupaten Tana Toraja", Province: "Sulawesi Selatan"},
		{Name: "Kabupaten Luwu Utara", Province: "Sulawesi Selatan"},
		{Name: "Kabupaten Luwu Timur", Province: "Sulawesi Selatan"},
		{Name: "Kabupaten Toraja Utara", Province: "Sulawesi Selatan"},
		{Name: "Kota Makassar", Province: "Sulawesi Selatan"},
		{Name: "Kota Parepare", Province: "Sulawesi Selatan"},
		{Name: "Kota Palopo", Province: "Sulawesi Selatan"},
		{Name: "Kabupaten Kolaka", Province: "Sulawesi Tenggara"},
		{Name: "Kabupaten Konawe", Province: "Sulawesi Tenggara"},
		{Name: "Kabupaten Muna", Province: "Sulawesi Tenggara"},
		{Name: "Kabupaten Buton", Province: "Sulawesi Tenggara"},
		{Name: "Kabupaten Konawe Selatan", Province: "Sulawesi Tenggara"},
		{Name: "Kabupaten Bombana", Province: "Sulawesi Tenggara"},
		{Name: "Kabupaten Wakatobi", Province: "Sulawesi Tenggara"},
		{Name: "Kabupaten Kolaka Utara", Province: "Sulawesi Tenggara"},
		{Name: "Kabupaten Konawe Utara", Province: "Sulawesi Tenggara"},
		{Name: "Kabupaten Buton Utara", Province: "Sulawesi Tenggara"},
		{Name: "Kabupaten Kolaka Timur", Province: "Sulawesi Tenggara"},
		{Name: "Kabupaten Konawe Kepulauan", Province: "Sulawesi Tenggara"},
		{Name: "Kabupaten Muna Barat", Province: "Sulawesi Tenggara"},
		{Name: "Kabupaten Buton Tengah", Province: "Sulawesi Tenggara"},
		{Name: "Kabupaten Buton Selatan", Province: "Sulawesi Tenggara"},
		{Name: "Kota Kendari", Province: "Sulawesi Tenggara"},
		{Name: "Kota Bau Bau", Province: "Sulawesi Tenggara"},
		{Name: "Kabupaten Gorontalo", Province: "Gorontalo"},
		{Name: "Kabupaten Boalemo", Province: "Gorontalo"},
		{Name: "Kabupaten Bone Bolango", Province: "Gorontalo"},
		{Name: "Kabupaten Pohuwato", Province: "Gorontalo"},
		{Name: "Kabupaten Gorontalo Utara", Province: "Gorontalo"},
		{Name: "Kota Gorontalo", Province: "Gorontalo"},
		{Name: "Kabupaten Pasangkayu", Province: "Sulawesi Barat"},
		{Name: "Kabupaten Mamuju", Province: "Sulawesi Barat"},
		{Name: "Kabupaten Mamasa", Province: "Sulawesi Barat"},
		{Name: "Kabupaten Polewali Mandar", Province: "Sulawesi Barat"},
		{Name: "Kabupaten Majene", Province: "Sulawesi Barat"},
		{Name: "Kabupaten Mamuju Tengah", Province: "Sulawesi Barat"},
		{Name: "Kabupaten Maluku Tengah", Province: "Maluku"},
		{Name: "Kabupaten Maluku Tenggara", Province: "Maluku"},
		{Name: "Kabupaten Kepulauan Tanimbar", Province: "Maluku"},
		{Name: "Kabupaten Buru", Province: "Maluku"},
		{Name: "Kabupaten Seram Bagian Timur", Province: "Maluku"},
		{Name: "Kabupaten Seram Bagian Barat", Province: "Maluku"},
		{Name: "Kabupaten Kepulauan Aru", Province: "Maluku"},
		{Name: "Kabupaten Maluku Barat Daya", Province: "Maluku"},
		{Name: "Kabupaten Buru Selatan", Province: "Maluku"},
		{Name: "Kota Ambon", Province: "Maluku"},
		{Name: "Kota Tual", Province: "Maluku"},
		{Name: "Kabupaten Halmahera Barat", Province: "Maluku Utara"},
		{Name: "Kabupaten Halmahera Tengah", Province: "Maluku Utara"},
		{Name: "Kabupaten Halmahera Utara", Province: "Maluku Utara"},
		{Name: "Kabupaten Halmahera Selatan", Province: "Maluku Utara"},
		{Name: "Kabupaten Kepulauan Sula", Province: "Maluku Utara"},
		{Name: "Kabupaten Halmahera Timur", Province: "Maluku Utara"},
		{Name: "Kabupaten Pulau Morotai", Province: "Maluku Utara"},
		{Name: "Kabupaten Pulau Taliabu", Province: "Maluku Utara"},
		{Name: "Kota Ternate", Province: "Maluku Utara"},
		{Name: "Kota Tidore Kepulauan", Province: "Maluku Utara"},
		{Name: "Kabupaten Jayapura", Province: "Papua"},
		{Name: "Kabupaten Kepulauan Yapen", Province: "Papua"},
		{Name: "Kabupaten Biak Numfor", Province: "Papua"},
		{Name: "Kabupaten Sarmi", Province: "Papua"},
		{Name: "Kabupaten Keerom", Province: "Papua"},
		{Name: "Kabupaten Waropen", Province: "Papua"},
		{Name: "Kabupaten Supiori", Province: "Papua"},
		{Name: "Kabupaten Mamberamo Raya", Province: "Papua"},
		{Name: "Kota Jayapura", Province: "Papua"},
		{Name: "Kabupaten Manokwari", Province: "Papua Barat"},
		{Name: "Kabupaten Fak Fak", Province: "Papua Barat"},
		{Name: "Kabupaten Teluk Bintuni", Province: "Papua Barat"},
		{Name: "Kabupaten Teluk Wondama", Province: "Papua Barat"},
		{Name: "Kabupaten Kaimana", Province: "Papua Barat"},
		{Name: "Kabupaten Manokwari Selatan", Province: "Papua Barat"},
		{Name: "Kabupaten Pegunungan Arfak", Province: "Papua Barat"},
		{Name: "Kabupaten Merauke", Province: "Papua Selatan"},
		{Name: "Kabupaten Boven Digoel", Province: "Papua Selatan"},
		{Name: "Kabupaten Mappi", Province: "Papua Selatan"},
		{Name: "Kabupaten Asmat", Province: "Papua Selatan"},
		{Name: "Kabupaten Nabire", Province: "Papua Tengah"},
		{Name: "Kabupaten Puncak Jaya", Province: "Papua Tengah"},
		{Name: "Kabupaten Paniai", Province: "Papua Tengah"},
		{Name: "Kabupaten Mimika", Province: "Papua Tengah"},
		{Name: "Kabupaten Puncak", Province: "Papua Tengah"},
		{Name: "Kabupaten Dogiyai", Province: "Papua Tengah"},
		{Name: "Kabupaten Intan Jaya", Province: "Papua Tengah"},
		{Name: "Kabupaten Deiyai", Province: "Papua Tengah"},
		{Name: "Kabupaten Jayawijaya", Province: "Papua Pegunungan"},
		{Name: "Kabupaten Pegunungan Bintang", Province: "Papua Pegunungan"},
		{Name: "Kabupaten Yahukimo", Province: "Papua Pegunungan"},
		{Name: "Kabupaten Tolikara", Province: "Papua Pegunungan"},
		{Name: "Kabupaten Mamberamo Tengah", Province: "Papua Pegunungan"},
		{Name: "Kabupaten Yalimo", Province: "Papua Pegunungan"},
		{Name: "Kabupaten Lanny Jaya", Province: "Papua Pegunungan"},
		{Name: "Kabupaten Nduga", Province: "Papua Pegunungan"},
		{Name: "Kabupaten Sorong", Province: "Papua Barat Daya"},
		{Name: "Kabupaten Sorong Selatan", Province: "Papua Barat Daya"},
		{Name: "Kabupaten Raja Ampat", Province: "Papua Barat Daya"},
		{Name: "Kabupaten Tambrauw", Province: "Papua Barat Daya"},
		{Name: "Kabupaten Maybrat", Province: "Papua Barat Daya"},
		{Name: "Kota Sorong", Province: "Papua Barat Daya"},
	}

	// One-time replace: the old seed list held ~46 plain city names (no
	// Kabupaten/Kota distinction) that share no exact name with the official
	// list above, so they'd otherwise sit alongside it as stale duplicates
	// forever. KotaAsal embeds AuditModel, so this is a soft delete (sets
	// deleted_at) — any peserta.origin_city_id still pointing at an old row
	// just stops resolving via Preload rather than breaking a live FK.
	names := make([]string, len(cities))
	for i, c := range cities {
		names[i] = c.Name
	}
	db.Where("name NOT IN ?", names).Delete(&kota_asal.KotaAsal{})

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

// seedBlazerSizes seeds the size rows the website's blazer size picker (see
// modules/peserta and the website's ShirtSizeTab) reads stock from. Stock
// starts at 0 for every size — real counts aren't known at seed time, and
// seeding a made-up number would misrepresent actual inventory. An admin
// must set real stock via the Blazer Size admin page before sizes show as
// available.
func seedBlazerSizes(db *gorm.DB) {
	sizes := []string{"XS", "S", "M", "L", "XL", "XXL", "XXXL"}
	for i, size := range sizes {
		var existing blazer_sizes.BlazerSize
		db.Where("size = ?", size).Attrs(blazer_sizes.BlazerSize{Size: size, Stock: 0, Order: i}).FirstOrCreate(&existing)
	}
}
