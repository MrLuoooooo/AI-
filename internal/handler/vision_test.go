package handler

import (
	"testing"

	"ai-vision-assistant/internal/component/frame"
	"ai-vision-assistant/internal/config"
	"ai-vision-assistant/internal/graph"
	"ai-vision-assistant/internal/service"

	"go.uber.org/zap"
)

func TestNewVisionHandler(t *testing.T) {
	svc := service.NewVisionService(
		&graph.VisionGraph{},
		&config.VisionConfig{MaxSessions: 10},
		zap.NewNop(),
	)
	sampler := frame.NewSampler(1000, 30, 640)
	h := NewVisionHandler(svc, sampler, zap.NewNop())
	if h == nil {
		t.Fatal("handler is nil")
	}
	if h.ActiveSessions() != 0 {
		t.Errorf("ActiveSessions = %d, want 0", h.ActiveSessions())
	}
}

func TestActiveSessions_Initial(t *testing.T) {
	svc := service.NewVisionService(&graph.VisionGraph{}, &config.VisionConfig{}, zap.NewNop())
	sampler := frame.NewSampler(1000, 30, 640)
	h := NewVisionHandler(svc, sampler, zap.NewNop())
	if h.ActiveSessions() != 0 {
		t.Errorf("ActiveSessions = %d, want 0", h.ActiveSessions())
	}
}
