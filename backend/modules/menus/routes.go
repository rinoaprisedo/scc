package menus

import "github.com/gin-gonic/gin"

// RegisterRoutes wires up /menus/*. List is intentionally gated by
// sessionAuth only (not the "menus:view" permission) since every
// authenticated user needs it to build the sidebar nav tree, regardless of
// whether their role can access the Menus *management* page. Individual
// item visibility is still filtered client-side by the user's per-menu
// permissions; write operations remain permission-guarded below.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, sessionAuth gin.HandlerFunc, guard func(menuPath, action string) gin.HandlerFunc) {
	g := rg.Group("/menus")
	g.GET("", sessionAuth, h.List)
	g.GET("/:uuid", guard("menus", "view"), h.Get)
	g.POST("", guard("menus", "create"), h.Create)
	g.PUT("/:uuid", guard("menus", "edit"), h.Update)
	g.DELETE("/:uuid", guard("menus", "delete"), h.Delete)
	g.PUT("/reorder", guard("menus", "edit"), h.Reorder)
}
