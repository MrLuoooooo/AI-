package graph

import (
	"context"
	"testing"

	"ai-vision-assistant/internal/config"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type fakeChatModel struct {
	fn func(ctx context.Context, msgs []*schema.Message) (*schema.Message, error)
}

func (m *fakeChatModel) Generate(ctx context.Context, msgs []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	return m.fn(ctx, msgs)
}
func (m *fakeChatModel) Stream(ctx context.Context, msgs []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	msg, _ := m.fn(ctx, msgs)
	return schema.StreamReaderFromArray([]*schema.Message{msg}), nil
}
func (m *fakeChatModel) BindTools(tools []*schema.ToolInfo) error { return nil }

func TestJsonStr(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", `"hello"`},
		{`say "hi"`, `"say \"hi\""`},
		{"line\nbreak", `"line\nbreak"`},
	}
	for _, tt := range tests {
		got := jsonStr(tt.input)
		if got != tt.want {
			t.Errorf("jsonStr(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestBuildVisionContent_NoText(t *testing.T) {
	content := buildVisionContent("", "abc123")
	if content == "" {
		t.Error("empty content")
	}
	// Should contain the default prompt
	if !contains(content, "请描述画面内容") {
		t.Errorf("missing default prompt: %s", content)
	}
	// Should contain image
	if !contains(content, "abc123") {
		t.Errorf("missing frame base64: %s", content)
	}
}

func TestBuildVisionContent_WithText(t *testing.T) {
	content := buildVisionContent("这是什么", "abc123")
	if !contains(content, "这是什么") {
		t.Errorf("missing user text: %s", content)
	}
	if !contains(content, "用中文回答") {
		t.Errorf("missing Chinese hint: %s", content)
	}
}

func TestBuildVisionContent_NoFrame(t *testing.T) {
	content := buildVisionContent("你好", "")
	if contains(content, "image") {
		t.Errorf("should not contain image when frame is empty: %s", content)
	}
}

func TestBuildMessages(t *testing.T) {
	g := &VisionGraph{
		prompt: "你是测试助手",
		detail: "low",
	}

	input := &VisionInput{
		Transcript: "hello",
		FrameB64:   "fakebase64",
		History: []*schema.Message{
			{Role: schema.User, Content: "历史消息1"},
			{Role: schema.Assistant, Content: "历史回复1"},
		},
	}
	msgs, err := g.buildMessages(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 4 { // system + 2 history + 1 user
		t.Fatalf("len(msgs) = %d, want 4", len(msgs))
	}
	if msgs[0].Role != schema.System {
		t.Errorf("messages[0].Role = %v, want System", msgs[0].Role)
	}
	if msgs[0].Content != "你是测试助手" {
		t.Errorf("system prompt = %q", msgs[0].Content)
	}
}

func TestBuildMessages_HistoryTruncation(t *testing.T) {
	g := &VisionGraph{prompt: "test"}
	// Create 30 history messages
	history := make([]*schema.Message, 30)
	for i := range history {
		history[i] = &schema.Message{
			Role: schema.User, Content: "msg",
		}
	}
	input := &VisionInput{History: history}
	msgs, err := g.buildMessages(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	// system + 20 truncated history + 1 user = 22
	if len(msgs) != 22 {
		t.Errorf("len(msgs) = %d, want 22", len(msgs))
	}
}

func TestNewVisionGraph(t *testing.T) {
	cm := &fakeChatModel{
		fn: func(ctx context.Context, msgs []*schema.Message) (*schema.Message, error) {
			return &schema.Message{Role: schema.Assistant, Content: "看到了你的画面"}, nil
		},
	}
	cfg := &config.VisionConfig{
		SystemPrompt: "测试提示词",
		Detail:       "low",
	}
	g, err := NewVisionGraph(cm, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if g == nil {
		t.Fatal("graph is nil")
	}
	if g.prompt != "测试提示词" {
		t.Errorf("prompt = %q", g.prompt)
	}
}

func TestNewVisionGraph_Defaults(t *testing.T) {
	cm := &fakeChatModel{
		fn: func(ctx context.Context, msgs []*schema.Message) (*schema.Message, error) {
			return &schema.Message{Role: schema.Assistant, Content: "ok"}, nil
		},
	}
	g, err := NewVisionGraph(cm, &config.VisionConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if g.prompt == "" {
		t.Error("prompt should have default value")
	}
	if len(g.prompt) < 10 {
		t.Errorf("prompt too short: %q", g.prompt)
	}
}

func TestVisionGraph_Invoke(t *testing.T) {
	cm := &fakeChatModel{
		fn: func(ctx context.Context, msgs []*schema.Message) (*schema.Message, error) {
			return &schema.Message{Role: schema.Assistant, Content: "回复"}, nil
		},
	}
	g, err := NewVisionGraph(cm, &config.VisionConfig{})
	if err != nil {
		t.Fatal(err)
	}
	msg, err := g.Invoke(context.Background(), &VisionInput{Transcript: "hi"})
	if err != nil {
		t.Fatal(err)
	}
	if msg.Content != "回复" {
		t.Errorf("content = %q", msg.Content)
	}
}

func TestVisionGraph_Stream(t *testing.T) {
	cm := &fakeChatModel{
		fn: func(ctx context.Context, msgs []*schema.Message) (*schema.Message, error) {
			return &schema.Message{Role: schema.Assistant, Content: "流式回复"}, nil
		},
	}
	g, err := NewVisionGraph(cm, &config.VisionConfig{})
	if err != nil {
		t.Fatal(err)
	}
	sr, err := g.Stream(context.Background(), &VisionInput{Transcript: "hi"})
	if err != nil {
		t.Fatal(err)
	}
	msg, err := sr.Recv()
	if err != nil {
		t.Fatal(err)
	}
	if msg.Content != "流式回复" {
		t.Errorf("content = %q", msg.Content)
	}
}

func TestVisionGraph_UpdatePrompt(t *testing.T) {
	cm := &fakeChatModel{
		fn: func(ctx context.Context, msgs []*schema.Message) (*schema.Message, error) {
			return &schema.Message{Role: schema.Assistant, Content: "ok"}, nil
		},
	}
	g, err := NewVisionGraph(cm, &config.VisionConfig{SystemPrompt: "old"})
	if err != nil {
		t.Fatal(err)
	}
	g.UpdatePrompt("new")
	if g.Prompt() != "new" {
		t.Errorf("Prompt = %q, want new", g.Prompt())
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
