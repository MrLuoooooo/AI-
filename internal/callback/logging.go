package callback

import (
	"context"
	"time"

	"github.com/cloudwego/eino/callbacks"
	"go.uber.org/zap"
)

// LoggingCallback Eino 全局回调：开始/结束/报错时打日志。
type LoggingCallback struct {
	logger *zap.Logger
}

// NewLoggingCallback 建日志回调，启动时注册到 Eino 全局。
func NewLoggingCallback(logger *zap.Logger) callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			logger.Info("component start",
				zap.String("name", info.Name),
				zap.String("type", info.Type),
				zap.Any("component", info.Component),
				zap.Any("_start_time", time.Now()),
			)
			return context.WithValue(ctx, ctxKeyStartTime, time.Now())
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			startTime, _ := ctx.Value(ctxKeyStartTime).(time.Time)
			duration := time.Since(startTime)
			logger.Info("component end",
				zap.String("name", info.Name),
				zap.String("type", info.Type),
				zap.Duration("duration", duration),
				zap.Any("output_summary", summarizeOutput(output)),
			)
			return ctx
		}).
		OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			logger.Error("component error",
				zap.String("name", info.Name),
				zap.String("type", info.Type),
				zap.Error(err),
			)
			return ctx
		}).
		Build()
}

type ctxKey string

const ctxKeyStartTime ctxKey = "start_time"

// summarizeOutput 把组件输出截短到 200 字，方便日志查看。
func summarizeOutput(output any) string {
	if output == nil {
		return ""
	}
	if s, ok := output.(string); ok {
		if len(s) > 200 {
			return s[:200] + "..."
		}
		return s
	}
	return ""
}
