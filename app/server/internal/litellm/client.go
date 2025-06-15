package litellm

import (
	"bytes"
	"context" // Import context package
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Model struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	ContextLength int     `json:"context_length"`
	Pricing       Pricing `json:"pricing"`
	Description   string  `json:"description,omitempty"`
}

// Pricing represents model pricing information.
// For local models, this will always be zero.
type Pricing struct {
	Prompt     string `json:"prompt"`
	Completion string `json:"completion"`
}

// ChatMessage represents a message in a chat conversation.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Client represents a LiteLLM API client.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new LiteLLM client.
func NewClient() *Client {
	// The client will connect to the LiteLLM proxy URL.
	// It defaults to http://localhost:4000 but can be overridden.
	baseURL := os.Getenv("LITELLM_URL")
	if baseURL == "" {
		baseURL = "http://localhost:4000"
	}
	log.Printf("[LITELLM DEBUG] Using proxy base URL: %s", baseURL)

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 90 * time.Second, // Increased timeout for local models
		},
	}
}

// GetAvailableModels fetches available models from the LiteLLM proxy.
func (c *Client) GetAvailableModels() ([]Model, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/v1/models", nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request to litellm proxy: %w. Is the proxy running?", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("litellm proxy returned non-200 status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// LiteLLM provides an OpenAI-compatible /models response.
	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error decoding response from litellm: %w", err)
	}

	// Adapt the response to the Model struct expected by the handlers.
	models := make([]Model, len(result.Data))
	for i, m := range result.Data {
		models[i] = Model{
			ID:            m.ID,
			Name:          m.ID, // Use ID as Name
			ContextLength: 8192, // Default context length
			Pricing: Pricing{ // All local models are free
				Prompt:     "0",
				Completion: "0",
			},
			Description: fmt.Sprintf("Locally hosted model: %s", m.ID),
		}
	}

	return models, nil
}

func (c *Client) GetChatCompletion(ctx context.Context, messages []ChatMessage, model string, temperature float64) (string, error) { // Add context.Context
	if len(messages) > 0 {
		log.Printf("[LITELLM DEBUG] Sending message to model %s: \"%s\"", model, messages[0].Content)
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

	// Create request with context for cancellation
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Check if the error is due to context cancellation
		if ctx.Err() == context.Canceled {
			return "", ctx.Err()
		}
		log.Printf("[LITELLM ERROR] HTTP request failed: %v", err)
		return "", fmt.Errorf("error making request to litellm proxy: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[LITELLM ERROR] API returned non-200 status: %s, Body: %s", resp.Status, string(body))
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
		return "", fmt.Errorf("no choices in response from litellm")
	}

	return result.Choices[0].Message.Content, nil
}
