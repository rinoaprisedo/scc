package peserta

import (
	"encoding/csv"
	"fmt"
	"net/http"
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
		strVal(u.DietaryRestrictionOther),
		strVal(u.PhoneNumber),
		strVal(u.PassportNumber),
		dateVal(u.PassportExpiry),
		strVal(u.JacketSize),
		strVal(u.PoloSize),
		strVal(u.NomorMeja),
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
