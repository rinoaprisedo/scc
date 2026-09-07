package peserta

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"regexp"
	"strings"
	"time"

	"baseadmin/backend/modules/activity_logs"
	"baseadmin/backend/modules/users"
	"baseadmin/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// defaultImportPassword mirrors PesertaFormModal's DEFAULT_PASSWORD on the
// admin frontend — every imported row gets this same password rather than
// a per-row column, matching the request that import skip password (and
// KTP photo, which can't meaningfully come from a spreadsheet cell anyway).
const defaultImportPassword = "scc2026"

// importColumnDefs is the single source of truth for the import template's
// header row AND how those headers map back to fields when a filled-in
// template is re-uploaded — keeping both in sync by construction rather
// than by two lists an editor has to remember to update together.
// exportColumns (export.go) reuses these same headers, so an exported file
// can be edited and re-imported as-is. There is no "Status" column here —
// every imported peserta is always created active, so there's nothing for
// the sheet to carry.
var importColumnDefs = []struct {
	Header string
	Field  string
}{
	{"Name", "name"},
	{"NIK", "nik"},
	{"Nomor KTP", "nomor_ktp"},
	{"Email", "email"},
	{"Title", "title"},
	{"Nama Depan", "first_name"},
	{"Nama Tengah", "middle_name"},
	{"Nama Belakang", "last_name"},
	{"Tanggal Lahir (YYYY-MM-DD)", "birth_date"},
	{"Kota Asal", "origin_city"},
	{"Bandara Terdekat", "nearest_airport"},
	{"Pantangan Makanan", "dietary_restriction"},
	{"Nomor HP", "phone_number"},
	{"Region", "region"},
	{"Cabang", "cabang"},
	{"Position", "position"},
	{"Nomor Passport", "passport_number"},
	{"Masa Berlaku Passport (YYYY-MM-DD)", "passport_expiry"},
	{"Blazer Size", "blazer_size"},
	{"Nomor Meja", "nomor_meja"},
	{"Deskripsi", "description"},
}

var importHeaderToField = func() map[string]string {
	m := make(map[string]string, len(importColumnDefs))
	for _, c := range importColumnDefs {
		m[strings.ToLower(c.Header)] = c.Field
	}
	return m
}()

// importTitleOptions/importDietaryOptions/importSizeOptions mirror the
// dropdown choices in PesertaFormModal.jsx on the admin frontend — kept in
// sync by hand since one lives in Go and the other in JS. Used both to
// populate the template's Excel data-validation dropdowns and to reject a
// re-uploaded row whose value doesn't match any of them (e.g. pasted over
// the dropdown, or edited outside Excel).
var importTitleOptions = []string{"Mr", "Mrs", "Ms"}
var importDietaryOptions = []string{"tidak ada pantangan", "tidak makan daging", "tidak makan ayam", "tidak makan seafood", "vegetarian"}
var importSizeOptions = []string{"S", "M", "L", "XL", "XXL", "XXXL"}

var importEmailRegex = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// importDigitsRegex backs NIK/Nomor KTP format validation — both are
// Indonesian ID numbers, digits only.
var importDigitsRegex = regexp.MustCompile(`^[0-9]+$`)

func containsOptionFold(options []string, v string) bool {
	for _, o := range options {
		if strings.EqualFold(o, v) {
			return true
		}
	}
	return false
}

type ImportRowError struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

// ImportPreview is the read-only dry-run result — how many data rows the
// sheet has, how many would successfully import, and every row that
// wouldn't (with why). Returned by ValidateImportExcel, called before the
// real import so the admin can fix the file first; nothing is written to
// the database to produce this.
type ImportPreview struct {
	TotalRows int              `json:"total_rows"`
	Valid     int              `json:"valid"`
	// ToCreate/ToUpdate split Valid by whether the row's NIK matches an
	// existing peserta (see prepareImportRows) — shown in the admin
	// frontend's preview step so it's clear the import isn't purely
	// additive.
	ToCreate int              `json:"to_create"`
	ToUpdate int              `json:"to_update"`
	Errors   []ImportRowError `json:"errors"`
}

// ImportResult is the real import's success response — every row in the
// file either all imports (all-or-nothing, see ImportExcel) or none do, so
// there's nothing left to report per row on success.
type ImportResult struct {
	Created int `json:"created"`
	Updated int `json:"updated"`
}

