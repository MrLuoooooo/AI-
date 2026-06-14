package qianwenmodel

import (
	"encoding/json"
	"testing"

	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"
)

func TestNewQianwenModel(t *testing.T) {
	m := NewQianwenModel("sk-test", "qwen-vl-plus", zap.NewNop())
	if m == nil {
		t.Fatal("model is nil")
	}
}

func TestWrapTextContent(t *testing.T) {
	got := wrapTextContent("hello")
	want := `[{"text":"hello"}]`
	if got != want {
		t.Errorf("wrapTextContent = %q, want %q", got, want)
	}
}

func TestWrapTextContent_SpecialChars(t *testing.T) {
	got := wrapTextContent(`say "hi"`)
	// Double quotes should be escaped
	if !contains(got, `\"hi\"`) {
		t.Errorf("quotes not escaped: %s", got)
	}
}

func TestJsonEscape(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{`say "hi"`, `say \"hi\"`},
		{"line\nbreak", "line\\nbreak"},
	}
	for _, tt := range tests {
		got := jsonEscape(tt.input)
		if got != tt.want {
			t.Errorf("jsonEscape(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestBuildRequest_PlainText(t *testing.T) {
	m := &QianwenModel{
		apiKey: "sk-test",
		model:  "qwen-vl-plus",
		client: nil,
		logger: zap.NewNop(),
	}
	input := []*schema.Message{
		{Role: schema.System, Content: "你是一个助手"},
		{Role: schema.User, Content: "hello"},
	}
	body, err := m.buildRequest(input, false, nil)
	if err != nil {
		t.Fatal(err)
	}

	var req dsRequest
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatal(err)
	}
	if req.Model != "qwen-vl-plus" {
		t.Errorf("Model = %q", req.Model)
	}
	if len(req.Input.Messages) != 2 {
		t.Fatalf("len(Messages) = %d", len(req.Input.Messages))
	}
	// System message should be wrapped in text content array
	content := string(req.Input.Messages[0].Content)
	if !contains(content, "你是一个助手") {
		t.Errorf("system content = %s, missing prompt", content)
	}
}

func TestBuildRequest_MultimodalFormat(t *testing.T) {
	m := &QianwenModel{
		apiKey: "sk-test",
		model:  "qwen-vl-plus",
		client: nil,
		logger: zap.NewNop(),
	}
	// Content already in DashScope multimodal format
	input := []*schema.Message{
		{Role: schema.User, Content: `[{"text":"请描述"},{"image":"data:image/jpeg;base64,abc"}]`},
	}
	body, err := m.buildRequest(input, false, nil)
	if err != nil {
		t.Fatal(err)
	}

	var req dsRequest
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatal(err)
	}
	// Should pass through as-is (raw JSON)
	content := string(req.Input.Messages[0].Content)
	if !contains(content, "请描述") {
		t.Errorf("content = %s, missing text", content)
	}
	if !contains(content, "abc") {
		t.Errorf("content = %s, missing image base64", content)
	}
}

func TestBuildRequest_StreamMode(t *testing.T) {
	m := &QianwenModel{
		apiKey: "sk-test",
		model:  "qwen-vl-plus",
		client: nil,
		logger: zap.NewNop(),
	}
	body, err := m.buildRequest([]*schema.Message{
		{Role: schema.User, Content: "hello"},
	}, true, nil)
	if err != nil {
		t.Fatal(err)
	}

	var req dsRequest
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatal(err)
	}
	if !req.Parameters.IncrementalOutput {
		t.Error("IncrementalOutput should be true for stream")
	}
}

func TestBuildRequest_NonStream(t *testing.T) {
	m := &QianwenModel{
		apiKey: "sk-test",
		model:  "qwen-vl-plus",
		client: nil,
		logger: zap.NewNop(),
	}
	body, err := m.buildRequest([]*schema.Message{
		{Role: schema.User, Content: "hello"},
	}, false, nil)
	if err != nil {
		t.Fatal(err)
	}

	var req dsRequest
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatal(err)
	}
	if req.Parameters.IncrementalOutput {
		t.Error("IncrementalOutput should be false for non-stream")
	}
}

func TestBindTools(t *testing.T) {
	m := &QianwenModel{}
	if err := m.BindTools(nil); err != nil {
		t.Errorf("BindTools = %v, want nil", err)
	}
}

func TestQianwenModel_ImplementsChatModel(t *testing.T) {
	// Compile-time check via var _ line in source
	// Runtime check: just verify it can be created and assigned
	var m interface{} = NewQianwenModel("sk-test", "qwen-vl-plus", zap.NewNop())
	if m == nil {
		t.Fatal("model is nil")
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
