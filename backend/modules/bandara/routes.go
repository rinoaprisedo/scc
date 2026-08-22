package bandara

import "github.com/gin-gonic/gin"

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, sessionAuth gin.HandlerFunc, guard func(menuPath, action string) gin.HandlerFunc) {
	g := rg.Group("/bandara")
	// Self-service dropdown data for the peserta website — logged in, but
	// without the admin "bandara:view" permission.
	g.GET("/options", sessionAuth, h.Options)
	g.GET("", guard("bandara", "view"), h.List)
	g.GET("/:uuid", guard("bandara", "view"), h.Get)
	g.POST("", guard("bandara", "create"), h.Create)
	g.PUT("/:uuid", guard("bandara", "edit"), h.Update)
	g.DELETE("/:uuid", guard("bandara", "delete"), h.Delete)
}
