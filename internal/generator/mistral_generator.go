package generator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"ragframework/internal/rag"
)

// MistralGenerator implements the Generator interface using Ollama
type MistralGenerator struct {
	Model string
	Host  string // e.g., "http://localhost:11434"
}

func NewMistralGenerator(host, model string) *MistralGenerator {
	return &MistralGenerator{
		Host:  host,
		Model: model,
	}
}

// Generate implements one-shot completion
func (m *MistralGenerator) Generate(ctx context.Context, prompt string, opts rag.GenerateOptions) (*rag.GenerationResult, error) {
	reqBody := map[string]interface{}{
		"model":      m.Model,
		"prompt":     prompt,
		"temperature": opts.Temperature,
		"stop":        opts.StopSequences,
		"options": map[string]interface{}{
			"num_predict": opts.MaxTokens,
		},
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", m.Host+"/api/generate", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("request creation failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var fullOutput string
	decoder := json.NewDecoder(resp.Body)
	for decoder.More() {
		var chunk map[string]interface{}
		if err := decoder.Decode(&chunk); err != nil {
			return nil, fmt.Errorf("stream decode failed: %w", err)
		}
		if part, ok := chunk["response"].(string); ok {
			fullOutput += part
		}
	}

	return &rag.GenerationResult{
		Text: fullOutput,
	}, nil
}

// GenerateStream implements streaming output
func (m *MistralGenerator) GenerateStream(ctx context.Context, prompt string, opts rag.GenerateOptions, onChunk func(rag.GenerationChunk)) error {
	reqBody := map[string]interface{}{
		"model":      m.Model,
		"prompt":     prompt,
		"stream":     true,
		"temperature": opts.Temperature,
		"stop":        opts.StopSequences,
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", m.Host+"/api/generate", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("stream request creation failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("stream request failed: %w", err)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	for {
		var chunk map[string]interface{}
		if err := decoder.Decode(&chunk); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("stream decode error: %w", err)
		}

		part, _ := chunk["response"].(string)
		done, _ := chunk["done"].(bool)

		onChunk(rag.GenerationChunk{
			Delta:  part,
			IsLast: done,
			Raw:    nil, // Ollama doesn't give raw response by default
		})

		if done {
			break
		}
	}
	return nil
}

// CountTokens is a placeholder (you can implement real tokenizer later)
func (m *MistralGenerator) CountTokens(text string) int {
	return len([]rune(text)) / 4 // Rough estimate (Mistral ≈ 4 chars per token)
}
