package service

import (
	"context"
	"testing"

	"ai-vision-assistant/internal/config"
	"ai-vision-assistant/internal/graph"

	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"
)

// mockGraph 用于测试 VisionService。
type mockGraph struct {
	invokeResult string
	invokeErr    error
}

func (m *mockGraph) Invoke(ctx context.Context, input *graph.VisionInput) (*schema.Message, error) {
	if m.invokeErr != nil {
		return nil, m.invokeErr
	}
	return &schema.Message{Role: schema.Assistant, Content: m.invokeResult}, nil
}

func (m *mockGraph) Stream(ctx context.Context, input *graph.VisionInput) (*schema.StreamReader[*schema.Message], error) {
	msg, err := m.Invoke(ctx, input)
	if err != nil {
		return nil, err
	}
	return schema.StreamReaderFromArray([]*schema.Message{msg}), nil
}

func newTestService() *VisionService {
	return NewVisionService(
		&graph.VisionGraph{}, // placeholder, replaced in tests that need graph
		&config.VisionConfig{MaxSessions: 10},
		zap.NewNop(),
	)
}

func TestCreateSession(t *testing.T) {
	svc := newTestService()
	id := svc.CreateSession()
	if id == "" {
		t.Error("session id should not be empty")
	}
	if svc.SessionCount() != 1 {
		t.Errorf("SessionCount = %d, want 1", svc.SessionCount())
	}
}

func TestCreateSession_UniqueIDs(t *testing.T) {
	svc := newTestService()
	id1 := svc.CreateSession()
	id2 := svc.CreateSession()
	if id1 == id2 {
		t.Error("session IDs should be unique")
	}
}

func TestCheckLimit_UnderLimit(t *testing.T) {
	svc := newTestService()
	if err := svc.CheckLimit(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCheckLimit_Reached(t *testing.T) {
	svc := NewVisionService(nil, &config.VisionConfig{MaxSessions: 1}, zap.NewNop())
	svc.CreateSession()
	err := svc.CheckLimit()
	if err == nil {
		t.Error("expected error when limit reached")
	}
}

func TestSetAndGetLatestFrame(t *testing.T) {
	svc := newTestService()
	id := svc.CreateSession()
	svc.SetLatestFrame(id, "base64frame")
	got := svc.GetLatestFrame(id)
	if got != "base64frame" {
		t.Errorf("GetLatestFrame = %q, want base64frame", got)
	}
}

func TestGetLatestFrame_Missing(t *testing.T) {
	svc := newTestService()
	got := svc.GetLatestFrame("nonexistent")
	if got != "" {
		t.Errorf("GetLatestFrame = %q, want empty", got)
	}
}

func TestCloseSession(t *testing.T) {
	svc := newTestService()
	id := svc.CreateSession()
	svc.CloseSession(id)
	if svc.SessionCount() != 0 {
		t.Errorf("SessionCount = %d, want 0", svc.SessionCount())
	}
}

func TestCloseSession_Nonexistent(t *testing.T) {
	svc := newTestService()
	// Should not panic
	svc.CloseSession("nonexistent")
}

func TestAppendHistory(t *testing.T) {
	svc := newTestService()
	id := svc.CreateSession()

	svc.appendHistory(id, "user msg", "ai reply")
	svc.appendHistory(id, "user msg2", "ai reply2")

	svc.mu.RLock()
	sess := svc.sessions[id]
	svc.mu.RUnlock()

	if sess == nil {
		t.Fatal("session not found")
	}
	if len(sess.History) != 4 {
		t.Errorf("len(History) = %d, want 4", len(sess.History))
	}
}

func TestAppendHistory_EmptyStrings(t *testing.T) {
	svc := newTestService()
	id := svc.CreateSession()

	svc.appendHistory(id, "", "ai reply")
	svc.appendHistory(id, "user msg", "")

	svc.mu.RLock()
	sess := svc.sessions[id]
	svc.mu.RUnlock()

	if len(sess.History) != 2 {
		t.Errorf("len(History) = %d, want 2 (empty strings skipped)", len(sess.History))
	}
}

func TestAppendHistory_Truncation(t *testing.T) {
	svc := newTestService()
	id := svc.CreateSession()

	// Add 15 rounds (30 messages)
	for i := 0; i < 15; i++ {
		svc.appendHistory(id, "user", "ai")
	}

	svc.mu.RLock()
	sess := svc.sessions[id]
	svc.mu.RUnlock()

	if len(sess.History) > 20 {
		t.Errorf("len(History) = %d, want <= 20", len(sess.History))
	}
}

func TestCostTracker_AddTokens(t *testing.T) {
	ct := NewCostTracker("session1")
	ct.AddTokens(100)
	ct.AddTokens(50)
	if ct.Tokens() != 150 {
		t.Errorf("Tokens = %d, want 150", ct.Tokens())
	}
}

func TestCostTracker_ZeroTokens(t *testing.T) {
	ct := NewCostTracker("session1")
	if ct.Tokens() != 0 {
		t.Errorf("Tokens = %d, want 0", ct.Tokens())
	}
}

func TestCostTracker_RPM(t *testing.T) {
	ct := NewCostTracker("session1")
	ct.AddTokens(120)
	rpm := ct.RPM()
	if rpm <= 0 {
		t.Errorf("RPM = %f, want > 0", rpm)
	}
}
