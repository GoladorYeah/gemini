package normalizer

import (
	"encoding/json"
	"fmt"
	"gemini/backend/internal/models"
	"log"
	"strings"
)

// Normalize processes the search query using an AI model to extract structured data.
func Normalize(req models.SearchRequest) (*models.NormalizedResponse, error) {
	prompt := fmt.Sprintf("Extract the product information from the following query. Return a JSON object with the fields 'title', 'category', and 'features'. Do not include any backticks or other special characters in the response. Query: %s", req.Query)

	jsonString, err := GenerateContent(prompt)
	if err != nil {
		return nil, err
	}

	log.Printf("Gemini raw response: %s", jsonString)

	// Remove backticks from the response
	jsonString = strings.TrimPrefix(jsonString, "```json\n")
	jsonString = strings.TrimSuffix(jsonString, "\n```")
	jsonString = strings.TrimSpace(jsonString)

	log.Printf("Gemini cleaned response: %s", jsonString)

	var normalizedResp models.NormalizedResponse
	err = json.Unmarshal([]byte(jsonString), &normalizedResp)
	if err != nil {
		return nil, err
	}

	return &normalizedResp, nil
}
