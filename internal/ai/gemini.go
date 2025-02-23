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

// Gemini is a provider for the Gemini API.
type Gemini struct {
	httpClient        httpClient
	apiKey, modelName string
}

var _ Provider = (*Gemini)(nil) // ensure the interface is implemented

type (
	geminiOptions struct {
		HttpClient httpClient
	}

	// GeminiOption allows to customize the Gemini provider.
	GeminiOption func(*geminiOptions)
)

// WithGeminiHttpClient sets the HTTP client for the Gemini provider.
func WithGeminiHttpClient(c httpClient) GeminiOption {
	return func(o *geminiOptions) { o.HttpClient = c }
}

// NewGemini creates a new Gemini provider.
func NewGemini(apiKey, model string, opt ...GeminiOption) *Gemini {
	var opts geminiOptions

	for _, o := range opt {
		o(&opts)
	}

	var p = Gemini{
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

func (p *Gemini) Query(ctx context.Context, changes, commits string, opts ...Option) (*Response, error) { //nolint:dupl
	var (
		opt          = options{}.Apply(opts...)
		instructions = GeneratePrompt(opts...)
	)

	if opt.MaxOutputTokens == 0 {
		opt.MaxOutputTokens = defaultMaxOutputTokens // set default value
	}

	// https://ai.google.dev/gemini-api/docs/text-generation?lang=rest
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
			return nil, errors.New("no response from the Gemini API")
		}

		return &Response{Prompt: instructions, Answer: parts[0]}, nil
	}

	return &Response{Prompt: instructions, Answer: answer}, nil
}

// newRequest creates a new HTTP request for the Gemini API.
func (p *Gemini) newRequest( //nolint:funlen
	ctx context.Context,
	instructions, changes, commits string,
	o options,
) (*http.Request, error) {
	type (
		safetySetting struct {
			Category  string `json:"category"`
			Threshold string `json:"threshold"`
		}

		contentPart struct {
			Text string `json:"text"`
		}

		content struct {
			Parts []contentPart `json:"parts"`
		}
	)

	var data struct {
		GenerationConfig struct { // https://ai.google.dev/api/generate-content#v1beta.GenerationConfig
			Temperature     float64 `json:"temperature"`
			MaxOutputTokens int64   `json:"maxOutputTokens"`
			TopP            float64 `json:"topP"`
			CandidateCount  int     `json:"candidateCount"`
		} `json:"generationConfig"`
		SystemInstruction struct {
			Parts struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"system_instruction"`
		// https://ai.google.dev/api/generate-content#v1beta.SafetySetting
		SafetySettings []safetySetting `json:"safetySettings"`
		Contents       []content       `json:"contents"`
	}

	data.GenerationConfig.Temperature = 0.1
	data.GenerationConfig.MaxOutputTokens = o.MaxOutputTokens
	data.GenerationConfig.TopP = 0.1
	data.GenerationConfig.CandidateCount = 1

	data.SystemInstruction.Parts.Text = instructions

	data.SafetySettings = []safetySetting{
		{Category: "HARM_CATEGORY_DANGEROUS_CONTENT", Threshold: "BLOCK_LOW_AND_ABOVE"},
		{Category: "HARM_CATEGORY_HARASSMENT", Threshold: "BLOCK_LOW_AND_ABOVE"},
		{Category: "HARM_CATEGORY_HATE_SPEECH", Threshold: "BLOCK_LOW_AND_ABOVE"},
		{Category: "HARM_CATEGORY_SEXUALLY_EXPLICIT", Threshold: "BLOCK_LOW_AND_ABOVE"},
	}

	data.Contents = []content{{Parts: []contentPart{
		{Text: wrapChanges(changes)},
		{Text: wrapCommits(commits)},
	}}}

	j, jErr := json.Marshal(data)
	if jErr != nil {
		return nil, jErr
	}

	// https://ai.google.dev/gemini-api/docs/text-generation?lang=rest
	req, rErr := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent",
		p.modelName,
	), bytes.NewReader(j))
	if rErr != nil {
		return nil, rErr
	}

	req.Header.Set("Content-Type", "application/json")

	// https://cloud.google.com/docs/authentication/api-keys-use#using-with-rest
	req.Header.Set("x-goog-api-key", p.apiKey)

	return req, nil
}

// responseToError converts the response from the Gemini API to an error.
func (p *Gemini) responseToError(resp *http.Response) error {
	var response struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err == nil && response.Error.Message != "" {
		return fmt.Errorf(
			"gemini API error: %s (status code: %d)",
			response.Error.Message, resp.StatusCode,
		)
	}

	return fmt.Errorf(
		"unexpected Gemini API response status code: %d (%s)",
		resp.StatusCode, http.StatusText(resp.StatusCode),
	)
}

// parseResponse parses the response from the Gemini API.
func (p *Gemini) parseResponse(resp *http.Response) (string, error) {
	var answer struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if dErr := json.NewDecoder(resp.Body).Decode(&answer); dErr != nil {
		return "", dErr
	}

	if len(answer.Candidates) == 0 || len(answer.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("no content found")
	}

	var texts = make([]string, 0, len(answer.Candidates[0].Content.Parts))

	for _, candidate := range answer.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				texts = append(texts, part.Text)
			}
		}
	}

	return strings.Trim(strings.Join(texts, "\n"), "\n\t "), nil
}
