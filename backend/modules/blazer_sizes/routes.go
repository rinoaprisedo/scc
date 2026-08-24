package blazer_sizes

import "github.com/gin-gonic/gin"

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, sessionAuth gin.HandlerFunc, guard func(menuPath, action string) gin.HandlerFunc) {
	g := rg.Group("/blazer-sizes")
	// Self-service dropdown data for the peserta website — logged in, but
	// without the admin "blazer-sizes:view" permission.
	g.GET("/options", sessionAuth, h.Options)
	g.GET("", guard("blazer-sizes", "view"), h.List)
	g.GET("/:uuid", guard("blazer-sizes", "view"), h.Get)
	g.POST("", guard("blazer-sizes", "create"), h.Create)
	g.PUT("/:uuid", guard("blazer-sizes", "edit"), h.Update)
	g.DELETE("/:uuid", guard("blazer-sizes", "delete"), h.Delete)
	g.PUT("/reorder", guard("blazer-sizes", "edit"), h.Reorder)
}
