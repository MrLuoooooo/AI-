package server

import (
	"ai-vision-assistant/internal/callback"
	"ai-vision-assistant/internal/config"
	"ai-vision-assistant/internal/logger"
	"ai-vision-assistant/internal/server/middleware"

	"github.com/cloudwego/eino/callbacks"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func ProvideConfig() (*config.Config, error) { return config.Load() }

func ProvideLogger(cfg *config.Config) (*zap.Logger, error) {
	return logger.NewLogger(&logger.Config{
		Level:      cfg.Log.Level,
		FilePath:   cfg.Log.FilePath,
		MaxSize:    cfg.Log.MaxSize,
		MaxBackups: cfg.Log.MaxBackups,
		MaxAge:     cfg.Log.MaxAge,
		Compress:   cfg.Log.Compress,
	}), nil
}

func ProvideRateLimiter(cfg *config.Config) *middleware.RateLimiter {
	rps := cfg.Server.RateLimitRPS
	if rps <= 0 {
		rps = 0
	}
	return middleware.NewRateLimiter(rps, rps*2)
}

func ProvideRouter(cfg *config.Config, logger *zap.Logger, rl *middleware.RateLimiter) *gin.Engine {
	return NewRouter(cfg, logger, rl)
}

var Module = fx.Module("vision-assistant",
	fx.Provide(
		ProvideConfig,
		ProvideLogger,
		ProvideRateLimiter,
		ProvideRouter,
	),
	fx.Invoke(func(logger *zap.Logger) {
		hdl := callback.NewLoggingCallback(logger)
		callbacks.AppendGlobalHandlers(hdl)
	}),
)
