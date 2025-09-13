package normalizer

import (
	"encoding/json"
	"fmt"
	"gemini/backend/internal/models"
	"log"
	"strings"
)

// ProductResponse represents the expected structure from Gemini
type ProductResponse struct {
	Title    string      `json:"title"`
	Category string      `json:"category"`
	Features interface{} `json:"features"` // Can be string or []string
}

// Normalize processes the search query using an AI model to extract structured data.
func Normalize(req models.SearchRequest) (*models.NormalizedResponse, error) {
	prompt := fmt.Sprintf("Extract the product information from the following query. Return a JSON object with the fields 'title', 'category', and 'features'. Do not include any backticks or other special characters in the response. Query: %s", req.Query)

	jsonString, err := GenerateContent(prompt)
	if err != nil {
		return nil, err
	}

	log.Printf("Gemini raw response: %s", jsonString)

	// Clean the response - remove any markdown formatting
	cleanResponse := strings.TrimSpace(jsonString)
	cleanResponse = strings.Trim(cleanResponse, "`")
	if strings.HasPrefix(cleanResponse, "json") {
		cleanResponse = strings.TrimPrefix(cleanResponse, "json")
		cleanResponse = strings.TrimSpace(cleanResponse)
	}

	// Remove backticks from the response
	cleanResponse = strings.TrimPrefix(cleanResponse, "```json\n")
	cleanResponse = strings.TrimSuffix(cleanResponse, "\n```")
	cleanResponse = strings.TrimSpace(cleanResponse)

	log.Printf("Gemini cleaned response: %s", cleanResponse)

	var productResp ProductResponse
	err = json.Unmarshal([]byte(cleanResponse), &productResp)
	if err != nil {
		log.Printf("Error unmarshaling Gemini response: %v", err)
		return nil, fmt.Errorf("failed to parse Gemini response: %v", err)
	}

	// Normalize the product
	normalized := &models.NormalizedResponse{
		Title:    productResp.Title,
		Category: productResp.Category,
		Features: []string{},
	}

	// Handle features field - can be string or array
	switch v := productResp.Features.(type) {
	case string:
		if v != "" && v != "Not specified in query" {
			// Split by common delimiters if it's a comma-separated string
			if strings.Contains(v, ",") {
				features := strings.Split(v, ",")
				for _, feature := range features {
					feature = strings.TrimSpace(feature)
					if feature != "" {
						normalized.Features = append(normalized.Features, feature)
					}
				}
			} else {
				normalized.Features = []string{v}
			}
		}
	case []interface{}:
		for _, feature := range v {
			if str, ok := feature.(string); ok && str != "" {
				normalized.Features = append(normalized.Features, strings.TrimSpace(str))
			}
		}
	case []string:
		for _, feature := range v {
			if feature != "" {
				normalized.Features = append(normalized.Features, strings.TrimSpace(feature))
			}
		}
	default:
		log.Printf("Unexpected features type: %T", v)
	}

	return normalized, nil
}
