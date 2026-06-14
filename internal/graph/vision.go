package graph

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"ai-vision-assistant/internal/config"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// VisionInput 用户输入：语音文字+当前帧+历史消息。
type VisionInput struct {
	Transcript string            // ASR 文字
	FrameB64   string            // 当前帧 Base64 JPEG
	History    []*schema.Message // 历史对话
}

// VisionGraph Eino 双节点图：build_messages → vision_model。
type VisionGraph struct {
	graph  compose.Runnable[*VisionInput, *schema.Message]
	prompt string
	detail string
	mu     sync.RWMutex
}

// NewVisionGraph 编译 Eino 图：build_messages 拼多模态输入 → vision_model 推理。
func NewVisionGraph(cm model.ChatModel, cfg *config.VisionConfig) (*VisionGraph, error) {
	prompt := cfg.SystemPrompt
	if prompt == "" {
		prompt = "你是一个 AI 视觉助手。请根据摄像头画面和用户的语音输入，用中文简洁回答。"
	}
	detail := cfg.Detail
	if detail == "" {
		detail = "low"
	}

	vg := &VisionGraph{
		prompt: prompt,
		detail: detail,
	}

	graph := compose.NewGraph[*VisionInput, *schema.Message]()

	// 节点 1：组装多模态消息
	graph.AddLambdaNode("build_messages", compose.InvokableLambda(
		vg.buildMessages,
	))

	// 节点 2：多模态模型推理
	graph.AddChatModelNode("vision_model", cm)

	// 连线
	graph.AddEdge(compose.START, "build_messages")
	graph.AddEdge("build_messages", "vision_model")
	graph.AddEdge("vision_model", compose.END)

	runnable, err := graph.Compile(context.Background(),
		compose.WithEagerExecution(),
	)
	if err != nil {
		return nil, fmt.Errorf("compile vision graph: %w", err)
	}

	vg.graph = runnable
	return vg, nil
}

// Invoke 跑一次图，返回 AI 回复。
func (g *VisionGraph) Invoke(ctx context.Context, input *VisionInput) (*schema.Message, error) {
	return g.graph.Invoke(ctx, input)
}

// Stream 流式跑图。
func (g *VisionGraph) Stream(ctx context.Context, input *VisionInput) (*schema.StreamReader[*schema.Message], error) {
	return g.graph.Stream(ctx, input)
}

// UpdatePrompt 运行时换系统提示词。
func (g *VisionGraph) UpdatePrompt(prompt string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.prompt = prompt
}

// Prompt 返回当前系统提示词。
func (g *VisionGraph) Prompt() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.prompt
}

// buildMessages 拼 [SystemPrompt, History..., UserMultimodalMessage]。
func (g *VisionGraph) buildMessages(ctx context.Context, input *VisionInput) ([]*schema.Message, error) {
	g.mu.RLock()
	prompt := g.prompt
	g.mu.RUnlock()

	messages := make([]*schema.Message, 0, len(input.History)+2)

	// 系统提示
	messages = append(messages, &schema.Message{
		Role:    schema.System,
		Content: prompt,
	})

	// 历史消息（截断最近 20 条）
	hist := input.History
	if len(hist) > 20 {
		hist = hist[len(hist)-20:]
	}
	messages = append(messages, hist...)

	// 当前用户消息（百炼多模态格式）
	content := buildVisionContent(input.Transcript, input.FrameB64)
	messages = append(messages, &schema.Message{
		Role:    schema.User,
		Content: content,
	})

	return messages, nil
}

// buildVisionContent 构建百炼多模态 content： [{"text":"用户问：xx【用中文回答】"},{"image":"..."}]
func buildVisionContent(text, frameB64 string) string {
	var sb strings.Builder
	sb.WriteString("[{\"text\":")
	if text == "" {
		sb.WriteString(jsonStr("请描述画面内容【用中文回答】"))
	} else {
		sb.WriteString(jsonStr(text + "【用中文回答】"))
	}
	sb.WriteString("}")
	if frameB64 != "" {
		sb.WriteString(",{\"image\":\"data:image/jpeg;base64,")
		sb.WriteString(frameB64)
		sb.WriteString("\"}")
	}
	sb.WriteString("]")
	return sb.String()
}

// jsonStr 做基本 JSON 字符串转义。
func jsonStr(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return `"` + s + `"`
}
