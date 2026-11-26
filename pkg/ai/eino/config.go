package eino

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"

	einoOpenAI "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

type Configs struct {
	Configs   []*Config `json:"configs"`
	LastModel string    `json:"last_model"`
	Mcp       McpConfig `json:"mcp"`
}

// McpConfig 控制本站点 MCP 端点的启用与鉴权。
// 存储于 ai_setting (provider.AiSettingKey)，通过 SettingAiForm 配置。
type McpConfig struct {
	Enabled        bool     `json:"enabled"`         // 是否启用 MCP 对外端点
	Token          string   `json:"token"`           // 鉴权 Bearer token，空则禁用
	ExposedTools   []string `json:"exposed_tools"`   // 旧版能力名白名单（C 阶段），按意图底层能力命中过滤；空则暴露全部
	ExposedIntents []string `json:"exposed_intents"` // 意图级白名单：意图名 / 能力域名(content 等) / "*"；非空时未知意图默认拒绝
	ToolListMode   string   `json:"tool_list_mode"`  // ""=完整 schema；"summary"=tools/list 仅返回轻量 schema，降低 token
	// EnableSetScope 是否开放 mcp_set_scope（两阶段能力域裁剪）。**默认 false。**
	//
	// 2026-10-07 新增，默认关闭：mcp_set_scope 的 scope 是**进程级全局**的，
	// 一个客户端调用会让所有并发客户端的 tools/list 一起收窄，且新会话恢复不了
	// （须重启服务）。在改成会话级之前不对外暴露。
	// 详见 pkg/mcp/intent/mcp.go 的 registerMeta 注释。
	EnableSetScope bool `json:"enable_set_scope"`
	RateLimit      int  `json:"rate_limit"` // 每分钟调用上限，0 不限

	// InvokeAdminId 是「通用 REST 调用」（api_invoke 意图）使用的管理员身份。
	// 为 0 时该能力不可用 —— 这是有意的安全设计：MCP/AI 通道无法从请求里推导出操作者，
	// 若隐式回落到超级管理员（AdminPermission 对 adminId==1 直接放行）等于发放万能钥匙。
	// 请显式指定一个权限受控的专用管理员账号。
	InvokeAdminId uint `json:"invoke_admin_id"`

	// ApiExposure 控制「通用 REST 调用」能触达哪些后台端点（G4 白名单）。
	// 这是意图白名单之下的第二道闸门：即使 api_* 意图被显式开放，
	// 具体端点仍需通过这里的策略才能被调用。零值配置 = 全部拒绝（fail closed）。
	ApiExposure ApiExposureConfig `json:"api_exposure"`
}

// ApiExposureConfig 是通用 REST 调用的开放范围策略。
//
// 为什么需要它：api_invoke 能力等价 shell —— 能列出并执行任意后台端点。
// 意图级开关（ExposedIntents）只决定"要不要开放这类能力"，
// 而这里决定"开放到什么程度"，两者是不同粒度。
//
// 判定顺序（全部 fail closed，任一不通过即拒绝）：
//  1. 内置硬规则（凭证类端点、未修复鉴权缺陷的模块、管理员变更的写操作）—— 不可配置；
//  2. DenyNS / DenyEndpoints —— 显式黑名单；
//  3. AllowNS —— 非空则为白名单，不在其内即拒绝；
//  4. Mode 与端点风险等级匹配。
type ApiExposureConfig struct {
	// Mode 决定允许的风险等级：
	//   ""|"off"        —— 关闭（默认，零值即关闭）
	//   "read"          —— 仅只读端点（推荐的起步配置）
	//   "read_write"    —— 读 + 写，仍拒绝 destructive
	//   "all"           —— 含 destructive（删除类），仅在明确需要时开启
	Mode string `json:"mode"`

	// AllowNS 命名空间白名单，前缀匹配：填 "archive" 命中 archive/...，
	// 填 "plugin/keyword" 只命中该插件。留空表示不按命名空间限制（仍受 Mode 约束）。
	AllowNS []string `json:"allow_ns"`

	// DenyNS 命名空间黑名单，前缀匹配，优先级高于 AllowNS。
	DenyNS []string `json:"deny_ns"`

	// DenyEndpoints 端点黑名单，格式 "METHOD /path"（METHOD 可省略表示所有方法），
	// 路径允许省略 /system/api 前缀。优先级最高（硬规则之外）。
	DenyEndpoints []string `json:"deny_endpoints"`
}

// Config holds the configuration for Eino AI integration.
type Config struct {
	APIKey          string  `json:"api_key"`
	Model           string  `json:"model"`
	BaseURL         string  `json:"base_url"`
	MaxTokens       int     `json:"max_tokens"`
	EnableReasoning bool    `json:"enable_reasoning"`
	Temperature     float64 `json:"temperature"`
	TimeoutSeconds  int     `json:"timeout_seconds"`
	MaxRetries      int     `json:"max_retries"`
	Name            string  `json:"name"`
}

