package ai

import (
	"context"
	"net/http"
)

type (
	// Provider is an interface for AI providers.
	Provider interface {
		// Query the remote provider for the given string.
		Query(_ context.Context, changes, commits string, _ ...Option) (*Response, error)
	}

	// Response is a response from an AI provider.
	Response struct {
		Prompt string // used to generate the answer
		Answer string // what the AI responded
	}
)

const defaultMaxOutputTokens = 500

// httpClient is an interface for the common HTTP client.
type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Do not forget to update the [SupportedProviders] function if you add or remove providers.
const (
	ProviderGemini     = "gemini"
	ProviderOpenAI     = "openai"
	ProviderOpenRouter = "openrouter"
)

// SupportedProviders returns a list of supported AI providers.
func SupportedProviders() []string {
	return []string{ProviderGemini, ProviderOpenAI, ProviderOpenRouter}
}

// IsProviderSupported checks if the given provider is supported.
func IsProviderSupported(s string) bool {
	for _, p := range SupportedProviders() {
		if s == p {
			return true
		}
	}

	return false
}
