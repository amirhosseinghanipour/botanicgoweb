package handlers

import (
	"net/http"
	"strconv"

	"botanic/internal/litellm" // <-- CHANGED

	"github.com/labstack/echo/v4"
)

// ModelsResponse represents the response structure for the models endpoint
type ModelsResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Free     []litellm.Model `json:"free"`    // <-- CHANGED
		NonFree  []litellm.Model `json:"nonFree"` // <-- CHANGED
		HasMore  bool            `json:"hasMore"`
		Page     int             `json:"page"`
		PageSize int             `json:"pageSize"`
	} `json:"data"`
	Error   string `json:"error,omitempty"`
	Details string `json:"details,omitempty"`
}

// GetModels handles the /api/models endpoint
func GetModels(c echo.Context) error {
	// Get pagination parameters (though we won't use them for local models)
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.QueryParam("pageSize"))
	if pageSize < 1 {
		pageSize = 50 // Default page size
	}

	// Get all models from LiteLLM
	client := litellm.NewClient() // <-- CHANGED
	allModels, err := client.GetAvailableModels()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch models from LiteLLM proxy")
	}

	// With LiteLLM + Ollama, all models are considered free.
	// We will place them all in the "Free" list and leave "NonFree" empty.
	var responseData struct {
		Free     []litellm.Model `json:"free"`
		NonFree  []litellm.Model `json:"nonFree"`
		HasMore  bool            `json:"hasMore"`
		Page     int             `json:"page"`
		PageSize int             `json:"pageSize"`
	}

	responseData.Free = allModels
	responseData.NonFree = []litellm.Model{} // Empty list for non-free models
	responseData.HasMore = false             // No pagination needed
	responseData.Page = 1
	responseData.PageSize = len(allModels)

	return c.JSON(http.StatusOK, ModelsResponse{
		Success: true,
		Data:    responseData,
	})
}
