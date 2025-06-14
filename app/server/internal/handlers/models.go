package handlers

import (
	"net/http"
	"strconv"

	"botanic/internal/openrouter"

	"github.com/labstack/echo/v4"
)

// ModelsResponse represents the response structure for the models endpoint
type ModelsResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Free     []openrouter.Model `json:"free"`
		NonFree  []openrouter.Model `json:"nonFree"`
		HasMore  bool               `json:"hasMore"`
		Page     int                `json:"page"`
		PageSize int                `json:"pageSize"`
	} `json:"data"`
	Error   string `json:"error,omitempty"`
	Details string `json:"details,omitempty"`
}

// GetModels handles the /api/models endpoint
func GetModels(c echo.Context) error {
	// Get pagination parameters
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.QueryParam("pageSize"))
	if pageSize < 1 {
		pageSize = 10 // Default page size
	}

	// Get all models from OpenRouter
	client := openrouter.NewClient()
	allModels, err := client.GetAvailableModels()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch models from OpenRouter")
	}

	// Separate free and non-free models
	var freeModels, nonFreeModels []openrouter.Model
	for _, model := range allModels {
		if model.Pricing.Prompt == "0" && model.Pricing.Completion == "0" {
			freeModels = append(freeModels, model)
		} else {
			nonFreeModels = append(nonFreeModels, model)
		}
	}

	// Calculate pagination for non-free models
	start := (page - 1) * pageSize
	end := start + pageSize
	hasMore := end < len(nonFreeModels)

	// If it's the first page, include free models
	var responseData struct {
		Free     []openrouter.Model `json:"free"`
		NonFree  []openrouter.Model `json:"nonFree"`
		HasMore  bool               `json:"hasMore"`
		Page     int                `json:"page"`
		PageSize int                `json:"pageSize"`
	}

	if page == 1 {
		responseData.Free = freeModels
	}

	// Get paginated non-free models
	if start < len(nonFreeModels) {
		if end > len(nonFreeModels) {
			end = len(nonFreeModels)
		}
		responseData.NonFree = nonFreeModels[start:end]
	}

	responseData.HasMore = hasMore
	responseData.Page = page
	responseData.PageSize = pageSize

	return c.JSON(http.StatusOK, ModelsResponse{
		Success: true,
		Data:    responseData,
	})
}
