package server

import (
	"testing"

	"ai-vision-assistant/internal/config"

	"go.uber.org/zap"
)

func TestProvideResolvedConfig_EnvFallback(t *testing.T) {
	cfg := &config.Config{
		ModelProvider: config.ModelProviderConfig{
			Cloud: config.CloudProviderConfig{
				ChatModel: "",
				BaseURL:   "",
				APIKey:    "",
				Type:      "dashscope",
			},
		},
	}
	rc := ProvideResolvedConfig(cfg, zap.NewNop())
	if rc.ChatModel == "" {
		t.Error("ChatModel should have fallback value")
	}
	if rc.BaseURL == "" {
		t.Error("BaseURL should have fallback value")
	}
	if rc.APIKey == "" {
		t.Error("APIKey should have fallback value")
	}
}

func TestProvideResolvedConfig_FromConfig(t *testing.T) {
	cfg := &config.Config{
		ModelProvider: config.ModelProviderConfig{
			Cloud: config.CloudProviderConfig{
				ChatModel: "test-model",
				BaseURL:   "http://test.url",
				APIKey:    "sk-custom",
				Type:      "dashscope",
			},
		},
	}
	rc := ProvideResolvedConfig(cfg, zap.NewNop())
	if rc.ChatModel != "test-model" {
		t.Errorf("ChatModel = %q, want test-model", rc.ChatModel)
	}
	if rc.BaseURL != "http://test.url" {
		t.Errorf("BaseURL = %q", rc.BaseURL)
	}
	if rc.APIKey != "sk-custom" {
		t.Errorf("APIKey = %q", rc.APIKey)
	}
}

func TestProvideResolvedConfig_ProviderField(t *testing.T) {
	cfg := &config.Config{
		ModelProvider: config.ModelProviderConfig{
			Cloud: config.CloudProviderConfig{
				Type: "dashscope",
			},
		},
	}
	rc := ProvideResolvedConfig(cfg, zap.NewNop())
	if rc.Provider != "cloud/dashscope" {
		t.Errorf("Provider = %q, want cloud/dashscope", rc.Provider)
	}
}

func TestProvideChatModel(t *testing.T) {
	rc := &ResolvedConfig{
		ChatModel: "qwen-vl-plus",
		BaseURL:   "http://localhost",
		APIKey:    "sk-test",
		Provider:  "cloud/dashscope",
	}
	cm := ProvideChatModel(rc, zap.NewNop())
	if cm == nil {
		t.Fatal("ChatModel is nil")
	}
}

func TestProvideFrameSampler(t *testing.T) {
	cfg := &config.Config{
		Vision: config.VisionConfig{
			FrameInterval: 500,
			FrameQuality:  50,
			MaxFrameWidth: 1280,
		},
	}
	s := ProvideFrameSampler(cfg)
	if s == nil {
		t.Fatal("Sampler is nil")
	}
}
