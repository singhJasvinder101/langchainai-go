# Ai-GO
[AiGO Documentation](https://deepwiki.com/singhJasvinder101/ai-go)

`ai-go` is a Go library that provides a unified, streamlined interface for interacting with various Large Language Model (LLM) providers. It abstracts away provider-specific SDKs, allowing you to switch between models from OpenAI, Google Gemini, Anthropic Claude, and local Ollama instances with minimal code changes.

## Features

*   **Unified API:** A single `Provider` interface for all supported LLMs.
*   **Multiple Providers:** Out-of-the-box support for Gemini, OpenAI, Claude, and Ollama.
*   **Streaming Support:** A `GenerateStream` method for handling real-time, token-by-token responses.
*   **Simple Configuration:** Centralized YAML-based configuration for API keys, models, and other settings.
*   **Extensible Design:** The provider factory pattern makes it easy to add new LLM providers.

## Installation

```bash
go get github.com/singhJasvinder101/ai-go
```

## Configuration

The library is configured using a YAML file. By default, it looks for `configs/config.yaml`, but you can specify a custom path during initialization.

Create a `config.yaml` file with your provider settings.

### Example `config.yaml`

```yaml
# llm provider to use. options: gemini, openai, claude, ollama
llm:
  provider: "gemini"

# log settings
log:
  level: "info" # debug, info, warn, error
  format: "json"  # json, text

# gemini provider settings
gemini:
  api_key: "YOUR_GEMINI_API_KEY"
  model: "gemini-pro"

# openai provider settings
openai:
  api_key: "YOUR_OPENAI_API_KEY"
  model: "gpt-4-turbo"

# claude provider settings
claude:
  api_key: "YOUR_CLAUDE_API_KEY"
  model: "claude-3-opus-20240229"
  max_tokens: 1024

# ollama provider settings
ollama:
  # base_url is optional and defaults to http://127.0.0.1:11434
  base_url: "http://127.0.0.1:11434"
  model: "llama3"
```

## Usage

You can use `ai-go` for both single, blocking generation and for streaming responses.

### Basic Generation

This example initializes a provider based on your configuration and makes a single request.

```go
package main

import (
	"context"
	"fmt"
	"log"

	aigo "github.com/singhJasvinder101/ai-go/ai-go"
	"github.com/singhJasvinder101/ai-go/internal/constants"
	"github.com/singhJasvinder101/ai-go/internal/llm/providers/gemini"
)

func main() {
	ctx := context.Background()

	// Initialize a new provider from your config file.
	// The provider name must match a configured provider (e.g., "gemini").
	provider, err := aigo.New(ctx, constants.ProviderGemini, "path/to/your/config.yaml")
	if err != nil {
		log.Fatalf("failed to create provider: %v", err)
	}
	defer provider.Close()

	// Create a provider-specific request.
	req := &gemini.GenerateRequest{
		Prompt: "Hello, what is the capital of France?",
	}

	// Generate a response.
	resp, err := provider.Generate(ctx, req)
	if err != nil {
		log.Fatalf("failed to generate response: %v", err)
	}

	fmt.Println(resp.GetText())
}
```

### Streaming Generation

For real-time applications, you can stream the response token by token.

```go
package main

import (
	"context"
	"fmt"
	"log"

	aigo "github.com/singhJasvinder101/ai-go/ai-go"
	"github.com/singhJasvinder101/ai-go/internal/constants"
	"github.com/singhJasvinder101/ai-go/internal/llm/providers/openai"
)

func main() {
	ctx := context.Background()
	provider, err := aigo.New(ctx, constants.ProviderOpenAI, "path/to/your/config.yaml")
	if err != nil {
		log.Fatalf("Failed to initialize provider: %v", err)
	}
	defer provider.Close()

	req := &openai.GenerateRequest{
		Prompt: "Write a short story about a robot who discovers music.",
	}

	responses, errs := provider.GenerateStream(ctx, req)

	fmt.Println("Streaming response:")
	for {
		select {
		case resp, ok := <-responses:
			if !ok {
				responses = nil
			} else {
				fmt.Print(resp.GetText())
			}
		case err, ok := <-errs:
			if !ok {
				errs = nil
			} else if err != nil {
				log.Fatalf("Stream error: %v", err)
			}
		}
		if responses == nil && errs == nil {
			break
		}
	}
	fmt.Println()
}
```

## Supported Providers

*   **Gemini** (`gemini`)
*   **OpenAI** (`openai`)
*   **Claude** (`claude`)
*   **Ollama** (`ollama`)

## Example CLI

This repository includes a simple command-line application in `cmd/ai-go` that demonstrates the streaming API. After configuring your `config.yaml`, you can run it directly.

```bash
# Ensure your configs/config.yaml is present and configured
go run ./cmd/ai-go
```

## TODO (Future Support):
- [ ] Add tools support for agentic workflows
- [ ] Add mcp support
- [ ] Add other llm providers
