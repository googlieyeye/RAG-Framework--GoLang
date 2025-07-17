package generator

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
	"ragframework/internal/rag"
)

type OpenAIGenerator struct {
	client *openai.Client
	model  string
}

func NewOpenAIGenerator(apiKey, model string) *OpenAIGenerator {
	client := openai.NewClient(apiKey)
	return &OpenAIGenerator{
		client: client,
		model:  model,
	}
}

func (g *OpenAIGenerator) Generate(ctx context.Context, prompt string, opts rag.GenerateOptions) (*rag.GenerationResult, error) {
	resp, err := g.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: g.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return nil, err
	}

	result := &rag.GenerationResult{
		Text: resp.Choices[0].Message.Content,
	}
	return result, nil
}
