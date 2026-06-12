package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"ai-vision-assistant/internal/config"
	"ai-vision-assistant/internal/graph"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// VisionService 视觉对话会话管理与编排。
type VisionService struct {
	visionGraph *graph.VisionGraph
	logger      *zap.Logger

	mu          sync.RWMutex
	sessions    map[string]*graph.VisionInput // sessionID → 最新输入（含历史）
	latestFrame map[string]string             // sessionID → 最新帧 Base64
	cfg         *config.VisionConfig
}

// NewVisionService 创建视觉对话服务。
func NewVisionService(vg *graph.VisionGraph, cfg *config.VisionConfig, logger *zap.Logger) *VisionService {
	return &VisionService{
		visionGraph: vg,
		logger:      logger,
		sessions:    make(map[string]*graph.VisionInput),
		latestFrame: make(map[string]string),
		cfg:         cfg,
	}
}

// CreateSession 创建新会话，返回会话 ID。
func (s *VisionService) CreateSession() string {
	id := uuid.New().String()
	s.mu.Lock()
	s.sessions[id] = &graph.VisionInput{}
	s.mu.Unlock()
	s.logger.Debug("vision session created", zap.String("session_id", id))
	return id
}

// CheckLimit 检查是否达到最大会话数。
func (s *VisionService) CheckLimit() error {
	s.mu.RLock()
	count := len(s.sessions)
	s.mu.RUnlock()
	max := s.cfg.MaxSessions
	if max <= 0 {
		max = 10
	}
	if count >= max {
		return fmt.Errorf("max sessions reached (%d/%d)", count, max)
	}
	return nil
}

// ProcessFrame 处理一帧画面 + 可选的语音文字。
// 返回 AI 的文本回复。
func (s *VisionService) ProcessFrame(ctx context.Context, sessionID, frameB64, transcript string) (*schema.Message, error) {
	s.mu.Lock()
	sess, ok := s.sessions[sessionID]
	if !ok {
		sess = &graph.VisionInput{}
		s.sessions[sessionID] = sess
	}
	// 更新当前输入
	local := &graph.VisionInput{
		Transcript: transcript,
		FrameB64:   frameB64,
		History:    append([]*schema.Message{}, sess.History...),
	}
	s.mu.Unlock()

	result, err := s.visionGraph.Invoke(ctx, local)
	if err != nil {
		s.logger.Error("vision invoke failed", zap.String("session_id", sessionID), zap.Error(err))
		return nil, err
	}

	// 追加到历史
	s.appendHistory(sessionID, transcript, result.Content)

	return result, nil
}

// ProcessFrameStream 流式处理，返回 Eino StreamReader。
// 流结束后自动更新会话历史（与 ProcessFrame 行为一致）。
func (s *VisionService) ProcessFrameStream(ctx context.Context, sessionID, frameB64, transcript string) (*schema.StreamReader[*schema.Message], error) {
	s.mu.Lock()
	sess, ok := s.sessions[sessionID]
	if !ok {
		sess = &graph.VisionInput{}
		s.sessions[sessionID] = sess
	}
	local := &graph.VisionInput{
		Transcript: transcript,
		FrameB64:   frameB64,
		History:    append([]*schema.Message{}, sess.History...),
	}
	s.mu.Unlock()

	raw, err := s.visionGraph.Stream(ctx, local)
	if err != nil {
		return nil, err
	}

	// 包装 reader：流结束后自动更新历史
	pr, pw := schema.Pipe[*schema.Message](64)
	go func() {
		var fullText string
		for {
			chunk, err := raw.Recv()
			if err != nil {
				break
			}
			fullText += chunk.Content
			pw.Send(chunk, nil)
		}
		// 流结束 → 写入历史
		if fullText != "" {
			s.appendHistory(sessionID, transcript, fullText)
		}
		pw.Close()
	}()
	return pr, nil
}

// appendHistory 追加用户消息和 AI 回复到会话历史。
func (s *VisionService) appendHistory(sessionID, userText, aiText string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[sessionID]
	if !ok {
		return
	}
	if userText != "" {
		sess.History = append(sess.History, &schema.Message{Role: schema.User, Content: userText})
	}
	if aiText != "" {
		sess.History = append(sess.History, &schema.Message{Role: schema.Assistant, Content: aiText})
	}
	if len(sess.History) > 20 {
		sess.History = sess.History[len(sess.History)-20:]
	}
}

// SetLatestFrame 存储会话最新帧。
func (s *VisionService) SetLatestFrame(sessionID, frameB64 string) {
	s.mu.Lock()
	s.latestFrame[sessionID] = frameB64
	s.mu.Unlock()
}

// GetLatestFrame 获取会话最新帧。
func (s *VisionService) GetLatestFrame(sessionID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.latestFrame[sessionID]
}

// CloseSession 关闭并清理会话。
func (s *VisionService) CloseSession(sessionID string) {
	s.mu.Lock()
	delete(s.sessions, sessionID)
	s.mu.Unlock()
	s.logger.Debug("vision session closed", zap.String("session_id", sessionID))
}

// SessionCount 返回当前活跃会话数。
func (s *VisionService) SessionCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sessions)
}

// ── 成本控制（设计文档用） ──

// CostTracker 追踪单会话的 token 消耗。
type CostTracker struct {
	mu        sync.Mutex
	sessionID string
	tokens    int64
	startedAt time.Time
}

// NewCostTracker 创建成本追踪器。
func NewCostTracker(sessionID string) *CostTracker {
	return &CostTracker{
		sessionID: sessionID,
		startedAt: time.Now(),
	}
}

// AddTokens 累加 token 消耗。
func (ct *CostTracker) AddTokens(n int64) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	ct.tokens += n
}

// Tokens 返回累计 token 数。
func (ct *CostTracker) Tokens() int64 {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	return ct.tokens
}

// RPM 返回当前每分钟 token 消耗率。
func (ct *CostTracker) RPM() float64 {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	elapsed := time.Since(ct.startedAt).Minutes()
	if elapsed <= 0 {
		return float64(ct.tokens)
	}
	return float64(ct.tokens) / elapsed
}
