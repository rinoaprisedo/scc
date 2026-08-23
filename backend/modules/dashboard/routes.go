package dashboard

import "github.com/gin-gonic/gin"

// RegisterRoutes wires the dashboard's read-only summary endpoint. Gated by
// session auth only — no menuPath guard — since the "Dashboard" menu is
// seeded at path "/" (the app's landing page for any authenticated admin),
// not a permission-checkable path like "peserta" or "qr-gate".
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, sessionAuth gin.HandlerFunc) {
	g := rg.Group("/dashboard")
	g.GET("/summary", sessionAuth, h.Summary)
}