// preparedRow is one fully-validated, ready-to-insert-or-update row.
// NewAirportName is set instead of resolving/creating the bandara row
// immediately — doing that write only happens once every row in the file
// has passed validation (see ImportExcel), so a validation-only pass never
// mutates master data. IsUpdate means User was seeded from an existing
// peserta found by NIK (see prepareImportRows) and should be saved over
// that row rather than inserted as a new one.
type preparedRow struct {
	RowNum         int
	User           users.User
	NewAirportName string
	IsUpdate       bool
}

// mergeStr picks the sheet's value when the column was filled in for this
// row, or keeps whatever the existing peserta already had when the column
// was left blank on an update row — so re-importing a partially-filled
// sheet only overwrites the cells it actually specifies, never clears a
// field just because that column was skipped. Always returns nil for a
// blank cell on a create row (existing is nil there).
func mergeStr(existing *string, value string, isUpdate bool) *string {
	if value != "" {
		return ptrOrNil(value)
	}
	if isUpdate {
		return existing
	}
	return nil
}

func allBlank(values map[string]string) bool {
	for _, v := range values {
		if v != "" {
			return false
		}
	}
	return true
}

// prepareImportRows parses and validates every data row in the uploaded
// sheet without writing anything to the database — shared by
// ValidateImportExcel (the dry-run preview) and ImportExcel (the real
// import re-validates from scratch before committing, since duplicate
// NIK/email checks against the DB could go stale between preview and
// confirm). A row that fails validation is recorded in rowErrors and
// otherwise skipped; totalRows counts every non-blank data row seen,
// valid or not.
func (s *Service) prepareImportRows(f *excelize.File) (prepared []preparedRow, rowErrors []ImportRowError, totalRows int, err error) {
	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil || len(rows) == 0 {
		return nil, nil, 0, errors.New("sheet is empty")
	}

	// rowErrors starts as an empty (non-nil) slice rather than the zero-value
	// nil so a clean file's response is JSON "errors": [] — the frontend
	// always does preview.errors.length, and Go's encoding/json renders a
	// nil slice as null, which would crash that read.
	rowErrors = []ImportRowError{}

	fieldByCol := map[int]string{}
	for i, header := range rows[0] {
		if field, ok := importHeaderToField[strings.ToLower(strings.TrimSpace(header))]; ok {
			fieldByCol[i] = field
		}
	}

	roleID, err := s.repo.PesertaRoleID()
	if err != nil {
		return nil, nil, 0, err
	}
	hashedDefault, err := bcrypt.GenerateFromPassword([]byte(defaultImportPassword), 12)
	if err != nil {
		return nil, nil, 0, err
	}

	seenNIK := map[string]int{}
	seenEmail := map[string]int{}

	for i, row := range rows[1:] {
		rowNum := i + 2 // 1-indexed, plus the header row
		values := map[string]string{}
		for col, cell := range row {
			if field, ok := fieldByCol[col]; ok {
				values[field] = strings.TrimSpace(cell)
			}
		}
		if allBlank(values) {
			continue // a fully empty trailing row isn't a data row at all
		}
		totalRows++

		if values["name"] == "" || values["nik"] == "" {
			rowErrors = append(rowErrors, ImportRowError{Row: rowNum, Message: "Name and NIK are required"})
			continue
		}

		nik := values["nik"]
		if !importDigitsRegex.MatchString(nik) {
			rowErrors = append(rowErrors, ImportRowError{Row: rowNum, Message: "NIK must contain digits only"})
			continue
		}
		if prevRow, ok := seenNIK[nik]; ok {
			rowErrors = append(rowErrors, ImportRowError{Row: rowNum, Message: fmt.Sprintf("Duplicate NIK — already used in row %d of this file", prevRow)})
			continue
		}
		// A NIK that already belongs to a peserta is an update row, not a
		// rejection — see mergeStr and the user-construction branch below.
		existing, err := s.repo.FindByNIK(nik)
		if err != nil {
			return nil, nil, 0, err
		}
		isUpdate := existing != nil

		email := values["email"]
		if email != "" && !importEmailRegex.MatchString(email) {
			rowErrors = append(rowErrors, ImportRowError{Row: rowNum, Message: "Email format is invalid"})
			continue
		}
		if email == "" {
			if isUpdate {
				email = existing.Email
			} else {
				email = nik + placeholderEmailDomain
			}
		}
		emailKey := strings.ToLower(email)
		if prevRow, ok := seenEmail[emailKey]; ok {
			rowErrors = append(rowErrors, ImportRowError{Row: rowNum, Message: fmt.Sprintf("Duplicate email — already used in row %d of this file", prevRow)})
			continue
		}
		var emailConflict bool
		if isUpdate {
			// Excludes the row's own existing account — its current email
			// coming back around (blank cell defaulted above, or resent
			// unchanged) must not "conflict" with itself.
			emailConflict, err = s.repo.EmailExistsExcluding(email, existing.ID)
		} else {
			emailConflict, err = s.repo.EmailExists(email)
		}
		if err != nil {
			return nil, nil, 0, err
		}
		if emailConflict {
			rowErrors = append(rowErrors, ImportRowError{Row: rowNum, Message: "Email is already registered"})
			continue
		}

		if values["title"] != "" && !containsOptionFold(importTitleOptions, values["title"]) {
			rowErrors = append(rowErrors, ImportRowError{Row: rowNum, Message: "Title must be one of: " + strings.Join(importTitleOptions, ", ")})
			continue
		}
		if values["dietary_restriction"] != "" && !containsOptionFold(importDietaryOptions, values["dietary_restriction"]) {
			rowErrors = append(rowErrors, ImportRowError{Row: rowNum, Message: "Pantangan Makanan must be one of: " + strings.Join(importDietaryOptions, ", ")})
			continue
		}
		if values["blazer_size"] != "" && !containsOptionFold(importSizeOptions, values["blazer_size"]) {
			rowErrors = append(rowErrors, ImportRowError{Row: rowNum, Message: "Blazer Size must be one of: " + strings.Join(importSizeOptions, ", ")})
			continue
		}

		if values["nomor_ktp"] != "" && !importDigitsRegex.MatchString(values["nomor_ktp"]) {
			rowErrors = append(rowErrors, ImportRowError{Row: rowNum, Message: "Nomor KTP must contain digits only"})
			continue
		}

		// parseDate silently returns nil on anything it can't parse — fine
		// for a truly blank cell, but a malformed one (wrong format, stray
		// text) must not just quietly vanish into a null column value.
		var birthDate *time.Time
		if values["birth_date"] != "" {
			birthDate = parseDate(values["birth_date"])
			if birthDate == nil {
				rowErrors = append(rowErrors, ImportRowError{Row: rowNum, Message: "Tanggal Lahir must be in YYYY-MM-DD format"})
				continue
			}
		} else if isUpdate {
			birthDate = existing.BirthDate
		}
		var passportExpiry *time.Time
		if values["passport_expiry"] != "" {
			passportExpiry = parseDate(values["passport_expiry"])
			if passportExpiry == nil {
				rowErrors = append(rowErrors, ImportRowError{Row: rowNum, Message: "Masa Berlaku Passport must be in YYYY-MM-DD format"})
				continue
			}
		} else if isUpdate {
			passportExpiry = existing.PassportExpiry
		}

		// Unlike the admin/website form (which has an explicit "Other" choice
		// backed by OriginCityOther free text), the import sheet's Kota Asal
		// column only offers a dropdown of real master data — so a value
		// that doesn't match one of those options is treated as a mistake
		// (typo, stale copy-paste) and rejected, not silently stored as
		// free-text "Other".
		var cityID *uint64
		if values["origin_city"] != "" {
			cityID, err = s.repo.FindCityIDByName(values["origin_city"])
			if err != nil {
				return nil, nil, 0, err
			}
			if cityID == nil {
				rowErrors = append(rowErrors, ImportRowError{Row: rowNum, Message: "Kota Asal must match one of the dropdown options"})
				continue
			}
		} else if isUpdate {
			cityID = existing.OriginCityID
		}

		var airportID *uint64
		newAirportName := ""
		if values["nearest_airport"] != "" {
			airportID, err = s.repo.FindAirportIDByName(values["nearest_airport"])
			if err != nil {
				return nil, nil, 0, err
			}
			if airportID == nil {
				newAirportName = values["nearest_airport"]
			}
		} else if isUpdate {
			airportID = existing.NearestAirportID
		}

		seenNIK[nik] = rowNum
		seenEmail[emailKey] = rowNum

		var user users.User
		if isUpdate {
			// Seed from the existing row so every column not touched below
			// (password, status, attendance status, role, KTP/passport
			// files, created_at/created_by, ...) is carried through as-is —
			// only the fields this sheet actually supplies get overwritten.
			user = *existing
			user.Name = values["name"]
			user.Email = email
			user.Title = mergeStr(existing.Title, values["title"], true)
			user.FirstName = mergeStr(existing.FirstName, values["first_name"], true)
			user.MiddleName = mergeStr(existing.MiddleName, values["middle_name"], true)
			user.LastName = mergeStr(existing.LastName, values["last_name"], true)
			user.BirthDate = birthDate
			user.OriginCityID = cityID
			user.NearestAirportID = airportID
			user.DietaryRestriction = mergeStr(existing.DietaryRestriction, values["dietary_restriction"], true)
			user.PhoneNumber = mergeStr(existing.PhoneNumber, values["phone_number"], true)
			user.Region = mergeStr(existing.Region, values["region"], true)
			user.Cabang = mergeStr(existing.Cabang, values["cabang"], true)
			user.Position = mergeStr(existing.Position, values["position"], true)
			user.NomorKtp = mergeStr(existing.NomorKtp, values["nomor_ktp"], true)
			user.PassportNumber = mergeStr(existing.PassportNumber, values["passport_number"], true)
			user.PassportExpiry = passportExpiry
			user.BlazerSize = mergeStr(existing.BlazerSize, values["blazer_size"], true)
			user.NomorMeja = mergeStr(existing.NomorMeja, values["nomor_meja"], true)
			user.Description = mergeStr(existing.Description, values["description"], true)
			recomputeAttendanceStatus(&user)
		} else {
			attendanceStatus := StatusBelumKonfirmasi
			user = users.User{
				Name:               values["name"],
				Email:              email,
				Password:           string(hashedDefault),
				Status:             users.StatusActive,
				RoleID:             &roleID,
				AttendanceStatus:   &attendanceStatus,
				Title:              ptrOrNil(values["title"]),
				FirstName:          ptrOrNil(values["first_name"]),
				MiddleName:         ptrOrNil(values["middle_name"]),
				LastName:           ptrOrNil(values["last_name"]),
				BirthDate:          birthDate,
				OriginCityID:       cityID,
				NearestAirportID:   airportID,
				DietaryRestriction: ptrOrNil(values["dietary_restriction"]),
				PhoneNumber:        ptrOrNil(values["phone_number"]),
				Region:             ptrOrNil(values["region"]),
				Cabang:             ptrOrNil(values["cabang"]),
				Position:           ptrOrNil(values["position"]),
				KtpNumber:          ptrOrNil(nik),
				NomorKtp:           ptrOrNil(values["nomor_ktp"]),
				PassportNumber:     ptrOrNil(values["passport_number"]),
				PassportExpiry:     passportExpiry,
				BlazerSize:         ptrOrNil(values["blazer_size"]),
				NomorMeja:          ptrOrNil(values["nomor_meja"]),
				Description:        ptrOrNil(values["description"]),
			}
		}

		prepared = append(prepared, preparedRow{RowNum: rowNum, User: user, NewAirportName: newAirportName, IsUpdate: isUpdate})
	}

	return prepared, rowErrors, totalRows, nil
}

