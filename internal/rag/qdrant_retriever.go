package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"ragframework/internal/embedder"
)

type QdrantRetriever struct {
	Host       string // Expect just "localhost:6333"
	Collection string
}

func NewQdrantRetriever(host string, collection string) *QdrantRetriever {
	return &QdrantRetriever{
		Host:       host,
		Collection: collection,
	}
}

type searchRequest struct {
	Vector      []float32 `json:"vector"`
	TopK        int       `json:"limit"`
	WithPayload bool      `json:"with_payload"`
}

type searchResponse struct {
	Result []struct {
		Payload map[string]interface{} `json:"payload"`
		Score   float64                `json:"score"`
	} `json:"result"`
}

func (qr *QdrantRetriever) Retrieve(ctx context.Context, query string, opts *RetrieveOptions) ([]ContextChunk, error) {
	topK := 5
	if opts != nil && opts.TopK > 0 {
		topK = opts.TopK
	}

	embedding, err := embedder.EmbedText(query)
	if err != nil {
		return nil, fmt.Errorf("embedding failed: %w", err)
	}

	reqBody := searchRequest{
		Vector:      embedding,
		TopK:        topK,
		WithPayload: true,
	}

	var bodyBuffer bytes.Buffer
	if err := json.NewEncoder(&bodyBuffer).Encode(reqBody); err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}

	url := fmt.Sprintf("http://%s/collections/%s/points/search", qr.Host, qr.Collection)
	req, err := http.NewRequestWithContext(ctx, "POST", url, &bodyBuffer)
	if err != nil {
		return nil, fmt.Errorf("request creation failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("qdrant request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("qdrant returned status code %d: %s", resp.StatusCode, string(body))
	}

	var parsed searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	fmt.Printf("🧲 Retrieved %d results from Qdrant\n", len(parsed.Result))
	for i, res := range parsed.Result {
		fmt.Printf("Result %d:\n  Score: %.4f\n  Payload: %+v\n", i+1, res.Score, res.Payload)
	}

	var chunks []ContextChunk
	for _, res := range parsed.Result {
		text, ok := res.Payload["text"].(string)
		if !ok {
			continue
		}
		chunks = append(chunks, ContextChunk{
			Text: text,
			Metadata: map[string]interface{}{
				"score": res.Score,
			},
		})
	}

	return chunks, nil
}

func (qr *QdrantRetriever) RetrieveStream(ctx context.Context, query string, opts *RetrieveOptions, onChunk func(ContextChunk)) error {
	chunks, err := qr.Retrieve(ctx, query, opts)
	if err != nil {
		return err
	}
	for _, chunk := range chunks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			onChunk(chunk)
		}
	}
	return nil
}


