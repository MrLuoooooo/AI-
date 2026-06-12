package config

import "time"

// Config holds all application configuration.
type Config struct {
	Server        ServerConfig        `mapstructure:"server"`
	Auth          AuthConfig          `mapstructure:"auth"`
	ModelProvider ModelProviderConfig `mapstructure:"model_provider"`
	VectorStore   VectorStoreConfig   `mapstructure:"vector_store"`
	Document      DocumentConfig      `mapstructure:"document"`
	Conversation  ConversationConfig  `mapstructure:"conversation"`
	Search        SearchConfig        `mapstructure:"search"`
	Cron          CronConfig          `mapstructure:"cron"`
	Log           LogConfig           `mapstructure:"log"`
	Stock         StockConfig         `mapstructure:"stock"`
	Agent         AgentConfig         `mapstructure:"agent"`
	Memory        MemoryConfig        `mapstructure:"memory"`
	Vision        VisionConfig        `mapstructure:"vision"`
}

// StockConfig holds stock data middleware configuration.
type StockConfig struct {
	DataDir string `mapstructure:"data_dir"`
}

// AgentConfig holds agent behavior configuration.
type AgentConfig struct {
	SystemPrompt string `mapstructure:"system_prompt"`
}

// VisionConfig holds AI vision assistant configuration.
type VisionConfig struct {
	Enabled       bool    `mapstructure:"enabled"`
	FrameInterval int     `mapstructure:"frame_interval"`    // ms between frames (default 1000 = 1fps)
	FrameQuality  int     `mapstructure:"frame_quality"`     // JPEG quality 1-100 (default 30)
	MaxFrameWidth int     `mapstructure:"max_frame_width"`   // resize max width (default 640)
	VADMode       int     `mapstructure:"vad_mode"`          // VAD aggressiveness 0-3 (default 1)
	VADFrameMS    int     `mapstructure:"vad_frame_ms"`      // VAD frame size in ms (default 30)
	VADSilenceMS  int     `mapstructure:"vad_silence_ms"`    // silence threshold ms (default 800)
	MaxTokenPerMin int   `mapstructure:"max_token_per_min"`  // rate limit tokens/minute (default 0=off)
	MaxSessions   int     `mapstructure:"max_sessions"`      // max concurrent sessions (default 10)
	SystemPrompt  string  `mapstructure:"system_prompt"`     // vision system prompt
	VisionModel   string  `mapstructure:"vision_model"`      // default vision-capable model
	Detail        string  `mapstructure:"detail"`            // OpenAI vision detail: low/high/auto
}

// MemoryConfig holds long-term memory configuration.
type MemoryConfig struct {
	Enabled       bool `mapstructure:"enabled"`
	RetrievalTopK int  `mapstructure:"retrieval_top_k"`
}

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	APIKey string `mapstructure:"api_key"`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Host         string   `mapstructure:"host"`
	Port         int      `mapstructure:"port"`
	Mode         string   `mapstructure:"mode"`
	WorkspaceDir string   `mapstructure:"workspace_dir"`
	AllowedDirs  []string `mapstructure:"allowed_dirs"`
	BashTimeout  int      `mapstructure:"bash_timeout"`
	CORSOrigins  []string `mapstructure:"cors_origins"`
	RateLimitRPS float64  `mapstructure:"rate_limit_rps"`
}

// ModelEntry defines a switchable model in the model_list.
type ModelEntry struct {
	Name           string `mapstructure:"name"  json:"name"`
	Provider       string `mapstructure:"provider" json:"provider"`
	ChatModel      string `mapstructure:"chat_model" json:"chat_model"`
	APIKey         string `mapstructure:"api_key" json:"api_key"`
	BaseURL        string `mapstructure:"base_url" json:"base_url"`
	EmbeddingModel string `mapstructure:"embedding_model" json:"embedding_model"`
}

// ModelProviderConfig configures local and remote LLM providers.
type ModelProviderConfig struct {
	Mode               string        `mapstructure:"mode"`
	EmbeddingDimension int           `mapstructure:"embedding_dimension"`
	Timeout            time.Duration `mapstructure:"timeout"`

	Local     LocalProviderConfig  `mapstructure:"local"`
	Cloud     CloudProviderConfig  `mapstructure:"cloud"`
	ModelList []ModelEntry         `mapstructure:"model_list"`
}

// LocalProviderConfig for Ollama.
type LocalProviderConfig struct {
	Enabled        bool   `mapstructure:"enabled"`
	BaseURL        string `mapstructure:"base_url"`
	ChatModel      string `mapstructure:"chat_model"`
	EmbeddingModel string `mapstructure:"embedding_model"`
}

// CloudProviderConfig for OpenAI-compatible APIs.
type CloudProviderConfig struct {
	Enabled        bool   `mapstructure:"enabled"`
	Type           string `mapstructure:"type"`
	APIKey         string `mapstructure:"api_key"`
	BaseURL        string `mapstructure:"base_url"`
	ChatModel      string `mapstructure:"chat_model"`
	EmbeddingModel string `mapstructure:"embedding_model"`
}

// VectorStoreConfig holds vector store configuration.
type VectorStoreConfig struct {
	Type          string              `mapstructure:"type"`
	Elasticsearch ElasticsearchConfig `mapstructure:"elasticsearch"`
}

// ElasticsearchConfig holds Elasticsearch configuration.
type ElasticsearchConfig struct {
	Addresses        []string `mapstructure:"addresses"`
	IndexName        string   `mapstructure:"index_name"`
	Username         string   `mapstructure:"username"`
	Password         string   `mapstructure:"password"`
	ConvIndexName    string   `mapstructure:"conv_index_name"`
	ConvMsgIndexName string   `mapstructure:"conv_msg_index_name"`
	DocIndexName     string   `mapstructure:"doc_index_name"`
}

// DocumentConfig holds document processing configuration.
type DocumentConfig struct {
	MaxFileSize  int64    `mapstructure:"max_file_size"`
	AllowedTypes []string `mapstructure:"allowed_types"`
	ChunkSize    int      `mapstructure:"chunk_size"`
	ChunkOverlap int      `mapstructure:"chunk_overlap"`
}

// SearchConfig holds web search configuration.
type SearchConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	APIKey  string `mapstructure:"api_key"`
	BaseURL string `mapstructure:"base_url"`
	Engine  string `mapstructure:"engine"`
	Format  string `mapstructure:"format"`
}

// ConversationConfig holds conversation configuration.
type ConversationConfig struct {
	MaxHistory  int    `mapstructure:"max_history"`
	StorageType string `mapstructure:"storage_type"`
}

// CronJobConfig defines a single scheduled task.
type CronJobConfig struct {
	Name        string `mapstructure:"name"`
	CronExpr    string `mapstructure:"cron_expr"`
	Prompt      string `mapstructure:"prompt"`
	AutoApprove bool   `mapstructure:"auto_approve"`
}

// CronConfig holds cron scheduler configuration.
type CronConfig struct {
	Enabled bool            `mapstructure:"enabled"`
	Jobs    []CronJobConfig `mapstructure:"jobs"`
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level      string `mapstructure:"level"`
	FilePath   string `mapstructure:"file_path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}
