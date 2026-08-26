package peserta

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"baseadmin/backend/modules/users"
	"baseadmin/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// exportColumns reuses importColumnDefs' headers (import.go) so an exported
// file can be edited and re-imported as-is — plus one export-only column
// that isn't part of the import shape at all ("Kehadiran"). No "Status"
// column: every peserta is active by default and status isn't tracked as
// export-worthy here. Order must match exportRow's return order exactly.
var exportColumns = func() []string {
	cols := make([]string, 0, len(importColumnDefs)+1)
	for _, c := range importColumnDefs {
		cols = append(cols, c.Header)
	}
	return append(cols, "Kehadiran")
}()

var attendanceStatusLabels = map[string]string{
	StatusBelumKonfirmasi:   "Belum Konfirmasi",
	StatusHadirBelumLengkap: "Hadir, Belum Isi Form",
	StatusTidakHadir:        "Tidak Hadir",
	StatusHadirLengkap:      "Hadir, Data Lengkap",
}

func attendanceStatusLabel(status *string) string {
	if status != nil {
		if label, ok := attendanceStatusLabels[*status]; ok {
			return label
		}
	}
	return attendanceStatusLabels[StatusBelumKonfirmasi]
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func dateVal(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(dateLayout)
}

// exportRow's field order must match exportColumns exactly.
func exportRow(u users.User) []string {
	originCity := strVal(u.OriginCityOther)
	if u.OriginCity != nil {
		originCity = u.OriginCity.Name
	}
	airport := ""
	if u.NearestAirport != nil {
		airport = u.NearestAirport.Name
	}
	return []string{
		u.Name,
		strVal(u.KtpNumber),
		strVal(u.NomorKtp),
		u.Email,
		strVal(u.Title),
		strVal(u.FirstName),
		strVal(u.MiddleName),
		strVal(u.LastName),
		dateVal(u.BirthDate),
		originCity,
		airport,
		strVal(u.DietaryRestriction),
		strVal(u.PhoneNumber),
		strVal(u.PassportNumber),
		dateVal(u.PassportExpiry),
		strVal(u.BlazerSize),
		strVal(u.NomorMeja),
		strVal(u.Description),
		attendanceStatusLabel(u.AttendanceStatus),
	}
}

// ExportCSV godoc
// @Summary		Export peserta as CSV
// @Tags			peserta
// @Security		SessionCookie
// @Param			search	query	string	false	"Search by name/email/phone/KTP/passport"
// @Produce		text/csv
// @Success		200
// @Router			/peserta/export/csv [get]
func (h *Handler) ExportCSV(c *gin.Context) {
	list, err := h.Service.ListAll(c.Query("search"))
	if err != nil {
		utils.Error(c, 500, "failed to export peserta")
		return
	}

	filename := fmt.Sprintf("peserta_%s.csv", time.Now().Format("20060102_150405"))
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "text/csv")

	w := csv.NewWriter(c.Writer)
	_ = w.Write(exportColumns)
	for _, u := range list {
		_ = w.Write(exportRow(u))
	}
	w.Flush()
}

// ExportExcel godoc
// @Summary		Export peserta as Excel
// @Tags			peserta
// @Security		SessionCookie
// @Param			search	query	string	false	"Search by name/email/phone/KTP/passport"
// @Produce		application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Success		200
// @Router			/peserta/export/excel [get]
func (h *Handler) ExportExcel(c *gin.Context) {
	list, err := h.Service.ListAll(c.Query("search"))
	if err != nil {
		utils.Error(c, 500, "failed to export peserta")
		return
	}

	f := excelize.NewFile()
	defer f.Close()
	const sheet = "Peserta"
	f.SetSheetName("Sheet1", sheet)

	for i, col := range exportColumns {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, col)
	}
	for r, u := range list {
		row := exportRow(u)
		for i, v := range row {
			cell, _ := excelize.CoordinatesToCellName(i+1, r+2)
			f.SetCellValue(sheet, cell, v)
		}
	}

	filename := fmt.Sprintf("peserta_%s.xlsx", time.Now().Format("20060102_150405"))
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	if err := f.Write(c.Writer); err != nil {
		c.Status(http.StatusInternalServerError)
	}
}

// ktpZipInvalidChars covers characters most filesystems (notably Windows,
// which many admins extracting this ZIP will be on) reject in a filename.
var ktpZipInvalidChars = strings.NewReplacer("/", "-", "\\", "-", ":", "-", "*", "-", "?", "-", "\"", "'", "<", "-", ">", "-", "|", "-")

// ktpZipEntryName builds "[Nama Depan] [Nama Tengah] [Nama Belakang] - [NIK]"
// (the passport-style name fields, not the account's Name, which can differ
// and isn't what an admin cross-referencing physical KTP scans wants) with
// the original upload's extension. Falls back to the account Name when none
// of the three passport name fields are filled in.
func ktpZipEntryName(u users.User) string {
	parts := make([]string, 0, 3)
	for _, p := range []*string{u.FirstName, u.MiddleName, u.LastName} {
		if p != nil && strings.TrimSpace(*p) != "" {
			parts = append(parts, strings.TrimSpace(*p))
		}
	}
	name := strings.Join(parts, " ")
	if name == "" {
		name = u.Name
	}
	nik := strVal(u.KtpNumber)
	ext := filepath.Ext(*u.KtpFile)
	return ktpZipInvalidChars.Replace(fmt.Sprintf("%s - %s%s", name, nik, ext))
}

// ExportKtpZip godoc
// @Summary		Export every peserta's uploaded KTP file as one ZIP
// @Tags			peserta
// @Security		SessionCookie
// @Param			search	query	string	false	"Search by name/email/phone/KTP/passport"
// @Produce		application/zip
// @Success		200
// @Router			/peserta/export/ktp-zip [get]
func (h *Handler) ExportKtpZip(c *gin.Context) {
	list, err := h.Service.ListAll(c.Query("search"))
	if err != nil {
		utils.Error(c, 500, "failed to export peserta")
		return
	}

	filename := fmt.Sprintf("peserta_ktp_%s.zip", time.Now().Format("20060102_150405"))
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/zip")

	zw := zip.NewWriter(c.Writer)
	defer zw.Close()

	// NIK is unique per active peserta (idx_users_ktp_number_active), so the
	// generated entry name is unique too — dedupe anyway so a data hiccup
	// (a missing NIK on two rows, both falling back to "") produces distinct
	// files inside the archive instead of one silently overwriting another.
	usedNames := map[string]int{}
	for _, u := range list {
		if u.KtpFile == nil || *u.KtpFile == "" {
			continue
		}
		rc, err := h.Service.OpenKtpFile(*u.KtpFile)
		if err != nil {
			continue // file missing on disk — skip rather than fail the whole export
		}

		baseName := ktpZipEntryName(u)
		entryName := baseName
		if n := usedNames[baseName]; n > 0 {
			ext := filepath.Ext(baseName)
			base := strings.TrimSuffix(baseName, ext)
			entryName = fmt.Sprintf("%s (%d)%s", base, n, ext)
		}
		usedNames[baseName]++

		w, err := zw.Create(entryName)
		if err == nil {
			_, _ = io.Copy(w, rc)
		}
		rc.Close()
	}
}
