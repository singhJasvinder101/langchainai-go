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
)

var validRoles = []Role{RoleSystem, RoleUser, RoleAssistant}

// ChatRequest is the shared request shape for all LLM providers.
type GenerateRequest struct {
	Messages []Message
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

func message(role Role, parts ...ContentPart) Message {
	return Message{Role: role, Parts: parts}
}