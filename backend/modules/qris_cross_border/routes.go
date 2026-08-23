package qris_cross_border

import "github.com/gin-gonic/gin"

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, sessionAuth gin.HandlerFunc, guard func(menuPath, action string) gin.HandlerFunc, uploadLimiter gin.HandlerFunc) {
	g := rg.Group("/qris-cross-border")

	// Participant-facing: any logged-in peserta can submit/view/delete their
	// own records, no qris-cross-border:* admin permission required — same
	// precedent as qr_gate's /scan and /my-scans.
	g.POST("/my", sessionAuth, uploadLimiter, h.CreateSelf)
	g.GET("/my", sessionAuth, h.MyList)
	g.DELETE("/my/:uuid", sessionAuth, h.DeleteSelf)

	g.GET("", guard("qris-cross-border", "view"), h.List)
	g.GET("/:uuid", guard("qris-cross-border", "view"), h.Get)
	g.POST("", guard("qris-cross-border", "create"), uploadLimiter, h.Create)
	g.PUT("/:uuid", guard("qris-cross-border", "edit"), uploadLimiter, h.Update)
	g.PUT("/:uuid/status", guard("qris-cross-border", "edit"), h.UpdateStatus)
	g.DELETE("/:uuid", guard("qris-cross-border", "delete"), h.Delete)
}
