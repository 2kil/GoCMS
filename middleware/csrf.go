package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const csrfSessionKey = "csrf_token"
const csrfCookieName = "cms_csrf"

func RequireCSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		token, _ := session.Get(csrfSessionKey).(string)
		if token == "" {
			var err error
			token, err = newCSRFToken()
			if err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			session.Set(csrfSessionKey, token)
			if err := session.Save(); err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}

		c.SetCookie(csrfCookieName, token, 0, "/adm1n", "", c.Request.TLS != nil, false)

		if isMutatingMethod(c.Request.Method) {
			submitted := c.GetHeader("X-CSRF-Token")
			if submitted == "" {
				submitted = c.PostForm("_csrf")
			}
			if submitted == "" || submitted != token {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
		}

		c.Next()
	}
}

func isMutatingMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func newCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
