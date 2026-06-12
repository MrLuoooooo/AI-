package qianwenmodel

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"
)

// QianwenModel 阿里云百炼多模态模型，直接调用 DashScope multimodal-generation 端点。
type QianwenModel struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
	logger  *zap.Logger
}

// NewQianwenModel 初始化百炼多模态模型。
func NewQianwenModel(apiKey, modelName string, logger *zap.Logger) model.ChatModel {
	return &QianwenModel{
		apiKey:  apiKey,
		model:   modelName,
		baseURL: "https://dashscope.aliyuncs.com/api/v1/services/aigc/multimodal-generation/generation",
		client:  &http.Client{Timeout: 60 * time.Second},
		logger:  logger,
	}
}

// -- 百炼请求结构 --

type dsContentPart struct {
	Text  string `json:"text,omitempty"`
	Image string `json:"image,omitempty"`
}

type dsMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"` // string or []dsContentPart
}

type dsRequest struct {
	Model  string `json:"model"`
	Input  struct {
		Messages []dsMessage `json:"messages"`
	} `json:"input"`
	Parameters struct {
		IncrementalOutput bool    `json:"incremental_output,omitempty"`
		MaxTokens         *int    `json:"max_tokens,omitempty"`
		Temperature       float64 `json:"temperature,omitempty"`
	} `json:"parameters,omitempty"`
}

type dsResponse struct {
	Output struct {
		Choices []struct {
			Message      dsOutputMessage `json:"message"`
			FinishReason string          `json:"finish_reason"`
		} `json:"choices"`
	} `json:"output"`
	RequestID string `json:"request_id"`
}

type dsOutputMessage struct {
	Content []dsContentPart `json:"content"`
	Role    string          `json:"role"`
}

// -- model.ChatModel 实现 --

func (m *QianwenModel) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	req, err := m.buildRequest(input, false, opts)
	if err != nil {
		return nil, err
	}
	resp, err := m.doRequest(ctx, req, nil)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *QianwenModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	req, err := m.buildRequest(input, true, opts)
	if err != nil {
		return nil, err
	}

	streamReader, streamWriter := schema.Pipe[*schema.Message](64)

	go func() {
		defer streamWriter.Close()
		_ = m.doStreamRequest(ctx, req, streamWriter)
	}()

	return streamReader, nil
}

func (m *QianwenModel) BindTools(tools []*schema.ToolInfo) error { return nil }

// -- 内部方法 --

func (m *QianwenModel) buildRequest(input []*schema.Message, stream bool, opts []model.Option) ([]byte, error) {
	req := dsRequest{Model: m.model}
	req.Input.Messages = make([]dsMessage, len(input))
	for i, msg := range input {
		dm := dsMessage{Role: string(msg.Role)}
		content := strings.TrimSpace(msg.Content)
		if strings.HasPrefix(content, "[") {
			// 已是百炼格式的 JSON 数组，直接使用
			dm.Content = json.RawMessage(content)
		} else {
			// 纯文本 → 包装为百炼格式 [{"text":"内容"}]
			dm.Content = json.RawMessage(wrapTextContent(content))
		}
		req.Input.Messages[i] = dm
	}

	if stream {
		req.Parameters.IncrementalOutput = true
	}
	req.Parameters.Temperature = 0.5

	return json.Marshal(req)
}

// wrapTextContent 将纯文本包装为百炼多模态 content 数组格式。
func wrapTextContent(text string) string {
	return `[{"text":"` + jsonEscape(text) + `"}]`
}

// jsonEscape 转义 JSON 字符串中的特殊字符。
func jsonEscape(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return s
}

func (m *QianwenModel) doRequest(ctx context.Context, body []byte, streamWriter *schema.StreamWriter[*schema.Message]) (*schema.Message, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "POST", m.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+m.apiKey)

	if streamWriter != nil {
		httpReq.Header.Set("X-DashScope-SSE", "enable")
	}

	resp, err := m.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("qianwen request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("qianwen read: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("qianwen error (%d): %s", resp.StatusCode, string(respBody))
	}

	var dsResp dsResponse
	if err := json.Unmarshal(respBody, &dsResp); err != nil {
		return nil, fmt.Errorf("qianwen unmarshal: %w", err)
	}

	// 从 content 数组中提取文本
	var fullText string
	for _, choice := range dsResp.Output.Choices {
		for _, part := range choice.Message.Content {
			fullText += part.Text
		}
	}

	return &schema.Message{Role: schema.Assistant, Content: fullText}, nil
}

func (m *QianwenModel) doStreamRequest(ctx context.Context, body []byte, sw *schema.StreamWriter[*schema.Message]) error {
	httpReq, err := http.NewRequestWithContext(ctx, "POST", m.baseURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+m.apiKey)
	httpReq.Header.Set("X-DashScope-SSE", "enable")

	resp, err := m.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("qianwen stream: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("qianwen stream error (%d): %s", resp.StatusCode, string(b))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" {
			continue
		}

		var event struct {
			Output struct {
				Choices []struct {
					Message      dsOutputMessage `json:"message"`
					FinishReason string          `json:"finish_reason"`
				} `json:"choices"`
			} `json:"output"`
		}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		var fullText string
		for _, choice := range event.Output.Choices {
			for _, part := range choice.Message.Content {
				fullText += part.Text
			}
		}

		if fullText != "" {
			sw.Send(&schema.Message{Role: schema.Assistant, Content: fullText}, nil)
		}
	}

	return nil
}

var _ model.ChatModel = (*QianwenModel)(nil)
