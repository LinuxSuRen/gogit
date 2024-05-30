package oneapi

type ChatMessage struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

type ChatPayload struct {
	Messages []ChatMessage `json:"messages"`
	Model    string        `json:"model"`
}

type ChatResponse struct {
	Created int64                `json:"created"`
	ID      string               `json:"id"`
	Object  string               `json:"object"`
	Usage   map[string]int       `json:"usage"`
	Choices []ChatResponseChoice `json:"choices"`
}

type ChatResponseChoice struct {
	FinishReason string      `json:"finish_reason"`
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
}

func NewChatPayload(message string, model string) ChatPayload {
	return ChatPayload{
		Messages: []ChatMessage{{Content: message, Role: "user"}},
		Model:    model,
	}
}