// ValidateImportExcel is the dry-run preview behind the admin frontend's
// "verify before import" step — parses and validates the file exactly like
// ImportExcel would, but never opens a transaction or writes anything.
func (s *Service) ValidateImportExcel(file multipart.File) (*ImportPreview, error) {
	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, errors.New("invalid Excel file")
	}
	defer f.Close()

	prepared, rowErrors, totalRows, err := s.prepareImportRows(f)
	if err != nil {
		return nil, err
	}
	toCreate, toUpdate := 0, 0
	for _, p := range prepared {
		if p.IsUpdate {
			toUpdate++
		} else {
			toCreate++
		}
	}
	return &ImportPreview{TotalRows: totalRows, Valid: len(prepared), ToCreate: toCreate, ToUpdate: toUpdate, Errors: rowErrors}, nil
}

// ImportExcel parses an uploaded .xlsx and creates or updates one Peserta
// per data row — a row whose NIK matches an existing peserta updates it
// (only the columns the sheet actually fills in, see mergeStr/
// prepareImportRows), any other row creates a new one — all inside a single
// transaction. Unlike a per-row-independent import, this is deliberately
// all-or-nothing: if even one row fails validation, nothing in the file is
// written — rowErrors is returned instead so the caller can show exactly
// which rows/why, matching the admin frontend's "fix the whole file, then
// import" flow rather than a partial import silently skipping bad rows.
func (s *Service) ImportExcel(file multipart.File, actorID *uint64) (*ImportResult, []ImportRowError, error) {
	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, nil, errors.New("invalid Excel file")
	}
	defer f.Close()

	prepared, rowErrors, _, err := s.prepareImportRows(f)
	if err != nil {
		return nil, nil, err
	}
	if len(rowErrors) > 0 {
		return nil, rowErrors, nil
	}
	if len(prepared) == 0 {
		return nil, nil, errors.New("no data rows to import")
	}

	// Two rows can reference the same not-yet-existing airport name; dedupe
	// within this batch so it isn't created twice inside one transaction.
	createdAirports := map[string]uint64{}
	created, updated := 0, 0
	err = s.repo.DB.Transaction(func(tx *gorm.DB) error {
		for i := range prepared {
			p := &prepared[i]
			if p.NewAirportName != "" {
				key := strings.ToLower(p.NewAirportName)
				if id, ok := createdAirports[key]; ok {
					p.User.NearestAirportID = &id
				} else {
					id, err := s.repo.CreateAirportTx(tx, p.NewAirportName, actorID)
					if err != nil {
						return fmt.Errorf("row %d: failed to create Bandara Terdekat: %w", p.RowNum, err)
					}
					createdAirports[key] = *id
					p.User.NearestAirportID = id
				}
			}
			if p.IsUpdate {
				p.User.UpdatedBy = actorID
				if err := s.repo.SaveTx(tx, &p.User); err != nil {
					return fmt.Errorf("row %d: %w", p.RowNum, err)
				}
				updated++
			} else {
				p.User.CreatedBy = actorID
				if err := tx.Create(&p.User).Error; err != nil {
					return fmt.Errorf("row %d: %w", p.RowNum, err)
				}
				created++
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return &ImportResult{Created: created, Updated: updated}, nil, nil
}

// AllCityNames/AllAirportNames back the import template's dropdown columns
// — thin delegations so the handler doesn't need to reach past Service into
// repo directly.
func (s *Service) AllCityNames() ([]string, error)    { return s.repo.AllCityNames() }
func (s *Service) AllAirportNames() ([]string, error) { return s.repo.AllAirportNames() }

// importListSheet is a hidden helper sheet on the generated template
// holding the full Kota Asal/Bandara Terdekat name lists — Excel's inline
// dropdown list (SetDropList) is capped at 255 characters, too short for
// real master data, so those two columns' data validation instead
// references a range on this sheet (SetSqrefDropList).
const importListSheet = "Lists"

const importTemplateMaxRow = 1000

// addDropdown attaches an inline-list dropdown (short, fixed option sets)
// to every data row of one template column.
func addDropdown(f *excelize.File, sheet string, colIndex int, options []string) {
	col, err := excelize.ColumnNumberToName(colIndex + 1)
	if err != nil {
		return
	}
	dv := excelize.NewDataValidation(true)
	dv.Sqref = fmt.Sprintf("%s2:%s%d", col, col, importTemplateMaxRow)
	if err := dv.SetDropList(options); err != nil {
		return
	}
	_ = f.AddDataValidation(sheet, dv)
}

// addRefDropdown attaches a dropdown sourced from a range on the hidden
// helper sheet — used for the two columns backed by DB master data
// (Kota Asal/Bandara Terdekat), which can be far longer than the 255-char
// inline list limit allows.
func addRefDropdown(f *excelize.File, sheet string, colIndex int, listCol string, count int) {
	if count == 0 {
		return
	}
	col, err := excelize.ColumnNumberToName(colIndex + 1)
	if err != nil {
		return
	}
	dv := excelize.NewDataValidation(true)
	dv.Sqref = fmt.Sprintf("%s2:%s%d", col, col, importTemplateMaxRow)
	dv.SetSqrefDropList(fmt.Sprintf("%s!$%s$1:$%s$%d", importListSheet, listCol, listCol, count))
	_ = f.AddDataValidation(sheet, dv)
}

// ImportTemplate godoc
// @Summary		Download the peserta Excel import template
// @Tags			peserta
// @Security		SessionCookie
// @Produce		application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Success		200
// @Router			/peserta/import/template [get]
func (h *Handler) ImportTemplate(c *gin.Context) {
	cityNames, err := h.Service.AllCityNames()
	if err != nil {
		utils.Error(c, 500, "failed to build import template")
		return
	}
	airportNames, err := h.Service.AllAirportNames()
	if err != nil {
		utils.Error(c, 500, "failed to build import template")
		return
	}

	f := excelize.NewFile()
	defer f.Close()
	const sheet = "Peserta"
	f.SetSheetName("Sheet1", sheet)

	colIndex := map[string]int{}
	for i, col := range importColumnDefs {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, col.Header)
		colIndex[col.Field] = i
	}

	if _, err := f.NewSheet(importListSheet); err != nil {
		utils.Error(c, 500, "failed to build import template")
		return
	}
	for i, name := range cityNames {
		cell, _ := excelize.CoordinatesToCellName(1, i+1)
		f.SetCellValue(importListSheet, cell, name)
	}
	for i, name := range airportNames {
		cell, _ := excelize.CoordinatesToCellName(2, i+1)
		f.SetCellValue(importListSheet, cell, name)
	}
	_ = f.SetSheetVisible(importListSheet, false)

	addDropdown(f, sheet, colIndex["title"], importTitleOptions)
	addDropdown(f, sheet, colIndex["dietary_restriction"], importDietaryOptions)
	addDropdown(f, sheet, colIndex["blazer_size"], importSizeOptions)
	addRefDropdown(f, sheet, colIndex["origin_city"], "A", len(cityNames))
	addRefDropdown(f, sheet, colIndex["nearest_airport"], "B", len(airportNames))

	c.Header("Content-Disposition", "attachment; filename=peserta_import_template.xlsx")
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	if err := f.Write(c.Writer); err != nil {
		c.Status(http.StatusInternalServerError)
	}
}

// ValidateImport godoc
// @Summary		Dry-run validate an Excel file before importing peserta
// @Tags			peserta
// @Security		SessionCookie
// @Accept			multipart/form-data
// @Param			file	formData	file	true	"Excel file (.xlsx)"
// @Success		200		{object}	utils.Response
// @Failure		400		{object}	utils.Response
// @Router			/peserta/import/validate [post]
func (h *Handler) ValidateImport(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		utils.Error(c, 400, "file is required")
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		utils.Error(c, 400, "failed to read uploaded file")
		return
	}
	defer file.Close()

	preview, err := h.Service.ValidateImportExcel(file)
	if err != nil {
		utils.Error(c, 400, err.Error())
		return
	}
	utils.Success(c, 200, "validation completed", preview)
}

// ImportExcel godoc
// @Summary		Bulk-import peserta from an Excel file (all-or-nothing)
// @Tags			peserta
// @Security		SessionCookie
// @Accept			multipart/form-data
// @Param			file	formData	file	true	"Excel file (.xlsx)"
// @Success		200		{object}	utils.Response
// @Failure		400		{object}	utils.Response
// @Router			/peserta/import [post]
func (h *Handler) ImportExcel(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		utils.Error(c, 400, "file is required")
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		utils.Error(c, 400, "failed to read uploaded file")
		return
	}
	defer file.Close()

	actorID := utils.CurrentUserID(c)
	result, rowErrors, err := h.Service.ImportExcel(file, actorID)
	if err != nil {
		utils.Error(c, 400, err.Error())
		return
	}
	if len(rowErrors) > 0 {
		c.JSON(400, utils.Response{Success: false, Message: "import cancelled — some rows are invalid", Data: gin.H{"errors": rowErrors}})
		return
	}

	activity_logs.LogActivity(actorID, activity_logs.ActionCreate, "peserta", "",
		nil, gin.H{"import_created": result.Created, "import_updated": result.Updated},
		c.ClientIP(), c.Request.UserAgent())
	utils.Success(c, 200, "import completed", result)
}
