package server

import (
	"net/http"
	"time"

	"ai-vision-assistant/internal/config"
	"ai-vision-assistant/internal/handler"
	"ai-vision-assistant/internal/server/middleware"
	"ai-vision-assistant/internal/server/version"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func NewRouter(cfg *config.Config, logger *zap.Logger, vh *handler.VisionHandler, rl *middleware.RateLimiter) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(middleware.CORS(cfg.Server.CORSOrigins))

	if cfg.Server.RateLimitRPS > 0 {
		engine.Use(rl.Middleware())
	}
	if cfg.Auth.APIKey != "" {
		engine.Use(middleware.Auth(&middleware.Config{APIKey: cfg.Auth.APIKey}))
	}
	engine.Use(middleware.Logger(logger))

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"version":   version.AppVersion,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	v1 := engine.Group("/api/v1")
	{
		v1.GET("/ws/vision", vh.HandleVision)
		v1.GET("/vision/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"active_sessions": vh.ActiveSessions()})
		})
	}

	return engine
}
