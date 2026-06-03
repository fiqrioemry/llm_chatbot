package openai

// ─── Chat ─────────────────────────────────────────────────────────────────────

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

type ChatMessage struct {
	Role       Role        `json:"role"`
	Content    interface{} `json:"content"` // string | []ContentPart for vision
	Name       string      `json:"name,omitempty"`
	ToolCallID string      `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall  `json:"tool_calls,omitempty"`
}

// ContentPart supports multi-modal messages (text + image_url)
type ContentPart struct {
	Type     string     `json:"type"` // "text" | "image_url"
	Text     string     `json:"text,omitempty"`
	ImageURL *ImageURL  `json:"image_url,omitempty"`
}

type ImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"` // "low" | "high" | "auto"
}

type ChatOptions struct {
	Temperature      float64  `json:"temperature,omitempty"`
	MaxTokens        int      `json:"max_tokens,omitempty"`
	TopP             float64  `json:"top_p,omitempty"`
	FrequencyPenalty float64  `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64  `json:"presence_penalty,omitempty"`
	Stop             []string `json:"stop,omitempty"`
	User             string   `json:"user,omitempty"`
}

type chatRequest struct {
	Model            string        `json:"model"`
	Messages         []ChatMessage `json:"messages"`
	Temperature      float64       `json:"temperature,omitempty"`
	MaxTokens        int           `json:"max_tokens,omitempty"`
	TopP             float64       `json:"top_p,omitempty"`
	FrequencyPenalty float64       `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64       `json:"presence_penalty,omitempty"`
	Stop             []string      `json:"stop,omitempty"`
	Stream           bool          `json:"stream,omitempty"`
	Tools            []Tool        `json:"tools,omitempty"`
	ToolChoice       interface{}   `json:"tool_choice,omitempty"` // "none"|"auto"|ToolChoiceObject
	User             string        `json:"user,omitempty"`
}

type ChatResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Model   string       `json:"model"`
	Choices []ChatChoice `json:"choices"`
	Usage   Usage        `json:"usage"`
}

type ChatChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ─── Streaming ────────────────────────────────────────────────────────────────

type StreamChunk struct {
	ID      string          `json:"id"`
	Object  string          `json:"object"`
	Created int64           `json:"created"`
	Model   string          `json:"model"`
	Choices []StreamChoice  `json:"choices"`
}

type StreamChoice struct {
	Index        int         `json:"index"`
	Delta        StreamDelta `json:"delta"`
	FinishReason *string     `json:"finish_reason"`
}

type StreamDelta struct {
	Role      Role       `json:"role,omitempty"`
	Content   string     `json:"content,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ─── Tool / Function calling ──────────────────────────────────────────────────

type Tool struct {
	Type     string   `json:"type"` // "function"
	Function Function `json:"function"`
}

type Function struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"` // JSON Schema object
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"` // "function"
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string
}

type ToolChoiceObject struct {
	Type     string            `json:"type"` // "function"
	Function ToolChoiceFn      `json:"function"`
}

type ToolChoiceFn struct {
	Name string `json:"name"`
}

// ─── Embeddings ───────────────────────────────────────────────────────────────

type embedRequest struct {
	Model          string      `json:"model"`
	Input          interface{} `json:"input"` // string | []string
	EncodingFormat string      `json:"encoding_format,omitempty"` // "float" | "base64"
	Dimensions     int         `json:"dimensions,omitempty"`
	User           string      `json:"user,omitempty"`
}

type EmbedResponse struct {
	Object string      `json:"object"`
	Data   []EmbedData `json:"data"`
	Model  string      `json:"model"`
	Usage  EmbedUsage  `json:"usage"`
}

type EmbedData struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float32 `json:"embedding"`
}

type EmbedUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// ─── Moderation ───────────────────────────────────────────────────────────────

type moderationRequest struct {
	Input string `json:"input"`
	Model string `json:"model,omitempty"`
}

type ModerationResponse struct {
	ID      string             `json:"id"`
	Model   string             `json:"model"`
	Results []ModerationResult `json:"results"`
}

type ModerationResult struct {
	Flagged        bool                   `json:"flagged"`
	Categories     ModerationCategories   `json:"categories"`
	CategoryScores ModerationScores       `json:"category_scores"`
}

type ModerationCategories struct {
	Sexual                bool `json:"sexual"`
	SexualMinors          bool `json:"sexual/minors"`
	Harassment            bool `json:"harassment"`
	HarassmentThreatening bool `json:"harassment/threatening"`
	Hate                  bool `json:"hate"`
	HateThreatening       bool `json:"hate/threatening"`
	Illicit               bool `json:"illicit"`
	IllicitViolent        bool `json:"illicit/violent"`
	SelfHarm              bool `json:"self-harm"`
	SelfHarmIntent        bool `json:"self-harm/intent"`
	SelfHarmInstructions  bool `json:"self-harm/instructions"`
	Violence              bool `json:"violence"`
	ViolenceGraphic       bool `json:"violence/graphic"`
}

type ModerationScores struct {
	Sexual                float64 `json:"sexual"`
	SexualMinors          float64 `json:"sexual/minors"`
	Harassment            float64 `json:"harassment"`
	HarassmentThreatening float64 `json:"harassment/threatening"`
	Hate                  float64 `json:"hate"`
	HateThreatening       float64 `json:"hate/threatening"`
	Illicit               float64 `json:"illicit"`
	IllicitViolent        float64 `json:"illicit/violent"`
	SelfHarm              float64 `json:"self-harm"`
	SelfHarmIntent        float64 `json:"self-harm/intent"`
	SelfHarmInstructions  float64 `json:"self-harm/instructions"`
	Violence              float64 `json:"violence"`
	ViolenceGraphic       float64 `json:"violence/graphic"`
}

// ─── Error ────────────────────────────────────────────────────────────────────

type apiError struct {
	Error apiErrorDetail `json:"error"`
}

type apiErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}
