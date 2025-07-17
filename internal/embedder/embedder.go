package embedder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func EmbedText(text string) ([]float32, error) {
	body := map[string]interface{}{
		"inputs": []string{text}, // always send as array of strings
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", "http://localhost:8082/embed", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding request failed with status %d", resp.StatusCode)
	}

	var embeddings [][]float32
	err = json.NewDecoder(resp.Body).Decode(&embeddings)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if len(embeddings) == 0 || len(embeddings[0]) == 0 {
		return nil, fmt.Errorf("empty embedding returned")
	}

	return embeddings[0], nil
}



