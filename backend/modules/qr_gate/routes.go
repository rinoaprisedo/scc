package qr_gate

import "github.com/gin-gonic/gin"

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, sessionAuth gin.HandlerFunc, guard func(menuPath, action string) gin.HandlerFunc) {
	g := rg.Group("/qr-gate")
	// Participant-facing: any logged-in peserta can scan/view their own
	// history, no qr-gate:* admin permission required.
	g.POST("/scan", sessionAuth, h.Scan)
	g.GET("/my-scans", sessionAuth, h.MyScans)

	g.GET("", guard("qr-gate", "view"), h.List)
	g.GET("/:uuid", guard("qr-gate", "view"), h.Get)
	g.GET("/:uuid/scans", guard("qr-gate", "view"), h.GateScans)
	g.POST("", guard("qr-gate", "create"), h.Create)
	g.PUT("/:uuid", guard("qr-gate", "edit"), h.Update)
	g.DELETE("/:uuid", guard("qr-gate", "delete"), h.Delete)
}
