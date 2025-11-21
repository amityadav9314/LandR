package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/amityadav/landr/pkg/pb/learning"
)

type Client struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: "https://api.groq.com/openai/v1/chat/completions",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type chatRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []choice `json:"choices"`
}

type choice struct {
	Message message `json:"message"`
}

// GenerateFlashcards sends the content to Groq and expects a JSON array of flashcards.
func (c *Client) GenerateFlashcards(content string) ([]*learning.Flashcard, error) {
	log.Printf("[AI] Starting flashcard generation, content length: %d", len(content))

	prompt := `
You are a helpful assistant that creates flashcards from text.
Analyze the following text and create 5 to 10 high-quality flashcards (Question and Answer pairs).
Return ONLY a raw JSON array of objects with "question" and "answer" fields. 
Do not include any markdown formatting (like json code blocks).
Do not include any other text.

Text:
` + content

	reqBody := chatRequest{
		Model: "openai/gpt-oss-120b", // Current Groq model (as of Nov 2024)
		Messages: []message{
			{Role: "user", Content: prompt},
		},
	}

	log.Printf("[AI] Using model: %s", reqBody.Model)

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("[AI] Failed to marshal request: %v", err)
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Printf("[AI] Failed to create HTTP request: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	log.Printf("[AI] Sending request to Groq API...")
	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("[AI] HTTP request failed: %v", err)
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[AI] Received response with status: %d", resp.StatusCode)

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("[AI] API error response: %s", string(bodyBytes))
		return nil, fmt.Errorf("api error: %d %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		log.Printf("[AI] Failed to decode response: %v", err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		log.Printf("[AI] No choices returned in response")
		return nil, fmt.Errorf("no choices returned")
	}

	rawContent := chatResp.Choices[0].Message.Content
	log.Printf("[AI] Raw AI response length: %d", len(rawContent))

	// Clean up potential markdown formatting if the model ignores instructions
	rawContent = cleanJSON(rawContent)
	log.Printf("[AI] Cleaned JSON length: %d", len(rawContent))

	var cards []*learning.Flashcard
	if err := json.Unmarshal([]byte(rawContent), &cards); err != nil {
		log.Printf("[AI] Failed to parse flashcards JSON: %v", err)
		log.Printf("[AI] Content was: %s", rawContent)
		return nil, fmt.Errorf("failed to parse flashcards json: %w. Content: %s", err, rawContent)
	}

	log.Printf("[AI] Successfully parsed %d flashcards", len(cards))
	return cards, nil
}

func cleanJSON(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return s
}
