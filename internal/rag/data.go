package rag

import (
	"time"
)

// ResponseFormat defines how generated outputs should be structured
type ResponseFormat struct {
	// Type specifies the output format (e.g., "json", "text", "markdown")
	Type string `json:"type"`

	// Schema defines the expected structure when Type="json"
	Schema interface{} `json:"schema,omitempty"`

	// Options contains format-specific settings like:
	// "indent": "true" (for JSON pretty printing)
	// "headers": "true" (for markdown section headers)
	Options map[string]string `json:"options,omitempty"`
}

// ContextChunk represents a unit of retrieved knowledge
type ContextChunk struct {
	// Text contains the actual content used for RAG
	Text string `json:"text"`

	// Metadata stores additional information about the chunk:
	// "source": "document.pdf" (origin document)
	// "score": 0.87 (relevance score)
	// "page": 42 (location in source)
	Metadata map[string]interface{} `json:"metadata"`

	// Embedding stores the vector representation (optional)
	// Used for advanced retrieval scenarios
	Embedding []float32 `json:"embedding,omitempty"`

	// Reserved for future agent capabilities like:
	// "confidence": 0.95
	// "temporal_validity": "2023-2024"
	_agentExtensions map[string]interface{} `json:"-"`
}

// GenerationResult contains the LLM's output
type GenerationResult struct {
	// Text is the generated content
	Text string `json:"text"`

	// Usage tracks computational resources consumed
	Usage *TokenUsage `json:"usage"`

	// FinishReason explains why generation stopped:
	// - "stop" (hit stop sequence)
	// - "length" (max tokens reached)
	// - "content_filter" (safety system)
	FinishReason string `json:"finish_reason,omitempty"`

	// Metadata contains model-specific information:
	// - "model": "gpt-4"
	// - "warnings": ["potential hallucination"]
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TokenUsage tracks token-level resource consumption
type TokenUsage struct {
	// InputTokens counts tokens in prompt/context
	Input int `json:"input_tokens"`

	// OutputTokens counts tokens in generated text
	Output int `json:"output_tokens"`

	// Total is Input + Output (convenience field)
	Total int `json:"total_tokens"`
}

// QueryResult contains the end-to-end RAG output
type QueryResult struct {
	// Answer is the final generated response
	Answer string `json:"answer"`

	// Contexts contains all retrieved chunks used
	Contexts []ContextChunk `json:"contexts,omitempty"`

	// Timestamp marks when generation completed
	Timestamp time.Time `json:"timestamp"`

	// Reserved for future agent workflow tracking:
	// - "reasoning_steps": [...]
	// - "tool_used": "calculator"
	_agentExtensions map[string]interface{} `json:"-"`
}


