package llm

import (
	"errors"
	"fmt"
	"slices"
)

// Role identifies who produced a chat message.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

var validRoles = []Role{RoleSystem, RoleUser, RoleAssistant, RoleTool}

// GenerateRequest is the shared request shape for all LLM providers.
type GenerateRequest struct {
	Messages   []Message
	Tools      []Tool
	ToolChoice *ToolChoice
}

// Message is a single turn in a conversation with one or more content parts.
type Message struct {
	Role  Role
	Parts []ContentPart
}

// Validate performs full structural validation on the message and every part.
// Optional: Generate/GenerateStream do not call this automatically; use it for
// explicit pre-flight checks or in tests. Provider conversion validates capability.
func (m Message) Validate() error {
	if !slices.Contains(validRoles, m.Role) {
		return fmt.Errorf("invalid message role %q", m.Role)
	}
	if len(m.Parts) == 0 {
		return errors.New("message parts are required")
	}
	for i, part := range m.Parts {
		if err := part.Validate(); err != nil {
			return fmt.Errorf("part at index %d: %w", i, err)
		}
	}
	return nil
}

// Validate performs full structural validation on all messages and parts.
// Optional: providers use PrepareRequest instead; call Validate explicitly when needed.
func (r *GenerateRequest) Validate() error {
	if r == nil {
		return errors.New("request is required")
	}
	if len(r.Messages) == 0 {
		return errors.New("messages are required")
	}
	for i, msg := range r.Messages {
		if err := msg.Validate(); err != nil {
			return fmt.Errorf("message at index %d: %w", i, err)
		}
	}
	if err := ValidateTools(r.Tools); err != nil {
		return err
	}
	if r.ToolChoice != nil && r.ToolChoice.Mode == ToolChoiceRequired && r.ToolChoice.Name == "" && len(r.Tools) == 0 {
		return errors.New("tools are required when tool choice is required")
	}
	return nil
}

// UserMessage builds a user message from content parts.
func UserMessage(parts ...ContentPart) Message {
	return message(RoleUser, parts...)
}

// SystemMessage builds a system message from content parts.
func SystemMessage(parts ...ContentPart) Message {
	return message(RoleSystem, parts...)
}

// AssistantMessage builds an assistant message from content parts.
func AssistantMessage(parts ...ContentPart) Message {
	return message(RoleAssistant, parts...)
}

// ToolMessage builds a tool result message for conversation history.
func ToolMessage(toolCallID, name, result string) Message {
	return message(RoleTool, ToolResultPart(toolCallID, name, result))
}

// AssistantToolCallsMessage builds an assistant message with optional text and tool calls.
func AssistantToolCallsMessage(text string, calls ...ToolCall) Message {
	parts := make([]ContentPart, 0, 1+len(calls))
	if text != "" {
		parts = append(parts, TextPart(text))
	}
	for _, call := range calls {
		parts = append(parts, ToolCallPart(call))
	}
	return message(RoleAssistant, parts...)
}

func message(role Role, parts ...ContentPart) Message {
	return Message{Role: role, Parts: parts}
}