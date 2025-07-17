package scripts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"

	"ragframework/internal/embedder"
	"ragframework/internal/reader"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate/entities/models"
)

// CreateSchema creates Weaviate schema (not needed for Qdrant)
func CreateSchema(client *weaviate.Client) {
	class := &models.Class{
		Class: "Document",
		Properties: []*models.Property{
			{
				Name:     "text",
				DataType: []string{"text"},
			},
		},
		Vectorizer: "none",
	}

	err := client.Schema().ClassCreator().WithClass(class).Do(context.Background())
	if err != nil {
		log.Println("⚠️ Schema might already exist or failed to create:", err)
	} else {
		log.Println("✅ Schema created successfully.")
	}
}

// UploadTexts uploads texts to either Weaviate or Qdrant
func UploadTexts(dbType string, texts []string, wClient *weaviate.Client) {
	for i, doc := range texts {
		vector, err := embedder.EmbedText(doc)
		if err != nil {
			log.Println("❌ Embedding failed:", err)
			continue
		}

		switch dbType {
		case "weaviate":
			_, err = wClient.Data().Creator().
				WithClassName("Document").
				WithProperties(map[string]interface{}{
					"text": doc,
				}).
				WithVector(vector).
				Do(context.Background())

			if err != nil {
				log.Println("❌ Failed to upload to Weaviate:", err)
			} else {
				log.Println("📄 Uploaded document to Weaviate:", doc)
			}

		case "qdrant":
	// Build correctly typed payload with float32 vectors
	point := Point{
		ID:      i,
		Vector:  vector,
		Payload: map[string]interface{}{"text": doc},
	}

	reqBody := UploadRequest{
		Points: []Point{point},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(reqBody); err != nil {
		log.Println("❌ Failed to encode Qdrant request:", err)
		continue
	}

	url := "http://localhost:6333/collections/documents/points?wait=true"
	resp, err := http.Post(url, "application/json", &buf)
	if err != nil {
		log.Println("❌ Failed to upload to Qdrant:", err)
		continue
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Println("📄 Uploaded document to Qdrant:", doc)
	} else {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("❌ Qdrant returned status %d: %s\n", resp.StatusCode, string(bodyBytes))
	}


		default:
			log.Println("❌ Unknown DB type:", dbType)
		}
	}
}

// UploadPDF extracts text from a PDF and returns it (does not upload).
func UploadPDF(filePath string) (string, error) {
	text, err := reader.ExtractTextFromPDF(filePath)
	if err != nil {
		return "", err
	}
	log.Println("📄 Extracted text from PDF")
	return text, nil
}

func generateRandomID() int64 {
	return rand.Int63()
}

type Point struct {
	ID      int                    `json:"id"`      // Use int or string for ID — Qdrant accepts both
	Vector  []float32              `json:"vector"`
	Payload map[string]interface{} `json:"payload"` // ✅ This is where you store the text
}

type UploadRequest struct {
	Points []Point `json:"points"`
}

func UploadSingleTextToQdrant(host, collection, text string, vector []float32) error {
	url := fmt.Sprintf("http://%s/collections/%s/points?wait=true", host, collection)

	payload := map[string]interface{}{
		"text": text, // ✅ This key must match retriever
	}

	point := Point{
		ID:      1,
		Vector:  vector,
		Payload: payload,
	}

	reqBody := UploadRequest{
		Points: []Point{point},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(reqBody); err != nil {
		return fmt.Errorf("❌ Failed to encode upload request: %w", err)
	}

	resp, err := http.Post(url, "application/json", &buf)
	if err != nil {
		return fmt.Errorf("❌ Upload request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("❌ Upload failed. Status: %d, Body: %s", resp.StatusCode, string(body))
	}

	fmt.Println("📄 Uploaded document to Qdrant:", text)
	return nil
}

func convertToFloat64(input []float32) []float64 {
	output := make([]float64, len(input))
	for i, v := range input {
		output[i] = float64(v)
	}
	return output
}
