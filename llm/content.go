package llm

import (
	"errors"
	"fmt"
	"strings"
)

// ContentPartType identifies a block within a message.
type ContentPartType string

const (
	PartText       ContentPartType = "text"
	PartImageURL   ContentPartType = "image_url"
	PartImage      ContentPartType = "image"
	PartFile       ContentPartType = "file"
	PartToolCall   ContentPartType = "tool_call"
	PartToolResult ContentPartType = "tool_result"
)

// ContentPart is one block of message content (text, image, file, tool call, etc.).
type ContentPart struct {
	Type       ContentPartType
	Text       string
	URL        string
	MIMEType   string
	Data       []byte
	Name       string
	ToolCallID string
	ToolCall   *ToolCall
}

// Validate checks that required fields are set for the part type.
// Optional: not run automatically on Generate; use explicitly or via GenerateRequest.Validate.
func (p ContentPart) Validate() error {
	switch p.Type {
	case PartText:
		if p.Text == "" {
			return errors.New("text is required")
		}
	case PartImageURL:
		if p.URL == "" {
			return errors.New("image url is required")
		}
	case PartImage, PartFile:
		if len(p.Data) == 0 {
			return fmt.Errorf("%s data is required", p.Type)
		}
		if p.MIMEType == "" {
			return fmt.Errorf("%s mime type is required", p.Type)
		}
	case PartToolCall:
		if p.ToolCall == nil {
			return errors.New("tool call is required")
		}
		return p.ToolCall.Validate()
	case PartToolResult:
		if p.ToolCallID == "" {
			return errors.New("tool call id is required")
		}
		if p.Name == "" {
			return errors.New("tool name is required")
		}
	default:
		return fmt.Errorf("unknown content part type %q", p.Type)
	}
	return nil
}

// TextPart returns a text content block.
func TextPart(text string) ContentPart {
	return ContentPart{Type: PartText, Text: text}
}

// ImageURLPart returns an image referenced by URL (https or data URI).
func ImageURLPart(url string) ContentPart {
	return ContentPart{Type: PartImageURL, URL: url}
}

// ImagePart returns raw image bytes (e.g. PNG, JPEG).
func ImagePart(mimeType string, data []byte) ContentPart {
	return ContentPart{Type: PartImage, MIMEType: mimeType, Data: data}
}

// FilePart returns a document or other file as raw bytes.
func FilePart(mimeType, name string, data []byte) ContentPart {
	return ContentPart{Type: PartFile, MIMEType: mimeType, Name: name, Data: data}
}

// ToolCallPart returns an assistant tool invocation block.
func ToolCallPart(call ToolCall) ContentPart {
	c := call
	return ContentPart{Type: PartToolCall, ToolCall: &c, Name: call.Name}
}

// ToolResultPart returns a tool result block (use with RoleTool messages).
func ToolResultPart(toolCallID, name, content string) ContentPart {
	return ContentPart{
		Type:       PartToolResult,
		ToolCallID: toolCallID,
		Name:       name,
		Text:       content,
	}
}

// JoinTextParts concatenates text parts; returns an error if a non-text part is present.
func JoinTextParts(parts []ContentPart) (string, error) {
	var out string
	for i, part := range parts {
		if part.Type != PartText {
			return "", fmt.Errorf("part at index %d: only text parts are allowed", i)
		}
		if part.Text == "" {
			return "", fmt.Errorf("part at index %d: text is required", i)
		}
		if out != "" {
			out += "\n"
		}
		out += part.Text
	}
	if out == "" {
		return "", errors.New("at least one text part is required")
	}
	return out, nil
}

// IsTextOnly reports whether all parts are plain text.
func IsTextOnly(parts []ContentPart) bool {
	for _, part := range parts {
		if part.Type != PartText {
			return false
		}
	}
	return len(parts) > 0
}

// TextFromParts concatenates text parts and ignores non-text parts.
func TextFromParts(parts []ContentPart) string {
	var b strings.Builder
	for _, part := range parts {
		if part.Type != PartText || part.Text == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(part.Text)
	}
	return b.String()
}
