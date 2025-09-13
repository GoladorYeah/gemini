package normalizer

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var (
	geminiClients      []*genai.GenerativeModel
	currentClientIndex int
	mu                 sync.Mutex
)

func InitGemini(apiKeys []string) {
	ctx := context.Background()
	for _, apiKey := range apiKeys {
		client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
		if err != nil {
			log.Printf("Failed to create Gemini client: %v", err)
			continue
		}
		// List models
		iter := client.ListModels(ctx)
		for {
			m, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Found model: %s", m.Name)
		}

		geminiClients = append(geminiClients, client.GenerativeModel("gemini-2.5-flash"))
	}
	if len(geminiClients) == 0 {
		log.Fatal("No valid Gemini API keys provided")
	}
}

func getNextClient() *genai.GenerativeModel {
	mu.Lock()
	defer mu.Unlock()
	client := geminiClients[currentClientIndex]
	currentClientIndex = (currentClientIndex + 1) % len(geminiClients)
	return client
}

func GenerateContent(prompt string) (string, error) {
	client := getNextClient()
	ctx := context.Background()
	resp, err := client.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	if part, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
		return string(part), nil
	}

	return "", fmt.Errorf("unexpected response format")
}
