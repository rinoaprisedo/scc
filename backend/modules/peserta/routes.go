package peserta

import "github.com/gin-gonic/gin"

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, sessionAuth gin.HandlerFunc, guard func(menuPath, action string) gin.HandlerFunc, uploadLimiter gin.HandlerFunc) {
	g := rg.Group("/peserta")

	// Self-service — a Peserta acting on their own record via the website,
	// not the admin-permission-gated routes below. Session auth only.
	g.PUT("/me", sessionAuth, h.UpdateMe)
	g.PUT("/me/attendance", sessionAuth, h.SetMyAttendance)
	g.POST("/me/ktp", sessionAuth, uploadLimiter, h.UploadMyKtp)
	g.POST("/me/passport", sessionAuth, uploadLimiter, h.UploadMyPassport)

	g.GET("", guard("peserta", "view"), h.List)
	g.GET("/export/csv", guard("peserta", "view"), h.ExportCSV)
	g.GET("/export/excel", guard("peserta", "view"), h.ExportExcel)
	g.GET("/export/ktp-zip", guard("peserta", "view"), h.ExportKtpZip)
	g.GET("/import/template", guard("peserta", "create"), h.ImportTemplate)
	g.POST("/import/validate", guard("peserta", "create"), uploadLimiter, h.ValidateImport)
	g.POST("/import", guard("peserta", "create"), uploadLimiter, h.ImportExcel)
	g.GET("/:uuid", guard("peserta", "view"), h.Get)
	g.GET("/:uuid/points", guard("peserta", "view"), h.PointHistory)
	g.POST("", guard("peserta", "create"), h.Create)
	g.PUT("/:uuid", guard("peserta", "edit"), h.Update)
	g.DELETE("/:uuid", guard("peserta", "delete"), h.Delete)
	g.PUT("/:uuid/status", guard("peserta", "edit"), h.UpdateStatus)
	g.POST("/:uuid/ktp", guard("peserta", "edit"), uploadLimiter, h.UploadKtp)
	g.POST("/:uuid/passport", guard("peserta", "edit"), uploadLimiter, h.UploadPassport)
}
