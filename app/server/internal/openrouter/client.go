package openrouter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log" // Import the log package
	"net/http"
	"os"
	"time"
)

// Model represents an OpenRouter model
type Model struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	ContextLength int     `json:"context_length"`
	Pricing       Pricing `json:"pricing"`
	Description   string  `json:"description,omitempty"`
}

// Pricing represents model pricing information
type Pricing struct {
	Prompt     string `json:"prompt"`
	Completion string `json:"completion"`
}

// ChatMessage represents a message in a chat conversation
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Client represents an OpenRouter API client
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new OpenRouter client
func NewClient() *Client {
	apiKey := os.Getenv("OPENROUTER_API_KEY")

	// This will print the first few chars of the key to verify it's loaded.
	if len(apiKey) > 5 {
		log.Printf("[OPENROUTER DEBUG] Using API Key starting with: \"%s...\"", apiKey[:5])
	} else {
		log.Printf("[OPENROUTER DEBUG] API Key is missing or too short!")
	}

	return &Client{
		baseURL: "https://openrouter.ai/api/v1",
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetAvailableModels fetches available models from OpenRouter
func (c *Client) GetAvailableModels() ([]Model, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/models", nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("User-Agent", "Botanic Chat/1.0")
	req.Header.Set("Referer", "https://botanic.chat")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if !isJSON(body) {
		return getFallbackModels(), nil
	}

	if resp.StatusCode != http.StatusOK {
		return getFallbackModels(), nil
	}

	var result struct {
		Data []struct {
			ID            string  `json:"id"`
			Name          string  `json:"name"`
			ContextLength int     `json:"context_length"`
			Pricing       Pricing `json:"pricing"`
			Description   string  `json:"description,omitempty"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if len(result.Data) == 0 {
		return getFallbackModels(), nil
	}

	models := make([]Model, len(result.Data))
	for i, m := range result.Data {
		models[i] = Model{
			ID:            m.ID,
			Name:          m.Name,
			ContextLength: m.ContextLength,
			Pricing:       m.Pricing,
			Description:   m.Description,
		}
	}

	return models, nil
}

// isJSON checks if the given byte slice is valid JSON
func isJSON(data []byte) bool {
	var v interface{}
	return json.Unmarshal(data, &v) == nil
}

// GetFreeModels filters the list of models to only include free ones
func GetFreeModels(models []Model) []Model {
	var freeModels []Model
	for _, model := range models {
		if model.Pricing.Prompt == "0" && model.Pricing.Completion == "0" {
			freeModels = append(freeModels, model)
		}
	}
	return freeModels
}

// GetChatCompletion gets a chat completion from OpenRouter
func (c *Client) GetChatCompletion(messages []ChatMessage, model string, temperature float64) (string, error) {
	if len(messages) > 0 {
		log.Printf("[OPENROUTER DEBUG] Sending message to AI: \"%s\"", messages[0].Content)
	}

	payload := struct {
		Model       string        `json:"model"`
		Messages    []ChatMessage `json:"messages"`
		Temperature float64       `json:"temperature"`
	}{
		Model:       model,
		Messages:    messages,
		Temperature: temperature,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("HTTP-Referer", "https://botanic.chat")
	req.Header.Set("X-Title", "Botanic Chat")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// --- FINAL DEBUGGING LINE ---
		// This will tell us if it's a network timeout or other connection error.
		log.Printf("[OPENROUTER ERROR] HTTP request failed: %v", err)
		return "", fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		// This will show us the error message from the OpenRouter API itself.
		log.Printf("[OPENROUTER ERROR] API returned non-200 status: %s, Body: %s", resp.Status, string(body))
		return "", fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return result.Choices[0].Message.Content, nil
}

// getFallbackModels returns a list of fallback models
func getFallbackModels() []Model {
	return []Model{
		{
			ID:            "mistralai/mistral-7b-instruct",
			Name:          "Mistral 7B Instruct",
			ContextLength: 8192,
			Pricing: Pricing{
				Prompt:     "0",
				Completion: "0",
			},
			Description: "A 7B parameter model fine-tuned for instruction following",
		},
		{
			ID:            "google/gemma-7b-it",
			Name:          "Gemma 7B",
			ContextLength: 8192,
			Pricing: Pricing{
				Prompt:     "0",
				Completion: "0",
			},
			Description: "Google's lightweight, open model for text generation",
		},
	}
}