// SelectedConfig resolves either a current selector value or a legacy model name.
func (c *Configs) SelectedConfig(selection string) *Config {
	if indexText, custom := strings.CutPrefix(selection, "custom:"); custom {
		index, err := strconv.Atoi(indexText)
		if err == nil && index >= 0 && index < len(c.Configs) {
			return c.Configs[index]
		}
		return nil
	}
	for _, cfg := range c.Configs {
		if cfg != nil && cfg.Model == selection {
			return cfg
		}
	}
	return nil
}

func (c *Configs) DefaultConfig() *Config {
	if cfg := c.SelectedConfig(c.LastModel); cfg != nil {
		return cfg
	}
	for _, cfg := range c.Configs {
		if cfg != nil {
			return cfg
		}
	}
	return nil
}

var (
	globalConfig *Config
	globalClient *einoOpenAI.ChatModel
	globalMu     sync.RWMutex
)

func GlobalConfig() *Config {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalConfig
}

func ClearGlobalConfig() {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalConfig, globalClient = nil, nil
}

// SetGlobalConfig initializes a client with the saved provider credentials.
func SetGlobalConfig(cfg *Config) error {
	if cfg == nil {
		return errors.New("config cannot be nil")
	}
	if cfg.APIKey == "" {
		return errors.New("API key is required")
	}
	if cfg.Model == "" {
		return errors.New("Model is required")
	}
	config := *cfg
	if config.MaxTokens == 0 {
		config.MaxTokens = 8192
	}
	if config.Temperature == 0 {
		config.Temperature = 0.7
	}
	if config.TimeoutSeconds == 0 {
		config.TimeoutSeconds = 120
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.BaseURL == "" {
		config.BaseURL = "https://api.deepseek.com"
	}

	options := &einoOpenAI.ChatModelConfig{
		BaseURL:   config.BaseURL,
		APIKey:    config.APIKey,
		Model:     config.Model,
		MaxTokens: &config.MaxTokens,
	}
	if config.EnableReasoning {
		options.ExtraFields = map[string]any{"deepseek_reasoning": true}
	} else {
		temperature := float32(config.Temperature)
		options.Temperature = &temperature
	}
	client, err := einoOpenAI.NewChatModel(context.Background(), options)
	if err != nil {
		return err
	}
	globalMu.Lock()
	defer globalMu.Unlock()
	globalConfig, globalClient = &config, client
	return nil
}

func GetClient() (*einoOpenAI.ChatModel, error) {
	globalMu.RLock()
	defer globalMu.RUnlock()
	if globalClient == nil {
		return nil, errors.New("client not configured")
	}
	return globalClient, nil
}

// GenerateText uses Eino ChatModel to generate text completion.
func GenerateText(ctx context.Context, prompt string, options ...GenerateOption) (string, error) {
	client, err := GetClient()
	if err != nil {
		return "", err
	}

	cfg := applyOptions(options)

	messages := []*schema.Message{
		schema.UserMessage(prompt),
	}

	msg, err := client.Generate(ctx, messages)
	if err != nil {
		return "", err
	}

	result := msg.Content
	if cfg.Stream {
		// For streaming, we still do a full generate since we concatenate
		// This path is kept for API compatibility
	}

	return result, nil
}

// GenerateStructured uses Eino ChatModel to generate structured JSON output.
func GenerateStructured[T any](ctx context.Context, prompt string, systemPrompt string) (*T, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	var messages []*schema.Message
	if systemPrompt != "" {
		messages = append(messages, schema.SystemMessage(systemPrompt))
	}
	messages = append(messages, schema.UserMessage(prompt))

	msg, err := client.Generate(ctx, messages)
	if err != nil {
		return nil, err
	}

	var parsed T
	if err := parseJSON(msg.Content, &parsed); err != nil {
		return nil, err
	}
	return &parsed, nil
}

// compose graph node types for content management
type GraphInput struct {
	Action   string `json:"action"` // "create", "update", "suggest"
	Title    string `json:"title"`
	Content  string `json:"content"`
	Category string `json:"category"`
	Keywords string `json:"keywords"`
	Language string `json:"language"`
}

type GraphOutput struct {
	Success      bool     `json:"success"`
	Message      string   `json:"message"`
	Title        string   `json:"title"`
	Content      string   `json:"content"`
	Keywords     string   `json:"keywords"`
	Description  string   `json:"description"`
	Suggestions  []string `json:"suggestions"`
	SEO          string   `json:"seo"`
	ErrorMessage string   `json:"error_message,omitempty"`
}

// GenerateOption allows configuring GenerateText calls.
type GenerateOption func(*generateConfig)

type generateConfig struct {
	Stream    bool
	MaxTokens int
}

func applyOptions(opts []GenerateOption) *generateConfig {
	cfg := &generateConfig{Stream: false}
	for _, o := range opts {
		o(cfg)
	}
	return cfg
}

func WithStream() GenerateOption {
	return func(c *generateConfig) { c.Stream = true }
}

func WithMaxTokens(n int) GenerateOption {
	return func(c *generateConfig) { c.MaxTokens = n }
}

func parseJSON(s string, v any) error {
	if s == "" {
		return nil
	}
	return nil
}
