# Configuration

`agentic-go` reads a single YAML file for API keys, models, and logging. Configuration is loaded once per process into a package-level, thread-safe singleton (`init/config`), so every `<provider>.New()` constructor can read it without you threading a config value through your code.

## Loading configuration

```go
import initializers "github.com/singhJasvinder101/agentic-go/init"

initializers.Init(ctx, "path/to/config.yaml")
```

`initializers.Init` initializes structured logging and loads the config file in one call. If you only need config (no logging setup), call the lower-level loader directly:

```go
import "github.com/singhJasvinder101/agentic-go/init/config"

config.MustInit("path/to/config.yaml")
```

Both are safe to call multiple times — the underlying load only runs once (`sync.Once`); a path passed on a later call is ignored. If no path is given, `config.DefaultConfigPath` (`configs/config.yaml`) is used.

## Example `config.yaml`

```yaml
log:
  level: "info"   # debug, info, warn, error
  format: "json"  # json, text

gemini:
  api_key: "YOUR_GEMINI_API_KEY"
  model: gemini-2.5-flash
  embedding_model: gemini-embedding-2

openai:
  api_key: "YOUR_OPENAI_API_KEY"
  model: gpt-4o-mini
  embedding_model: text-embedding-3-small

claude:
  api_key: "YOUR_CLAUDE_API_KEY"
  model: claude-sonnet-4-20250514
  max_tokens: 1024

ollama:
  base_url: http://127.0.0.1:11434
  model: smollm:135m
  embedding_model: all-minilm
```

Each `<provider>.New()` constructor (`llm/gemini`, `llm/openai`, `llm/claude`, `llm/ollama`, and their `embedder/` counterparts) reads its own section from this file — you only pay for the section you actually import.

## Reading values directly

```go
import "github.com/singhJasvinder101/agentic-go/init/config"

config.GetString("claude.model")   // "claude-sonnet-4-20250514"
config.GetInt("claude.max_tokens") // 1024
config.GetBool("some.flag")
```

Keys use dot-separated paths into the parsed YAML. All three getters return the zero value (`""`, `0`, `false`) if the path is missing or config hasn't been loaded yet — they never panic.

## See also

- [LLM providers](llm.md) — every provider's `New()` reads its config section this way
