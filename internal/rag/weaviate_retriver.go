package rag

import (
	"context"
	"fmt"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
)

type WeaviateRetriever struct {
	Client    *weaviate.Client
	ClassName string
}

func NewWeaviateRetriever(host string, className string) (*WeaviateRetriever, error) {
	cfg := weaviate.Config{
		Host:   host,
		Scheme: "http",
	}

	client, err := weaviate.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create weaviate client: %w", err)
	}

	return &WeaviateRetriever{
		Client:    client,
		ClassName: className,
	}, nil
}

func (wr *WeaviateRetriever) Retrieve(ctx context.Context, query string, opts *RetrieveOptions) ([]ContextChunk, error) {
	topK := 5
	if opts != nil && opts.TopK > 0 {
		topK = opts.TopK
	}

	result, err := wr.Client.GraphQL().Get().
		WithClassName(wr.ClassName).
		WithFields(
			graphql.Field{Name: "text"},
			graphql.Field{
				Name: "_additional",
				Fields: []graphql.Field{
					{Name: "certainty"},
				},
			},
		). // 🟢 This period was missing
		WithNearText(wr.Client.GraphQL().NearTextArgBuilder().
			WithConcepts([]string{query}),
		).
		WithLimit(topK).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("weaviate query failed: %w", err)
	}

	rawDocs, ok := result.Data["Get"].(map[string]interface{})[wr.ClassName].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected format from weaviate response")
	}

	var chunks []ContextChunk
	for _, doc := range rawDocs {
		item := doc.(map[string]interface{})
		additional := item["_additional"].(map[string]interface{})

		chunks = append(chunks, ContextChunk{
			Text: item["text"].(string),
			Metadata: map[string]interface{}{
				"certainty": additional["certainty"],
			},
		})
	}
	return chunks, nil
}

func (wr *WeaviateRetriever) RetrieveStream(ctx context.Context, query string, opts *RetrieveOptions, onChunk func(ContextChunk)) error {
	chunks, err := wr.Retrieve(ctx, query, opts)
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
