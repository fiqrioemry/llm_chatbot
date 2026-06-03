package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"server/pkg/response"
)

const (
	defaultBaseURL     = "https://api.openai.com/v1"
	defaultTimeout     = 60 * time.Second
	streamTimeout      = 5 * time.Minute
	defaultChatModel   = "gpt-4o-mini"
	defaultEmbedModel  = "text-embedding-3-small"
	defaultMaxTokens   = 2048
	defaultTemperature = 0.7
)

// Client wraps the OpenAI REST API.
// All methods return AppError so callers can pass them directly to response.HandleError.
type Client struct {
	apiKey     string
	baseURL    string
	chatModel  string
	embedModel string
	http       *http.Client
	streamHTTP *http.Client // longer timeout for streaming
}

func New(apiKey, chatModel, embedModel string) *Client {
	if chatModel == "" {
		chatModel = defaultChatModel
	}
	if embedModel == "" {
		embedModel = defaultEmbedModel
	}
	return &Client{
		apiKey:     apiKey,
		baseURL:    defaultBaseURL,
		chatModel:  chatModel,
		embedModel: embedModel,
		http:       &http.Client{Timeout: defaultTimeout},
		streamHTTP: &http.Client{Timeout: streamTimeout},
	}
}

// ─── Embed ────────────────────────────────────────────────────────────────────

// Embed returns the embedding vector for a single text.
func (c *Client) Embed(ctx context.Context, text string) ([]float32, error) {
	vectors, err := c.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	return vectors[0], nil
}

// EmbedBatch returns embedding vectors for multiple texts in a single API call.
// Returns results in the same order as the input slice.
func (c *Client) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	req := embedRequest{
		Model:          c.embedModel,
		Input:          texts,
		EncodingFormat: "float",
	}

	var res EmbedResponse
	if err := c.post(ctx, c.http, "/embeddings", req, &res); err != nil {
		return nil, err
	}

	vectors := make([][]float32, len(res.Data))
	for _, d := range res.Data {
		vectors[d.Index] = d.Embedding
	}

	log.Printf("[provider:openai] EmbedBatch — texts: %d, tokens: %d", len(texts), res.Usage.TotalTokens)
	return vectors, nil
}

// ─── Chat ─────────────────────────────────────────────────────────────────────

// Chat sends a non-streaming chat completion and returns the assistant reply.
func (c *Client) Chat(ctx context.Context, messages []ChatMessage, opts *ChatOptions) (string, error) {
	res, err := c.ChatFull(ctx, messages, opts)
	if err != nil {
		return "", err
	}
	if len(res.Choices) == 0 {
		return "", response.InternalErr("OpenAI returned no choices", "OPENAI_NO_CHOICES")
	}

	content, ok := res.Choices[0].Message.Content.(string)
	if !ok {
		return "", response.InternalErr("unexpected content type in OpenAI response", "OPENAI_CONTENT_TYPE")
	}
	return content, nil
}

// ChatFull returns the full ChatResponse including usage stats.
func (c *Client) ChatFull(ctx context.Context, messages []ChatMessage, opts *ChatOptions) (*ChatResponse, error) {
	req := c.buildChatRequest(messages, opts, false, nil)

	var res ChatResponse
	if err := c.post(ctx, c.http, "/chat/completions", req, &res); err != nil {
		return nil, err
	}

	log.Printf("[provider:openai] Chat — model: %s, prompt_tokens: %d, completion_tokens: %d",
		res.Model, res.Usage.PromptTokens, res.Usage.CompletionTokens)
	return &res, nil
}

// ─── Streaming ────────────────────────────────────────────────────────────────

// ChatStream sends a streaming chat completion.
// onChunk is called for each text delta received from the stream.
// The function blocks until the stream ends or ctx is cancelled.
func (c *Client) ChatStream(ctx context.Context, messages []ChatMessage, opts *ChatOptions, onChunk func(delta string)) error {
	req := c.buildChatRequest(messages, opts, true, nil)

	body, err := json.Marshal(req)
	if err != nil {
		return response.InternalErr("failed to encode chat request", "OPENAI_ENCODE_ERR")
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return response.InternalErr("failed to build stream request", "OPENAI_REQUEST_ERR")
	}
	c.setHeaders(httpReq)

	httpRes, err := c.streamHTTP.Do(httpReq)
	if err != nil {
		return response.InternalErr(fmt.Sprintf("OpenAI stream request failed: %v", err), "OPENAI_STREAM_ERR")
	}
	defer httpRes.Body.Close()

	if httpRes.StatusCode != http.StatusOK {
		return c.parseHTTPError(httpRes)
	}

	scanner := bufio.NewScanner(httpRes.Body)
	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimPrefix(line, "data: ")
		if payload == "[DONE]" {
			break
		}

		var chunk StreamChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			log.Printf("[provider:openai] ChatStream — failed to parse chunk: %v", err)
			continue
		}

		if len(chunk.Choices) > 0 {
			onChunk(chunk.Choices[0].Delta.Content)
		}
	}

	if err := scanner.Err(); err != nil {
		return response.InternalErr(fmt.Sprintf("stream read error: %v", err), "OPENAI_STREAM_READ_ERR")
	}

	return nil
}

