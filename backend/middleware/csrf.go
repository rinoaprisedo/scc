package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

const csrfCookieName = "csrf_token"
const csrfHeaderName = "X-CSRF-Token"

// CSRF implements the double-submit-cookie pattern: a token is issued as a
// non-httpOnly cookie and must be echoed back in a request header on any
// state-changing request. domain lets the cookie be shared across
// subdomains (e.g. ".example.com") when the frontend(s) and API live on
// different subdomains of the same site — an empty domain falls back to a
// host-only cookie, which is fine when everything shares one hostname (dev).
func CSRF(secure bool, domain string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(csrfCookieName)
		if err != nil || token == "" {
			token = generateCSRFToken()
			c.SetCookie(csrfCookieName, token, 86400, "/", domain, secure, false)
		}

		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		header := c.GetHeader(csrfHeaderName)
		if header == "" || subtle.ConstantTimeCompare([]byte(header), []byte(token)) != 1 {
			c.AbortWithStatusJSON(403, gin.H{"success": false, "message": "invalid or missing CSRF token"})
			return
		}
		c.Next()
	}
}

func generateCSRFToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
