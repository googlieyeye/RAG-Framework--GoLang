package main

import (
	"flag"
	"fmt"
	"log"
	"ragframework/internal/embedder"
	"ragframework/internal/llm"
	"ragframework/internal/retriever"
	"ragframework/scripts"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
)

func main() {
	// CLI Flags
	db := flag.String("db", "qdrant", "Database to use: 'qdrant' or 'weaviate'")
	text := flag.String("text", "", "Text to embed and upload")
	pdf := flag.String("pdf", "", "Path to PDF file to upload")
	query := flag.String("query", "", "User question for LLM to answer")
	llmProvider := flag.String("llm", "openai", "LLM provider to use: 'openai', etc.")
	host := flag.String("host", "localhost:6333", "Qdrant host (ignored for Weaviate)")
	collection := flag.String("collection", "documents", "Collection name for Qdrant")

	flag.Parse()

	// Validate DB
	if *db != "qdrant" && *db != "weaviate" {
		log.Fatalf("❌ Invalid DB: choose 'qdrant' or 'weaviate'")
	}

	// Initialize Weaviate client (if needed)
	var wClient *weaviate.Client
	if *db == "weaviate" {
		var err error
		wClient, err = scripts.InitWeaviateClient()
		if err != nil {
			log.Fatalf("❌ Failed to init Weaviate client: %v", err)
		}
	}

	// Handle PDF or Text Upload
	var doc string
	if *text != "" {
		doc = *text
	} else if *pdf != "" {
		var err error
		doc, err = scripts.UploadPDF(*pdf)
		if err != nil {
			log.Fatalf("❌ Failed to extract PDF: %v", err)
		}
	}

	// Upload text if present
	if doc != "" {
		vector, err := embedder.EmbedText(doc)
		if err != nil {
			log.Fatalf("❌ Embedding failed: %v", err)
		}
		switch *db {
		case "qdrant":
			err := scripts.UploadSingleTextToQdrant(*host, *collection, doc, vector)
			if err != nil {
				log.Fatalf("❌ Upload to Qdrant failed: %v", err)
			}
		case "weaviate":
			scripts.UploadTexts("weaviate", []string{doc}, wClient)
		}
		fmt.Println("✅ Upload complete!")
	}

	// Handle Retrieval + LLM generation if user provides a query
	if *query != "" {
		fmt.Println("🔍 Retrieving context from", *db)

		embeddedQuery, err := embedder.EmbedText(*query)
		if err != nil {
			log.Fatalf("❌ Failed to embed query: %v", err)
		}

		var contexts []string
		switch *db {
		case "qdrant":
			contexts, err = retriever.RetrieveFromQdrant(*host, *collection, embeddedQuery)
		case "weaviate":
			contexts, err = retriever.RetrieveFromWeaviate(wClient, embeddedQuery)
		}
		if err != nil {
			log.Fatalf("❌ Retrieval failed: %v", err)
		}

		if len(contexts) == 0 {
			log.Println("⚠️ No relevant documents found.")
		}

		// LLM generation
		fmt.Println("🧠 Generating answer using", *llmProvider)
		response, err := llm.GenerateAnswer(*query, contexts, *llmProvider)
		if err != nil {
			log.Fatalf("❌ LLM generation failed: %v", err)
		}

		fmt.Println("📣 Final Answer:\n", response)
	}
}

