package llm

// FinishReason indicates why the model stopped generating for a choice.
type FinishReason string

const (
	FinishReasonStop          FinishReason = "stop"
	FinishReasonLength        FinishReason = "length"
	FinishReasonToolCalls     FinishReason = "tool_calls"
	FinishReasonContentFilter FinishReason = "content_filter"
)

type UsageMetadata struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// Choice is one completion alternative in a non-streaming response.
type Choice struct {
	Index        int
	Message      Message
	FinishReason FinishReason
}

// StreamChoice is one completion alternative in a streaming chunk.
type StreamChoice struct {
	Index        int
	Delta        Message
	FinishReason FinishReason
}

// GenerateResponse is the unified non-streaming completion result for all providers.
type GenerateResponse struct {
	ID      string
	Model   string
	Object  string
	Created int64
	Choices []Choice
	Usage   *UsageMetadata
	Raw     any
}

// StreamResponse is one chunk from a streaming completion.
type StreamResponse struct {
	ID      string
	Model   string
	Choices []StreamChoice
	Raw     any
}

// NewChoice builds a choice with an assistant message from content parts.
func NewChoice(index int, parts []ContentPart, reason FinishReason) Choice {
	if parts == nil {
		parts = []ContentPart{}
	}
	return Choice{
		Index:        index,
		Message:      Message{Role: RoleAssistant, Parts: parts},
		FinishReason: reason,
	}
}

// NewStreamChoice builds a stream choice with a delta message from content parts.
func NewStreamChoice(index int, parts []ContentPart, reason FinishReason) StreamChoice {
	if parts == nil {
		parts = []ContentPart{}
	}
	return StreamChoice{
		Index:        index,
		Delta:        Message{Role: RoleAssistant, Parts: parts},
		FinishReason: reason,
	}
}

// Text returns concatenated text from the choice message parts.
func (c Choice) Text() string {
	return TextFromParts(c.Message.Parts)
}

// Parts returns the choice message content parts.
func (c Choice) Parts() []ContentPart {
	return c.Message.Parts
}

// Text returns concatenated text from the stream delta parts.
func (s StreamChoice) Text() string {
	return TextFromParts(s.Delta.Parts)
}

// Text returns text from the first choice, or empty when there are no choices.
func (r *GenerateResponse) Text() string {
	if r == nil || len(r.Choices) == 0 {
		return ""
	}
	return r.Choices[0].Text()
}

// FirstChoice returns the first choice, or nil.
func (r *GenerateResponse) FirstChoice() *Choice {
	if r == nil || len(r.Choices) == 0 {
		return nil
	}
	return &r.Choices[0]
}

// AssistantMessage returns the first choice as a message for conversation history.
func (r *GenerateResponse) AssistantMessage() Message {
	if r == nil || len(r.Choices) == 0 {
		return Message{Role: RoleAssistant, Parts: []ContentPart{}}
	}
	parts := make([]ContentPart, len(r.Choices[0].Message.Parts))
	copy(parts, r.Choices[0].Message.Parts)
	return Message{Role: RoleAssistant, Parts: parts}
}

// Text returns text from the first stream choice delta.
func (r *StreamResponse) Text() string {
	if r == nil || len(r.Choices) == 0 {
		return ""
	}
	return r.Choices[0].Text()
}

// RawAs returns the provider-native value with a type assertion.
func RawAs[T any](raw any) (T, bool) {
	v, ok := raw.(T)
	return v, ok
}
