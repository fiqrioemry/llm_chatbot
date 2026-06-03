package constant

const (
	CreateConversationSuccess = "Conversation created"
	ListConversationsSuccess  = "Conversations retrieved"
	DeleteConversationSuccess = "Conversation deleted"
	ListMessagesSuccess       = "Messages retrieved"

	ErrConversationNotFound   = "Conversation not found"
	CodeConversationNotFound  = "CONV_NOT_FOUND"
	ErrConversationForbidden  = "You don't have access to this conversation"
	CodeConversationForbidden = "CONV_FORBIDDEN"
	ErrMessageEmpty           = "Message content cannot be empty"
	CodeMessageEmpty          = "MSG_EMPTY"
)
