package sliders

import "github.com/gin-gonic/gin"

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, sessionAuth gin.HandlerFunc, guard func(menuPath, action string) gin.HandlerFunc, uploadLimiter gin.HandlerFunc) {
	g := rg.Group("/sliders")
	// Self-service carousel data for the peserta website — logged in, but
	// without the admin "sliders:view" permission. Same precedent as
	// blazer_sizes' /options.
	g.GET("/active", sessionAuth, h.Active)
	g.GET("", guard("sliders", "view"), h.List)
	g.GET("/:uuid", guard("sliders", "view"), h.Get)
	g.POST("", guard("sliders", "create"), uploadLimiter, h.Create)
	g.PUT("/:uuid", guard("sliders", "edit"), uploadLimiter, h.Update)
	g.DELETE("/:uuid", guard("sliders", "delete"), h.Delete)
	g.PUT("/reorder", guard("sliders", "edit"), h.Reorder)
}
