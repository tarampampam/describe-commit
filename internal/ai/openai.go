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

const openAIDefaultBaseURL = "https://api.openai.com"

type OpenAI struct {
	httpClient        httpClient
	apiKey, modelName string
	baseURL           string
}

var _ Provider = (*OpenAI)(nil)

type (
	openaiOptions struct {
		HttpClient httpClient
		BaseURL    string
	}

	// OpenAIOption allows to customize the OpenAI provider.
	OpenAIOption func(*openaiOptions)
)

// WithOpenAIHttpClient sets the HTTP client for the OpenAI provider.
func WithOpenAIHttpClient(c httpClient) OpenAIOption {
	return func(o *openaiOptions) { o.HttpClient = c }
}

// WithOpenAIBaseURL overrides the default OpenAI API base URL.
// Use this to point the provider at an OpenAI-compatible endpoint, e.g. a local Ollama instance.
func WithOpenAIBaseURL(url string) OpenAIOption {
	return func(o *openaiOptions) { o.BaseURL = url }
}

// NewOpenAI creates a new OpenAI provider.
func NewOpenAI(apiKey, model string, opt ...OpenAIOption) *OpenAI { //nolint:dupl
	var opts openaiOptions

	for _, o := range opt {
		o(&opts)
	}

	var p = OpenAI{
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

	if opts.BaseURL != "" {
		p.baseURL = strings.TrimRight(opts.BaseURL, "/")
	}

	return &p
}

func (p *OpenAI) Query(
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
			return nil, errors.New("no response from the OpenAI API")
		}

		return &Response{Prompt: instructions, Answer: parts[0]}, nil
	}

	return &Response{Prompt: instructions, Answer: answer}, nil
}

// newRequest creates a new HTTP request for the OpenAI API.
func (p *OpenAI) newRequest(
	ctx context.Context,
	instructions, changes, commits string,
	o options,
) (*http.Request, error) {
	type message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	// https://platform.openai.com/docs/api-reference/chat
	j, jErr := json.Marshal(struct {
		Model               string    `json:"model"`
		Messages            []message `json:"messages"`
		Store               bool      `json:"store"`
		Temperature         float64   `json:"temperature"`
		TopP                float64   `json:"top_p"`
		HowMany             int       `json:"n"` // How many chat completion choices to generate for each input message
		MaxCompletionTokens int64     `json:"max_completion_tokens"`
	}{
		Model:               p.modelName,
		Store:               false,
		Temperature:         0.1, //nolint:mnd
		TopP:                0.1, //nolint:mnd
		HowMany:             1,
		MaxCompletionTokens: o.MaxOutputTokens,
		Messages: []message{
			{Role: "system", Content: instructions},
			{Role: "user", Content: wrapChanges(changes)},
			{Role: "user", Content: wrapCommits(commits)},
		},
	})
	if jErr != nil {
		return nil, jErr
	}

	base := openAIDefaultBaseURL
	if p.baseURL != "" {
		base = p.baseURL
	}

	req, rErr := http.NewRequestWithContext(ctx,
		http.MethodPost,
		base+"/v1/chat/completions",
		bytes.NewReader(j),
	)
	if rErr != nil {
		return nil, rErr
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))

	return req, nil
}

// responseToError converts the response from the OpenAI API to an error.
func (p *OpenAI) responseToError(resp *http.Response) error {
	// https://platform.openai.com/docs/guides/error-codes
	var body struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	retryable := resp.StatusCode == http.StatusTooManyRequests || // 429 - rate limit
		resp.StatusCode == http.StatusInternalServerError || // 500
		resp.StatusCode == http.StatusBadGateway || // 502
		resp.StatusCode == http.StatusServiceUnavailable || // 503
		resp.StatusCode == http.StatusGatewayTimeout // 504

	var err error

	if dErr := json.NewDecoder(resp.Body).Decode(&body); dErr == nil && body.Error.Message != "" {
		err = fmt.Errorf("OpenAI API error: %s (status code: %d)", body.Error.Message, resp.StatusCode)
	} else {
		err = fmt.Errorf("unexpected OpenAI API response status code: %d (%s)",
			resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	if retryable {
		return newRetryableError(err)
	}

	return err
}

// parseResponse parses the response from the OpenAI API.
func (p *OpenAI) parseResponse(resp *http.Response) (string, error) {
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

	if len(answer.Choices) == 0 {
		return "", errors.New("no response from the OpenAI API")
	}

	var texts = make([]string, 0, len(answer.Choices))

	for _, choice := range answer.Choices {
		if choice.Message.Content != "" {
			texts = append(texts, choice.Message.Content)
		}
	}

	return strings.Trim(strings.Join(texts, "\n"), "\n\t "), nil
}
