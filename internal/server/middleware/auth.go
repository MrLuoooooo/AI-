package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"ai-vision-assistant/internal/model"
)

// UserContextKey is the key used to store authenticated user info in the request context.
const UserContextKey = "auth_user"

// Auth API Key 鉴权中间件。配了 key 就拦截，没配全放行。
func Auth(cfg *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := "anonymous"
		auth := c.GetHeader("Authorization")

		if auth != "" && strings.HasPrefix(auth, "Bearer ") {
			token := strings.TrimPrefix(auth, "Bearer ")

			if cfg.APIKey != "" {
				if token == cfg.APIKey {
					user = "authenticated"
				} else {
					// Production: reject mismatched keys.
					c.Set(UserContextKey, "invalid_token")
					c.AbortWithStatusJSON(http.StatusUnauthorized, model.APIEnvelope{
						Code:    401,
						Message: "unauthorized",
					})
					return
				}
			} else if token != "" {
				// Dev mode: accept any non-empty token.
				user = "authenticated"
			}
		} else if cfg.APIKey != "" {
			// Production: no Authorization header → 401.
			c.Set(UserContextKey, "anonymous")
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.APIEnvelope{
				Code:    401,
				Message: "unauthorized",
			})
			return
		}

		c.Set(UserContextKey, user)
		c.Next()
	}
}

// Config mirrors the authentication section of the application config.
type Config struct {
	APIKey string
}
