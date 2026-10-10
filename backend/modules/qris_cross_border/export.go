package qris_cross_border

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"baseadmin/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

var exportColumns = []string{"Nama Peserta", "NIK", "Deskripsi", "Region", "Cabang", "Position", "Nominal (MYR)", "Nominal (IDR)", "Created At (WIB)"}

// ExportExcel godoc
// @Summary		Export QRIS cross border submissions as Excel
// @Tags			qris-cross-border
// @Security		SessionCookie
// @Param			search			query	string	false	"Search by peserta name/NIK/merchant/no. referensi"
// @Param			trx_status		query	string	false	"Filter by transaction status"
// @Param			peserta_uuid	query	string	false	"Filter by peserta"
// @Produce		application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Success		200
// @Router			/qris-cross-border/export/excel [get]
func (h *Handler) ExportExcel(c *gin.Context) {
	// Export is a payout/recap sheet, so it's always approved-only regardless
	// of whatever status filter the list page currently has applied.
	list, err := h.Service.ListAll(c.Query("search"), StatusApproved, c.Query("trx_status"), c.Query("peserta_uuid"))
	if err != nil {
		utils.Error(c, 500, "failed to export qris cross border")
		return
	}

	f := excelize.NewFile()
	defer f.Close()
	const sheet = "Qris Cross Border"
	f.SetSheetName("Sheet1", sheet)

	// Nominal cells are written as text (not numbers) on request, so the
	// sheet shows exactly the formatted string regardless of locale. "@" is
	// Excel's text format, which stops it re-parsing "1,000" back to a number.
	textStyle, _ := f.NewStyle(&excelize.Style{NumFmt: 49})

	for i, col := range exportColumns {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, col)
	}
	for i, row := range list {
		r := i + 2
		peserta := []string{"", "", "", "", "", ""}
		if p := row.Peserta; p != nil {
			peserta = []string{p.Name, deref(p.KtpNumber), deref(p.Description), deref(p.Region), deref(p.Cabang), deref(p.Position)}
		}
		for j, v := range peserta {
			cell, _ := excelize.CoordinatesToCellName(j+1, r)
			f.SetCellValue(sheet, cell, v)
		}
		myrCell, _ := excelize.CoordinatesToCellName(len(peserta)+1, r)
		idrCell, _ := excelize.CoordinatesToCellName(len(peserta)+2, r)
		f.SetCellStr(sheet, myrCell, formatNominal(row.NominalAsing))
		f.SetCellStr(sheet, idrCell, formatNominal(row.NominalRupiah))
		f.SetCellStyle(sheet, myrCell, idrCell, textStyle)
		createdCell, _ := excelize.CoordinatesToCellName(len(peserta)+3, r)
		f.SetCellStr(sheet, createdCell, row.CreatedAt.In(wib).Format("2006-01-02 15:04:05"))
	}

	filename := fmt.Sprintf("qris_cross_border_%s.xlsx", time.Now().Format("20060102_150405"))
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	if err := f.Write(c.Writer); err != nil {
		c.Status(http.StatusInternalServerError)
	}
}

// Fixed +7 offset rather than time.LoadLocation — same reason as
// peserta's export: the alpine production image has no tzdata.
var wib = time.FixedZone("WIB", 7*60*60)

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// formatNominal renders an amount with comma thousand separators, and two
// decimals only when there's a fractional part: 1000 → "1,000",
// 1234.5 → "1,234.50".
func formatNominal(v float64) string {
	str := strconv.FormatFloat(math.Abs(v), 'f', 2, 64)
	intPart, frac := str[:len(str)-3], str[len(str)-2:]

	var b strings.Builder
	if v < 0 {
		b.WriteByte('-')
	}
	for i, d := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(d)
	}
	if frac != "00" {
		b.WriteString("." + frac)
	}
	return b.String()
}

// ExportCount godoc
// @Summary		Count distinct peserta included in / excluded from the Excel export
// @Tags			qris-cross-border
// @Security		SessionCookie
// @Param			search			query	string	false	"Search by peserta name/NIK/merchant/no. referensi"
// @Param			trx_status		query	string	false	"Filter by transaction status"
// @Param			peserta_uuid	query	string	false	"Filter by peserta"
// @Success		200	{object}	utils.Response
// @Router			/qris-cross-border/export/count [get]
func (h *Handler) ExportCount(c *gin.Context) {
	included, excluded, err := h.Service.CountExportPeserta(c.Query("search"), StatusApproved, c.Query("trx_status"), c.Query("peserta_uuid"))
	if err != nil {
		utils.Error(c, 500, "failed to count qris cross border export")
		return
	}
	utils.Success(c, 200, "ok", gin.H{"peserta_count": included, "excluded_count": excluded})
}
