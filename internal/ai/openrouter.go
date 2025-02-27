package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// OpenRouter is a provider for the OpenRouter API.
type OpenRouter struct {
	httpClient        httpClient
	apiKey, modelName string
}

var _ Provider = (*OpenRouter)(nil) // ensure the interface is implemented

type (
	openRouterOptions struct {
		HttpClient httpClient
	}

	// OpenRouterOption allows to customize the OpenRouter provider.
	OpenRouterOption func(*openRouterOptions)
)

// WithOpenRouterHttpClient sets the HTTP client for the OpenRouter provider.
func WithOpenRouterHttpClient(c httpClient) OpenRouterOption {
	return func(o *openRouterOptions) { o.HttpClient = c }
}

// NewOpenRouter creates a new OpenRouter provider.
func NewOpenRouter(apiKey, model string, opt ...OpenRouterOption) *OpenRouter {
	var opts openRouterOptions

	for _, o := range opt {
		o(&opts)
	}

	var p = OpenRouter{
		httpClient: opts.HttpClient,
		apiKey:     apiKey,
		modelName:  model,
	}

	if p.httpClient == nil { // set default HTTP client
		p.httpClient = &http.Client{
			Timeout:   60 * time.Second,                         //nolint:mnd
			Transport: &http.Transport{ForceAttemptHTTP2: true}, // use HTTP/2 (why not?)
		}
	}

	return &p
}

func (p *OpenRouter) Query( //nolint:dupl
	ctx context.Context,
	changes, commits string,
	opts ...Option,
) (*Response, error) {
	var (
		opt          = options{}.Apply(opts...)
		instructions = GeneratePrompt(opts...)
	)

	if opt.MaxOutputTokens == 0 {
		opt.MaxOutputTokens = defaultMaxOutputTokens // set default value
	}

	req, rErr := p.newRequest(ctx, instructions, changes, commits, opt)
	if rErr != nil {
		return nil, rErr
	}

	resp, rErr := p.httpClient.Do(req)
	if rErr != nil {
		return nil, rErr
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, p.responseToError(resp)
	}

	answer, aErr := p.parseResponse(resp)
	if aErr != nil {
		return nil, aErr
	}

	if opt.ShortMessageOnly {
		var parts = strings.Split(answer, "\n")

		if len(parts) == 0 {
			return nil, errors.New("no response from the OpenRouter API")
		}

		return &Response{Prompt: instructions, Answer: parts[0]}, nil
	}

	return &Response{Prompt: instructions, Answer: answer}, nil
}

// newRequest creates a new HTTP request for the OpenRouter API.
func (p *OpenRouter) newRequest(
	ctx context.Context,
	instructions, changes, commits string,
	o options,
) (*http.Request, error) {
	type message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	// https://openrouter.ai/docs/api-reference/parameters
	j, jErr := json.Marshal(struct {
		Model       string    `json:"model"`
		Messages    []message `json:"messages"`
		Temperature float64   `json:"temperature"`
		TopP        float64   `json:"top_p"`
		HowMany     int       `json:"n"` // How many chat completion choices to generate for each input message
		MaxTokens   int64     `json:"max_tokens"`
	}{
		Model:       p.modelName,
		Temperature: 0.1, //nolint:mnd
		TopP:        0.1, //nolint:mnd
		HowMany:     1,
		MaxTokens:   o.MaxOutputTokens,
		Messages: []message{
			{Role: "system", Content: instructions},
			{Role: "user", Content: wrapChanges(changes)},
			{Role: "user", Content: wrapCommits(commits)},
		},
	})
	if jErr != nil {
		return nil, jErr
	}

	req, rErr := http.NewRequestWithContext(ctx,
		http.MethodPost,
		"https://openrouter.ai/api/v1/chat/completions",
		bytes.NewReader(j),
	)
	if rErr != nil {
		return nil, rErr
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))

	return req, nil
}

// responseToError converts the response from the OpenRouter API to an error.
func (p *OpenRouter) responseToError(resp *http.Response) error {
	var response struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err == nil && response.Error.Message != "" {
		return fmt.Errorf(
			"OpenRouter API error: %s (status code: %d)",
			response.Error.Message, resp.StatusCode,
		)
	}

	return fmt.Errorf(
		"unexpected OpenRouter API response status code: %d (%s)",
		resp.StatusCode, http.StatusText(resp.StatusCode),
	)
}

// parseResponse parses the response from the OpenRouter API.
func (p *OpenRouter) parseResponse(resp *http.Response) (string, error) {
	var answer struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if dErr := json.NewDecoder(resp.Body).Decode(&answer); dErr != nil {
		return "", dErr
	}

	if len(answer.Choices) == 0 || len(answer.Choices[0].Message.Content) == 0 {
		return "", errors.New("no content found")
	}

	var texts = make([]string, 0, len(answer.Choices))

	for _, choice := range answer.Choices {
		if text := choice.Message.Content; text != "" {
			texts = append(texts, text)
		}
	}

	return strings.Trim(strings.Join(texts, "\n"), "\n\t "), nil
}
