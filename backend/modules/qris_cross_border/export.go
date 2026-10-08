package qris_cross_border

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"baseadmin/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

var exportColumns = []string{"Nama Peserta", "NIK", "Nominal (MYR)", "Nominal (IDR)"}

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

	// Display-only currency formats — cells stay numeric so the sheet can
	// still be summed/sorted. MYR only shows decimals when there are any, so
	// whole amounts read "RM 1,000" rather than "RM 1,000.00".
	idrFmt, myrFmt, myrDecFmt := `"Rp "#,##0`, `"RM "#,##0`, `"RM "#,##0.00`
	idrStyle, _ := f.NewStyle(&excelize.Style{CustomNumFmt: &idrFmt})
	myrStyle, _ := f.NewStyle(&excelize.Style{CustomNumFmt: &myrFmt})
	myrDecStyle, _ := f.NewStyle(&excelize.Style{CustomNumFmt: &myrDecFmt})

	for i, col := range exportColumns {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, col)
	}
	for i, row := range list {
		r := i + 2
		name, nik := "", ""
		if row.Peserta != nil {
			name = row.Peserta.Name
			if row.Peserta.KtpNumber != nil {
				nik = *row.Peserta.KtpNumber
			}
		}
		nameCell, _ := excelize.CoordinatesToCellName(1, r)
		nikCell, _ := excelize.CoordinatesToCellName(2, r)
		myrCell, _ := excelize.CoordinatesToCellName(3, r)
		idrCell, _ := excelize.CoordinatesToCellName(4, r)
		f.SetCellValue(sheet, nameCell, name)
		f.SetCellValue(sheet, nikCell, nik)
		f.SetCellValue(sheet, myrCell, row.NominalAsing)
		f.SetCellValue(sheet, idrCell, row.NominalRupiah)
		if row.NominalAsing == math.Trunc(row.NominalAsing) {
			f.SetCellStyle(sheet, myrCell, myrCell, myrStyle)
		} else {
			f.SetCellStyle(sheet, myrCell, myrCell, myrDecStyle)
		}
		f.SetCellStyle(sheet, idrCell, idrCell, idrStyle)
	}

	filename := fmt.Sprintf("qris_cross_border_%s.xlsx", time.Now().Format("20060102_150405"))
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	if err := f.Write(c.Writer); err != nil {
		c.Status(http.StatusInternalServerError)
	}
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
