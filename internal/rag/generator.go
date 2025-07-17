package rag

import (
	"context"
	"encoding/json"
)

// Generator defines a generic interface for any LLM implementation.
type Generator interface {
	// Generate returns a complete response in one go.
	Generate(
		ctx context.Context,
		prompt string,
		opts GenerateOptions,
	) (*GenerationResult, error)

	// GenerateStream streams the response token by token or chunk by chunk.
	GenerateStream(
		ctx context.Context,
		prompt string,
		opts GenerateOptions,
		onChunk func(GenerationChunk),
	) error

	// CountTokens estimates token count (for budgeting, splitting, etc.).
	CountTokens(text string) int
}

// GenerationChunk represents a single streamed chunk of output.
type GenerationChunk struct {
	Delta string          `json:"delta"`            // Newly generated token(s)
	IsLast bool           `json:"is_last"`          // Whether this is the final chunk
	Raw    json.RawMessage `json:"raw,omitempty"`   // Full raw API response (optional)
	
	// Reserved for future internal metadata / debugging
	_agentExtensions map[string]interface{} `json:"-"`
}
