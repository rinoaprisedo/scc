package kota_asal

import "github.com/gin-gonic/gin"

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, sessionAuth gin.HandlerFunc, guard func(menuPath, action string) gin.HandlerFunc) {
	g := rg.Group("/kota-asal")
	// Self-service dropdown data for the peserta website — logged in, but
	// without the admin "kota-asal:view" permission.
	g.GET("/options", sessionAuth, h.Options)
	g.GET("", guard("kota-asal", "view"), h.List)
	g.GET("/:uuid", guard("kota-asal", "view"), h.Get)
	g.POST("", guard("kota-asal", "create"), h.Create)
	g.PUT("/:uuid", guard("kota-asal", "edit"), h.Update)
	g.DELETE("/:uuid", guard("kota-asal", "delete"), h.Delete)
}
