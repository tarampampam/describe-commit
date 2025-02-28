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

type AnthropicAPIResponse struct {
	Content    		[]AnthropicContent 	`json:"content"`
}

type AnthropicContent struct {
	Type   string  `json:"type"`
	Text   string  `json:"text,omitempty"`
}


type Anthropic struct {
	httpClient        httpClient
	apiKey, modelName, anthropicVersion string
}

var _ Provider = (*Anthropic)(nil)

type (
	AnthropicOptions struct {
		HttpClient httpClient
	}

	// AnthropicOption allows to customize the Anthropic provider.
	AnthropicOption func(*AnthropicOptions)
)

// WithAnthropicHttpClient sets the HTTP client for the Anthropic provider.
func WithAnthropicHttpClient(c httpClient) AnthropicOption {
	return func(o *AnthropicOptions) { o.HttpClient = c }
}

// NewAnthropic creates a new Anthropic provider.
func NewAnthropic(apiKey, model string, version string, opt ...AnthropicOption) *Anthropic {
	var opts AnthropicOptions

	for _, o := range opt {
		o(&opts)
	}

	var p = Anthropic{
		httpClient: opts.HttpClient,
		apiKey:     apiKey,
		modelName:  model,
		anthropicVersion: version,
	}

	if p.httpClient == nil { // set default HTTP client
		p.httpClient = &http.Client{
			Timeout:   60 * time.Second,                         //nolint:mnd
			Transport: &http.Transport{ForceAttemptHTTP2: true}, // use HTTP/2 (why not?)
		}
	}

	return &p
}

func (p *Anthropic) Query( //nolint:dupl
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
			return nil, errors.New("no response from the Anthropic API")
		}

		return &Response{Prompt: instructions, Answer: parts[0]}, nil
	}

	return &Response{Prompt: instructions, Answer: answer}, nil
}

// newRequest creates a new HTTP request for the Anthropic API.
func (p *Anthropic) newRequest(
	ctx context.Context,
	instructions, changes, commits string,
	o options,
) (*http.Request, error) {
	type message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	// https://docs.anthropic.com/en/api/messages
	j, jErr := json.Marshal(struct {
		Model               string    `json:"model"`
		Messages            []message `json:"messages"`
		Stream				bool	  `json:"stream"`
		Temperature         float64   `json:"temperature"`
		TopP                float64   `json:"top_p"`
		MaxTokens 			int64     `json:"max_tokens"`
		System				string	  `json:"system"`
	}{
		Model:               p.modelName,
		Stream:              false,
		System:				 instructions,
		Temperature:         0.1, //nolint:mnd
		TopP:                0.1, //nolint:mnd
		MaxTokens: 			 o.MaxOutputTokens,
		Messages: []message{
			{Role: "user", Content: wrapChanges(changes)},
			{Role: "user", Content: wrapCommits(commits)},
		},
	})
	if jErr != nil {
		return nil, jErr
	}

	// https://ai.google.dev/gemini-api/docs/text-generation?lang=rest
	req, rErr := http.NewRequestWithContext(ctx,
		http.MethodPost,
		"https://api.anthropic.com/v1/messages",
		bytes.NewReader(j),
	)
	if rErr != nil {
		return nil, rErr
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", p.anthropicVersion)

	return req, nil
}

// responseToError converts the response from the Anthropic API to an error.
func (p *Anthropic) responseToError(resp *http.Response) error {
	var response struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err == nil && response.Error.Message != "" {
		return fmt.Errorf(
			"Anthropic API error: %s (status code: %d)",
			response.Error.Message, resp.StatusCode,
		)
	}

	return fmt.Errorf(
		"unexpected Anthropic API response status code: %d (%s)",
		resp.StatusCode, http.StatusText(resp.StatusCode),
	)
}

// parseResponse parses the response from the Anthropic API.
func (p *Anthropic) parseResponse(resp *http.Response) (string, error) {
	var answer AnthropicAPIResponse
	if dErr := json.NewDecoder(resp.Body).Decode(&answer); dErr != nil {
		return "", dErr
	}

	if len(answer.Content) == 0 {
		return "", errors.New("no response from the Anthropic API")
	}

	var texts = make([]string, 0, len(answer.Content))

	for _, message := range answer.Content {
		if message.Text != "" {
			texts = append(texts, message.Text)
		}
	}

	return strings.Trim(strings.Join(texts, "\n"), "\n\t "), nil
}
