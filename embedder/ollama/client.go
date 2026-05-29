package ollama

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/ollama/ollama/api"
	"github.com/singhJasvinder101/langchainai-go/init/config"
	"github.com/singhJasvinder101/langchainai-go/internal/constants"
)

func newAPIClient() (*api.Client, error) {
	baseURL := config.GetString(constants.ConfigOllamaBaseURL)
	if baseURL == "" {
		return api.ClientFromEnvironment()
	}

	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("ollama: invalid base_url %q: %w", baseURL, err)
	}

	return api.NewClient(parsed, http.DefaultClient), nil
}
