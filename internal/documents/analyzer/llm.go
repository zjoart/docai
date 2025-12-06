package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type AnalysisResult struct {
	Summary  string                 `json:"summary"`
	Type     string                 `json:"type"`
	Metadata map[string]interface{} `json:"metadata"`
}

type Analyzer struct {
	client *openai.Client
	model  string
}

func NewAnalyzer(apiKey string) *Analyzer {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://openrouter.ai/api/v1"

	return &Analyzer{
		client: openai.NewClientWithConfig(config),
		model:  "gpt-4o-mini",
	}
}

func (a *Analyzer) AnalyzeText(ctx context.Context, text string) (*AnalysisResult, error) {
	if len(text) > 100000 {
		text = text[:100000]
	}

	prompt := fmt.Sprintf(`Analyze the following document text and return a JSON object with the following fields:
1. "summary": A concise summary of the document.
2. "type": The deduced document type (e.g., Invoice, CV, Report, Letter, Contract, Other).
3. "metadata": A flat JSON object containing extracted key fields (e.g., date, author, total_amount, invoice_number).

Return ONLY the JSON.

Document Text:
%s`, text)

	resp, err := a.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: a.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return nil, err
	}

	content := resp.Choices[0].Message.Content
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var result AnalysisResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %v, content: %s", err, content)
	}

	return &result, nil
}
