package llm

import (
	"errors"
	"fmt"
)

func ConvertToMessages(messages []Message) ([]Message, error) {
	if len(messages) == 0 {
		return nil, errors.New("messages are required")
	}
	for i, msg := range messages {
		if err := checkMessageLight(msg); err != nil {
			return nil, fmt.Errorf("message at index %d: %w", i, err)
		}
	}
	return messages, nil
}

func PrepareRequest(req *GenerateRequest) ([]Message, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}
	return ConvertToMessages(req.Messages)
}

func checkMessageLight(msg Message) error {
	if msg.Role != "" && !isValidRole(msg.Role) {
		return fmt.Errorf("invalid message role %q", msg.Role)
	}
	if len(msg.Parts) == 0 {
		return errors.New("message parts are required")
	}
	return nil
}

func isValidRole(role Role) bool {
	switch role {
	case RoleSystem, RoleUser, RoleAssistant, RoleTool:
		return true
	default:
		return false
	}
}
