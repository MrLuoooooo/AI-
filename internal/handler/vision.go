package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"

	"ai-vision-assistant/internal/component/frame"
	"ai-vision-assistant/internal/model"
	"ai-vision-assistant/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// VisionHandler 处理 WebSocket 视觉对话连接。
type VisionHandler struct {
	svc     *service.VisionService
	sampler *frame.Sampler
	logger  *zap.Logger
	sessionCount int64
}

// NewVisionHandler 创建 VisionHandler。
func NewVisionHandler(svc *service.VisionService, sampler *frame.Sampler, logger *zap.Logger) *VisionHandler {
	return &VisionHandler{
		svc:     svc,
		sampler: sampler,
		logger:  logger,
	}
}

// HandleVision 处理 WebSocket 升级并进入视觉对话循环。
func (h *VisionHandler) HandleVision(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("ws upgrade failed", zap.Error(err))
		return
	}
	defer conn.Close()

	// 并发会话限制
	if err := h.svc.CheckLimit(); err != nil {
		h.writeResponse(conn, model.VisionResponse{
			Type:  model.VisionMsgError,
			Error: err.Error(),
		})
		return
	}

	sessionID := h.svc.CreateSession()
	atomic.AddInt64(&h.sessionCount, 1)
	defer func() {
		h.svc.CloseSession(sessionID)
		atomic.AddInt64(&h.sessionCount, -1)
	}()

	h.logger.Info("vision ws connected", zap.String("session", sessionID))

	// 告知客户端会话已建立
	h.writeResponse(conn, model.VisionResponse{
		Type:      model.VisionMsgStatus,
		SessionID: sessionID,
		Content:   "session_ready",
	})

	// 设置读超时 + pong handler 保活
	conn.SetReadDeadline(time.Now().Add(120 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(120 * time.Second))
		return nil
	})

	// 定时发送 ping（每 30 秒）
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()
	go func() {
		for range pingTicker.C {
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}()

	ctx := c.Request.Context()

	// 主循环：收帧 → 采样 → 多模态推理 → 回复
	for {
		msgType, rawMsg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				h.logger.Warn("ws read error", zap.String("session", sessionID), zap.Error(err))
			}
			break
		}

		// 只处理文本消息（JSON）
		if msgType != websocket.TextMessage {
			continue
		}

		var req model.VisionRequest
		if err := json.Unmarshal(rawMsg, &req); err != nil {
			h.logger.Warn("ws invalid json", zap.String("session", sessionID))
			continue
		}

		// 心跳
		if req.Type == model.VisionMsgPing {
			h.writeResponse(conn, model.VisionResponse{Type: model.VisionMsgPong, SessionID: sessionID})
			continue
		}

		// 帧处理：采样 → 推理
		if req.Type == model.VisionMsgFrame {
			shouldSample, compressedFrame, err := h.sampler.ShouldSample(req.Frame)
			if err != nil {
				h.logger.Error("frame sampling error", zap.Error(err))
				continue
			}
			if !shouldSample {
				continue
			}

			go h.doInference(ctx, conn, sessionID, compressedFrame, req.Transcript)
			continue
		}

		// 纯文字输入（无帧）
		if req.Type == model.VisionMsgTranscript && req.Transcript != "" {
			go h.doInference(ctx, conn, sessionID, "", req.Transcript)
		}
	}

	h.logger.Info("vision ws disconnected", zap.String("session", sessionID))
}

// doInference 执行多模态推理并发送回复。
func (h *VisionHandler) doInference(ctx context.Context, conn *websocket.Conn, sessionID, frameB64, transcript string) {
	result, err := h.svc.ProcessFrame(ctx, sessionID, frameB64, transcript)
	if err != nil {
		h.logger.Error("vision inference failed", zap.String("session", sessionID), zap.Error(err))
		h.writeResponse(conn, model.VisionResponse{
			Type:      model.VisionMsgError,
			SessionID: sessionID,
			Error:     "推理失败：" + err.Error(),
		})
		return
	}

	if result.Content != "" {
		h.writeResponse(conn, model.VisionResponse{
			Type:      model.VisionMsgReply,
			SessionID: sessionID,
			Content:   result.Content,
		})
	}
}

// writeResponse 写 JSON 响应到 WebSocket。
func (h *VisionHandler) writeResponse(conn *websocket.Conn, resp model.VisionResponse) {
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	data, err := json.Marshal(resp)
	if err != nil {
		h.logger.Error("json marshal response", zap.Error(err))
		return
	}
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		h.logger.Warn("ws write failed", zap.Error(err))
	}
}

// ActiveSessions 返回当前活跃 WebSocket 连接数。
func (h *VisionHandler) ActiveSessions() int64 {
	return atomic.LoadInt64(&h.sessionCount)
}
