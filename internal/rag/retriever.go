package rag

import "context"

type Retriever interface {
	Retrieve(ctx context.Context, query string, opts *RetrieveOptions) ([]ContextChunk, error)
	RetrieveStream(ctx context.Context, query string, opts *RetrieveOptions, onChunk func(ContextChunk)) error
}
