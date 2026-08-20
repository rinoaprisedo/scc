package menu_sections

import "github.com/gin-gonic/gin"

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, guard func(menuPath, action string) gin.HandlerFunc) {
	g := rg.Group("/menu-sections")
	g.GET("", guard("menu-sections", "view"), h.List)
	g.GET("/:uuid", guard("menu-sections", "view"), h.Get)
	g.POST("", guard("menu-sections", "create"), h.Create)
	g.PUT("/:uuid", guard("menu-sections", "edit"), h.Update)
	g.DELETE("/:uuid", guard("menu-sections", "delete"), h.Delete)
	g.PUT("/reorder", guard("menu-sections", "edit"), h.Reorder)
}
