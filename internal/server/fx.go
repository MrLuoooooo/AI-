package server

import (
	"os"

	"ai-vision-assistant/internal/callback"
	"ai-vision-assistant/internal/component/frame"
	"ai-vision-assistant/internal/component/qianwenmodel"
	"ai-vision-assistant/internal/config"
	"ai-vision-assistant/internal/graph"
	"ai-vision-assistant/internal/handler"
	"ai-vision-assistant/internal/logger"
	"ai-vision-assistant/internal/server/middleware"
	"ai-vision-assistant/internal/service"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type ResolvedConfig struct {
	ChatModel, BaseURL, APIKey, Provider string
}

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

func ProvideResolvedConfig(cfg *config.Config, logger *zap.Logger) *ResolvedConfig {
	c := cfg.ModelProvider.Cloud
	cm := c.ChatModel
	if cm == "" {
		cm = os.Getenv("LLM_CHAT_MODEL")
	}
	if cm == "" {
		cm = "qwen-vl-plus"
	}
	bu := c.BaseURL
	if bu == "" {
		bu = os.Getenv("LLM_BASE_URL")
	}
	if bu == "" {
		bu = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	}
	ak := c.APIKey
	if ak == "" {
		ak = os.Getenv("GOAGENT_MODEL_PROVIDER_CLOUD_API_KEY")
	}
	if ak == "" {
		ak = "sk-placeholder"
	}
	logger.Info("model resolved", zap.String("chat_model", cm), zap.String("base_url", bu))
	return &ResolvedConfig{
		ChatModel: cm,
		BaseURL:   bu,
		APIKey:    ak,
		Provider:  "cloud/" + c.Type,
	}
}

func ProvideChatModel(rc *ResolvedConfig, logger *zap.Logger) model.ChatModel {
	return qianwenmodel.NewQianwenModel(rc.APIKey, rc.ChatModel, logger)
}

func ProvideFrameSampler(cfg *config.Config) *frame.Sampler {
	return frame.NewSampler(cfg.Vision.FrameInterval, cfg.Vision.FrameQuality, cfg.Vision.MaxFrameWidth)
}

func ProvideVisionGraph(cm model.ChatModel, cfg *config.Config) (*graph.VisionGraph, error) {
	return graph.NewVisionGraph(cm, &cfg.Vision)
}

func ProvideVisionService(vg *graph.VisionGraph, cfg *config.Config, logger *zap.Logger) *service.VisionService {
	return service.NewVisionService(vg, &cfg.Vision, logger)
}

func ProvideVisionHandler(svc *service.VisionService, sampler *frame.Sampler, logger *zap.Logger) *handler.VisionHandler {
	return handler.NewVisionHandler(svc, sampler, logger)
}

func ProvideRateLimiter(cfg *config.Config) *middleware.RateLimiter {
	rps := cfg.Server.RateLimitRPS
	if rps <= 0 {
		rps = 0
	}
	return middleware.NewRateLimiter(rps, rps*2)
}

func ProvideRouter(cfg *config.Config, logger *zap.Logger, vh *handler.VisionHandler, rl *middleware.RateLimiter) *gin.Engine {
	return NewRouter(cfg, logger, vh, rl)
}

var Module = fx.Module("vision-assistant",
	fx.Provide(
		ProvideConfig,
		ProvideLogger,
		ProvideResolvedConfig,
		ProvideChatModel,
		ProvideFrameSampler,
		ProvideVisionGraph,
		ProvideVisionService,
		ProvideVisionHandler,
		ProvideRateLimiter,
		ProvideRouter,
	),
	fx.Invoke(func(logger *zap.Logger) {
		hdl := callback.NewLoggingCallback(logger)
		callbacks.AppendGlobalHandlers(hdl)
	}),
)
