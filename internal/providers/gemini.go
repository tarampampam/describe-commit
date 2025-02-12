package providers

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/genai"
)

type Gemini struct {
	client *genai.Client
	model  string
}

func NewGemini(ctx context.Context, apiKey, model string) *Gemini {
	client, clientErr := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if clientErr != nil {
		panic(clientErr) // should never happen
	}

	return &Gemini{
		client: client,
		model:  model,
	}
}

var geminiSafetySettings = []*genai.SafetySetting{
	{Category: genai.HarmCategoryDangerousContent, Threshold: genai.HarmBlockThresholdBlockLowAndAbove},
	{Category: genai.HarmCategoryHarassment, Threshold: genai.HarmBlockThresholdBlockLowAndAbove},
	{Category: genai.HarmCategoryHateSpeech, Threshold: genai.HarmBlockThresholdBlockLowAndAbove},
	{Category: genai.HarmCategorySexuallyExplicit, Threshold: genai.HarmBlockThresholdBlockLowAndAbove},
}

func (p *Gemini) Query(ctx context.Context, query string, opts ...Option) (string, error) {
	var (
		opt = options{}.Apply(opts...)

		temperature, topP            float64 = 0.2, 0.1
		candidateCount, maxOutTokens int64   = 1, 500
	)

	var modelInstructions = []*genai.Part{
		{Text: "Generate a concise and informative git commit message subject line based on the provided code diff"},
		{Text: "Use the conventional commit format (<type>(<scope>): <message>) where appropriate"},
		{Text: "Focus on the WHAT and WHY of the change"},
		{Text: "Keep the commit message within 72 characters for the first line"},
		{Text: "The message should be imperative (e.g., \"Fix bug\", \"Add feature\") and describe what the change does"},
		{Text: "If the change is complex, provide a short summary followed by a blank line and a more detailed explanation in bullet points"},
		{Text: "Do not add a period at the end of each line"},
	}

	if opt.ShortMessageOnly {
		modelInstructions = append(modelInstructions,
			&genai.Part{Text: "Return only the short commit message (usually the first line) without the detailed explanation"},
		)
	} else {
		modelInstructions = append(modelInstructions,
			&genai.Part{Text: "Avoid generic messages like \"Updated files\" or \"Fixed bugs\". Be specific"},
			&genai.Part{Text: "Use present tense (e.g., \"Fix\", \"Add\", \"Refactor\") instead of past tense (e.g., \"Fixed\", \"Added\", \"Refactored\")"},
		)
	}

	result, err := p.client.Models.GenerateContent(
		ctx,
		p.model,
		genai.Text(query),
		&genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{
				Parts: modelInstructions,
				Role:  "model",
			},
			SafetySettings:  geminiSafetySettings,
			Temperature:     &temperature,
			TopP:            &topP,
			CandidateCount:  &candidateCount,
			MaxOutputTokens: &maxOutTokens,
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

	var out = strings.Trim(strings.Join(texts, "\n"), "\n")

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
