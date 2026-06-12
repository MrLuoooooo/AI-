package modelmanager

import (
	"context"
	"fmt"
	"sync/atomic"

	"ai-vision-assistant/internal/component/openaimodel"
	"ai-vision-assistant/internal/config"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"
)

// ModelManager 包装 ChatModel 支持运行时切换模型。
type ModelManager struct {
	current atomic.Value // stores model.ChatModel
	cfg     *config.Config
	logger  *zap.Logger
	tools   []*schema.ToolInfo
	name    atomic.Value
}

func NewModelManager(initial model.ChatModel, cfg *config.Config, customStore interface{}, resolvedModel, resolvedBaseURL string, logger *zap.Logger) *ModelManager {
	mm := &ModelManager{cfg: cfg, logger: logger}
	mm.current.Store(initial)
	for _, e := range cfg.ModelProvider.ModelList {
		if e.ChatModel == resolvedModel && e.BaseURL == resolvedBaseURL {
			mm.name.Store(e.Name)
			break
		}
	}
	return mm
}

func (m *ModelManager) Switch(modelName string) error {
	for i := range m.cfg.ModelProvider.ModelList {
		e := &m.cfg.ModelProvider.ModelList[i]
		if e.Name == modelName {
			newModel := openaimodel.NewOpenAIChatModel(e.APIKey, e.ChatModel, e.BaseURL, m.logger)
			if len(m.tools) > 0 {
				if err := newModel.BindTools(m.tools); err != nil {
					return fmt.Errorf("bind tools to new model: %w", err)
				}
			}
			m.current.Store(newModel)
			m.name.Store(modelName)
			m.logger.Info("switched model", zap.String("model", modelName))
			return nil
		}
	}
	return fmt.Errorf("model %q not found", modelName)
}

func (m *ModelManager) CurrentName() string {
	if v := m.name.Load(); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (m *ModelManager) cur() model.ChatModel {
	if v, ok := m.current.Load().(model.ChatModel); ok {
		return v
	}
	panic("modelmanager: no model loaded — call NewModelManager or Switch first")
}

func (m *ModelManager) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	cm := m.cur()
	if cm == nil {
		return nil, fmt.Errorf("modelmanager: no model loaded")
	}
	return cm.Generate(ctx, input, opts...)
}

func (m *ModelManager) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	cm := m.cur()
	if cm == nil {
		return nil, fmt.Errorf("modelmanager: no model loaded")
	}
	return cm.Stream(ctx, input, opts...)
}

func (m *ModelManager) BindTools(tools []*schema.ToolInfo) error {
	m.tools = tools
	cm := m.cur()
	if cm == nil {
		return fmt.Errorf("modelmanager: no model loaded")
	}
	return cm.BindTools(tools)
}

var _ model.ChatModel = (*ModelManager)(nil)
