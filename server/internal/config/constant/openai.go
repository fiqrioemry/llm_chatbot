package constant

const (
	ErrOpenAIUnavailable    = "AI service is temporarily unavailable"
	ErrOpenAIRateLimited    = "AI service rate limit reached, please try again later"
	ErrOpenAIContextTooLong = "Input is too long for the AI model"
	ErrOpenAIModerated      = "Your message was flagged by content moderation"
	ErrOpenAIBadResponse    = "Received an unexpected response from the AI service"
	ErrEmbeddingFailed      = "Failed to generate text embedding"
)

const (
	CodeOpenAIUnavailable    = "OPENAI_UNAVAILABLE"
	CodeOpenAIRateLimited    = "OPENAI_RATE_LIMITED"
	CodeOpenAIContextTooLong = "OPENAI_CONTEXT_TOO_LONG"
	CodeOpenAIModerated      = "OPENAI_MODERATED"
	CodeOpenAIBadResponse    = "OPENAI_BAD_RESPONSE"
	CodeEmbeddingFailed      = "EMBEDDING_FAILED"
)
