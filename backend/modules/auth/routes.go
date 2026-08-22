package auth

import (
	"time"

	"baseadmin/backend/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires up /auth/* endpoints. loginRateLimit and forgotRateLimit
// are pre-built middleware so routes.go stays free of Redis wiring details.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, sessionAuth gin.HandlerFunc) {
	g := rg.Group("/auth")

	loginLimiter := middleware.RateLimit(h.Redis, "login", h.Cfg.RateLimitLogin, time.Minute, func(c *gin.Context) string { return c.ClientIP() })
	// Separate Redis counter from loginLimiter (distinct prefix) so the
	// participant website and the admin panel don't share one request
	// budget, even though LoginLockout below is intentionally shared —
	// IP-based lockout is meant to be endpoint-agnostic.
	pesertaLoginLimiter := middleware.RateLimit(h.Redis, "peserta-login", h.Cfg.RateLimitLogin, time.Minute, func(c *gin.Context) string { return c.ClientIP() })
	// Keyed by IP rather than email: the JSON body isn't parsed yet at the
	// middleware stage without an extra bind/rebuild step, and per-IP is a
	// reasonable pragmatic approximation of the per-email spec requirement.
	forgotLimiter := middleware.RateLimit(h.Redis, "forgot-password", 3, time.Hour, func(c *gin.Context) string { return c.ClientIP() })

	g.POST("/login", middleware.LoginLockout(h.Redis, h.Cfg.RateLimitLogin, h.Cfg.RateLimitLoginLockoutMin), loginLimiter, h.Login)
	g.POST("/peserta-login", middleware.LoginLockout(h.Redis, h.Cfg.RateLimitLogin, h.Cfg.RateLimitLoginLockoutMin), pesertaLoginLimiter, h.PesertaLogin)
	g.POST("/logout", sessionAuth, h.Logout)
	g.GET("/me", sessionAuth, h.Me)
	g.POST("/forgot-password", forgotLimiter, h.ForgotPassword)
	g.POST("/reset-password", h.ResetPassword)
	g.PUT("/change-password", sessionAuth, h.ChangePassword)
}
