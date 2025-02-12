package ai

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/genai"
)

type Gemini struct {
	SafetySettings []*genai.SafetySetting
	client         *genai.Client
	model          string
}

var _ Provider = (*Gemini)(nil) // ensure the interface is implemented

func NewGemini(ctx context.Context, apiKey, model string) *Gemini {
	client, clientErr := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if clientErr != nil {
		panic(clientErr) // should never happen
	}

	return &Gemini{
		SafetySettings: []*genai.SafetySetting{
			{Category: genai.HarmCategoryDangerousContent, Threshold: genai.HarmBlockThresholdBlockLowAndAbove},
			{Category: genai.HarmCategoryHarassment, Threshold: genai.HarmBlockThresholdBlockLowAndAbove},
			{Category: genai.HarmCategoryHateSpeech, Threshold: genai.HarmBlockThresholdBlockLowAndAbove},
			{Category: genai.HarmCategorySexuallyExplicit, Threshold: genai.HarmBlockThresholdBlockLowAndAbove},
		},
		client: client,
		model:  model,
	}
}

func (p *Gemini) Query(ctx context.Context, query string, opts ...Option) (string, error) { //nolint:funlen
	var (
		opt     = options{}.Apply(opts...)
		prompts = generatePrompt(opt)
	)

	var instructions = make([]*genai.Part, 0, len(prompts))

	for _, prompt := range prompts {
		instructions = append(instructions, &genai.Part{Text: prompt})
	}

	var (
		temperature, topP            float64 = 0.2, 0.1
		candidateCount, maxOutTokens int64   = 1, 500
	)

	result, err := p.client.Models.GenerateContent(
		ctx,
		p.model,
		genai.Text(query),
		&genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{Parts: instructions, Role: "model"},
			SafetySettings:    p.SafetySettings,
			Temperature:       &temperature,
			TopP:              &topP,
			CandidateCount:    &candidateCount,
			MaxOutputTokens:   &maxOutTokens,
		},
	)
	if err != nil {
		return "", err
	}

	if len(result.Candidates) == 0 {
		return "", errors.New("no candidates found")
	}

	if result.Candidates[0].Content == nil {
		return "", errors.New("no content found")
	}

	var texts = make([]string, 0, len(result.Candidates[0].Content.Parts))

	for _, candidate := range result.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				if part.Thought {
					continue
				}

				texts = append(texts, part.Text)
			}
		}
	}

	var out = strings.Trim(strings.Join(texts, "\n"), "\n\t ")

	if opt.ShortMessageOnly {
		var parts = strings.Split(out, "\n")

		if len(parts) > 0 {
			return parts[0], nil
		} else {
			return "", errors.New("no response from the Gemini API")
		}
	}

	return out, nil
}