// ─── Tool / Function calling ──────────────────────────────────────────────────

// ChatWithTools sends a chat completion with tool definitions.
// Returns the full response so callers can inspect ToolCalls in the choice.
func (c *Client) ChatWithTools(ctx context.Context, messages []ChatMessage, tools []Tool, opts *ChatOptions) (*ChatResponse, error) {
	req := c.buildChatRequest(messages, opts, false, tools)

	var res ChatResponse
	if err := c.post(ctx, c.http, "/chat/completions", req, &res); err != nil {
		return nil, err
	}

	log.Printf("[provider:openai] ChatWithTools — model: %s, finish_reason: %s, total_tokens: %d",
		res.Model, res.Choices[0].FinishReason, res.Usage.TotalTokens)
	return &res, nil
}

// ─── Moderation ───────────────────────────────────────────────────────────────

// Moderate checks whether text violates OpenAI's usage policies.
// Returns true if the content is flagged.
func (c *Client) Moderate(ctx context.Context, text string) (bool, *ModerationResponse, error) {
	req := moderationRequest{
		Input: text,
		Model: "omni-moderation-latest",
	}

	var res ModerationResponse
	if err := c.post(ctx, c.http, "/moderations", req, &res); err != nil {
		return false, nil, err
	}

	if len(res.Results) == 0 {
		return false, &res, nil
	}

	log.Printf("[provider:openai] Moderate — flagged: %v", res.Results[0].Flagged)
	return res.Results[0].Flagged, &res, nil
}

// ─── RAG helpers ─────────────────────────────────────────────────────────────

// BuildRAGMessages constructs the message slice for a RAG completion:
// a system prompt that injects retrieved context chunks, followed by
// the conversation history ending with the current user question.
func BuildRAGMessages(systemPrompt, context, question string, history []ChatMessage) []ChatMessage {
	system := systemPrompt
	if context != "" {
		system += "\n\n--- Context ---\n" + context + "\n--- End Context ---"
	}

	msgs := make([]ChatMessage, 0, len(history)+2)
	msgs = append(msgs, ChatMessage{Role: RoleSystem, Content: system})
	msgs = append(msgs, history...)
	msgs = append(msgs, ChatMessage{Role: RoleUser, Content: question})
	return msgs
}

// ─── internal helpers ─────────────────────────────────────────────────────────

func (c *Client) buildChatRequest(messages []ChatMessage, opts *ChatOptions, stream bool, tools []Tool) chatRequest {
	req := chatRequest{
		Model:    c.chatModel,
		Messages: messages,
		Stream:   stream,
	}

	// apply defaults
	req.MaxTokens = defaultMaxTokens
	req.Temperature = defaultTemperature

	if opts != nil {
		if opts.MaxTokens > 0 {
			req.MaxTokens = opts.MaxTokens
		}
		if opts.Temperature > 0 {
			req.Temperature = opts.Temperature
		}
		if opts.TopP > 0 {
			req.TopP = opts.TopP
		}
		if opts.FrequencyPenalty != 0 {
			req.FrequencyPenalty = opts.FrequencyPenalty
		}
		if opts.PresencePenalty != 0 {
			req.PresencePenalty = opts.PresencePenalty
		}
		if len(opts.Stop) > 0 {
			req.Stop = opts.Stop
		}
		if opts.User != "" {
			req.User = opts.User
		}
	}

	if len(tools) > 0 {
		req.Tools = tools
		req.ToolChoice = "auto"
	}

	return req
}

func (c *Client) post(ctx context.Context, client *http.Client, path string, payload, target interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return response.InternalErr("failed to encode request body", "OPENAI_ENCODE_ERR")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return response.InternalErr("failed to build HTTP request", "OPENAI_REQUEST_ERR")
	}
	c.setHeaders(req)

	res, err := client.Do(req)
	if err != nil {
		return response.InternalErr(fmt.Sprintf("OpenAI request failed: %v", err), "OPENAI_HTTP_ERR")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return c.parseHTTPError(res)
	}

	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return response.InternalErr("failed to decode OpenAI response", "OPENAI_DECODE_ERR")
	}
	return nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
}

func (c *Client) parseHTTPError(res *http.Response) error {
	raw, _ := io.ReadAll(res.Body)

	var apiErr apiError
	if err := json.Unmarshal(raw, &apiErr); err == nil && apiErr.Error.Message != "" {
		msg := apiErr.Error.Message

		switch res.StatusCode {
		case http.StatusUnauthorized:
			return response.UnauthorizedErr(msg, "OPENAI_UNAUTHORIZED")
		case http.StatusTooManyRequests:
			return response.NewError(429, msg, "OPENAI_RATE_LIMITED")
		case http.StatusBadRequest:
			return response.BadRequestErr(msg, "OPENAI_BAD_REQUEST")
		default:
			return response.InternalErr(msg, "OPENAI_API_ERR")
		}
	}

	return response.InternalErr(
		fmt.Sprintf("OpenAI returned status %d", res.StatusCode),
		"OPENAI_API_ERR",
	)
}
