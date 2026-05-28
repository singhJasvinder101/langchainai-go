package constants

const (
	ConfigLLMProvider = "llm.provider"

	ConfigLogLevel  = "log.level"
	ConfigLogFormat = "log.format"

	ConfigGeminiAPIKey = "gemini.api_key"
	ConfigGeminiModel  = "gemini.model"

	ConfigOpenAIAPIKey = "openai.api_key"
	ConfigOpenAIModel  = "openai.model"

	ConfigClaudeAPIKey    = "claude.api_key"
	ConfigClaudeModel     = "claude.model"
	ConfigClaudeMaxTokens = "claude.max_tokens"

	ConfigOllamaBaseURL = "ollama.base_url"
	ConfigOllamaModel   = "ollama.model"
)

const (
	DefaultOllamaBaseURL   = "http://127.0.0.1:11434"
	DefaultOllamaModel     = "smollm:135m"
	DefaultClaudeMaxTokens = 1024
)
