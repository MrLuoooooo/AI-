package model

// VisionSession 一个视觉对话会话的状态。
type VisionSession struct {
	ID            string `json:"id"`
	CreatedAt     int64  `json:"created_at"`
	LastFrameAt   int64  `json:"last_frame_at"`   // 最后一帧时间戳
	LastSpeechAt  int64  `json:"last_speech_at"`   // 最后语音时间戳
	FrameCount    int64  `json:"frame_count"`      // 累计发送帧数
	TokenUsed     int64  `json:"token_used"`       // 累计 token 消耗
}

// VisionRequest 客户端 WebSocket 发来的消息。
type VisionRequest struct {
	Type       string `json:"type"`       // "frame" | "transcript" | "ping"
	Frame      string `json:"frame"`      // Base64 编码的 JPEG 帧
	Transcript string `json:"transcript"` // ASR 识别出的文字
	SessionID  string `json:"session_id"` // 复用的会话 ID
}

// VisionResponse 服务端推给客户端的消息。
type VisionResponse struct {
	Type       string `json:"type"`       // "reply" | "status" | "error" | "pong"
	Content    string `json:"content"`    // AI 的文本回复
	SessionID  string `json:"session_id"` // 当前会话 ID
	TokenUsed  int64  `json:"token_used"` // 本次消耗 token
	Error      string `json:"error"`      // 错误信息
}

const (
	VisionMsgFrame      = "frame"
	VisionMsgTranscript = "transcript"
	VisionMsgPing       = "ping"
	VisionMsgReply      = "reply"
	VisionMsgStatus     = "status"
	VisionMsgError      = "error"
	VisionMsgPong       = "pong"
)
