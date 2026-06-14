// @title           AI Vision Assistant
// @version         1.0.0
// @description     摄像头 + 麦克风 = AI 视觉对话
// @host            localhost:8080
// @BasePath        /api/v1
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"ai-vision-assistant/internal/config"
	"ai-vision-assistant/internal/server"
	"ai-vision-assistant/internal/server/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	app := fx.New(
		server.Module,
		fx.Invoke(startServer),
	)
	app.Run()
}

func startServer(
	lc fx.Lifecycle,
	cfg *config.Config,
	logger *zap.Logger,
	engine *gin.Engine,
	rateLimiter *middleware.RateLimiter,
) {
	gin.SetMode(cfg.Server.Mode)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: engine,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("AI Vision Assistant starting", zap.String("addr", srv.Addr))
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Error("Server failed", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Shutting down server")
			rateLimiter.Stop()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := srv.Shutdown(shutdownCtx); err != nil {
				logger.Error("HTTP server shutdown error", zap.Error(err))
			}
			logger.Info("Server stopped")
			return nil
		},
	})
}
