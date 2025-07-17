package rag

// RetrieveOptions configures the context retrieval process
type RetrieveOptions struct {
	// TopK limits the number of returned chunks
	// Set to 0 for no limit (not recommended)
	TopK int `json:"top_k"`

	// ScoreThreshold sets minimum relevance score (0.0–1.0)
	// Chunks below this score are filtered out
	ScoreThreshold float64 `json:"score_threshold,omitempty"`

	// Filters narrow results by metadata:
	// - "author": "John Doe"
	// - "date_after": "2023-01-01"
	Filters map[string]interface{} `json:"filters,omitempty"`

	// Reserved for future agent-specific retrieval:
	// - "verify_with_tool": true
	// - "freshness_priority": 0.8
	_agentExtensions map[string]interface{} `json:"-"`
}

// GenerateOptions controls text generation behavior
type GenerateOptions struct {
	// Model specifies which LLM variant to use
	// Format is implementation-dependent
	Model string `json:"model,omitempty"`

	// Temperature controls randomness (0.0–2.0):
	// - 0.0 = deterministic
	// - 1.0 = default
	// - 2.0 = highly creative
	Temperature float64 `json:"temperature,omitempty"`

	// MaxTokens sets hard limit on output length
	MaxTokens int `json:"max_tokens,omitempty"`

	// StopSequences halt generation when encountered
	// Useful for tool integration and structured outputs
	StopSequences []string `json:"stop_sequences,omitempty"`

	// ResponseFormat specifies output structure
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`

	// Reserved for future agent control:
	// - "allow_tool_use": true
	// - "self_correction_attempts": 3
	_agentExtensions map[string]interface{} `json:"-"`
}

// QueryOptions configures the end-to-end RAG pipeline
type QueryOptions struct {
	// Retrieve configures context fetching
	Retrieve *RetrieveOptions `json:"retrieve,omitempty"`

	// Generate configures final output generation
	Generate *GenerateOptions `json:"generate,omitempty"`

	// Hybrid enables fallback to direct LLM when:
	// - Retrieval fails
	// - Retrieved scores are below threshold
	Hybrid bool `json:"hybrid,omitempty"`

	// Reserved for future agent orchestration:
	// - "max_retrieval_rounds": 2
	// - "tool_injection_seq": 3
	_agentExtensions map[string]interface{} `json:"-"`
}
