package qris_cross_border

import (
	"fmt"
	"net/http"
	"time"

	"baseadmin/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

var exportColumns = []string{"Nama Peserta", "NIK", "Nominal (IDR)"}

// ExportExcel godoc
// @Summary		Export QRIS cross border submissions as Excel
// @Tags			qris-cross-border
// @Security		SessionCookie
// @Param			search			query	string	false	"Search by peserta name/NIK/merchant/no. referensi"
// @Param			status			query	string	false	"Filter by status"
// @Param			trx_status		query	string	false	"Filter by transaction status"
// @Param			peserta_uuid	query	string	false	"Filter by peserta"
// @Produce		application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Success		200
// @Router			/qris-cross-border/export/excel [get]
func (h *Handler) ExportExcel(c *gin.Context) {
	list, err := h.Service.ListAll(c.Query("search"), c.Query("status"), c.Query("trx_status"), c.Query("peserta_uuid"))
	if err != nil {
		utils.Error(c, 500, "failed to export qris cross border")
		return
	}

	f := excelize.NewFile()
	defer f.Close()
	const sheet = "Qris Cross Border"
	f.SetSheetName("Sheet1", sheet)

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
		nominalCell, _ := excelize.CoordinatesToCellName(3, r)
		f.SetCellValue(sheet, nameCell, name)
		f.SetCellValue(sheet, nikCell, nik)
		f.SetCellValue(sheet, nominalCell, row.NominalRupiah)
	}

	filename := fmt.Sprintf("qris_cross_border_%s.xlsx", time.Now().Format("20060102_150405"))
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	if err := f.Write(c.Writer); err != nil {
		c.Status(http.StatusInternalServerError)
	}
}
