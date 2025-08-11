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
	client := openrouter.NewClient()
	allModels, err := client.GetAvailableModels()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch models from OpenRouter")
	}

	free := openrouter.GetFreeModels(allModels)

	// Build a set of free model IDs for quick lookup
	freeSet := make(map[string]struct{}, len(free))
	for _, m := range free {
		freeSet[m.ID] = struct{}{}
	}

	// Non-free are those not in freeSet
	nonFree := make([]openrouter.Model, 0, len(allModels)-len(free))
	for _, m := range allModels {
		if _, isFree := freeSet[m.ID]; !isFree {
			nonFree = append(nonFree, m)
		}
	}

    // Read pagination params for non-free models
    page := 1
    pageSize := 10
    if p := c.QueryParam("page"); p != "" {
        if v, err := strconv.Atoi(p); err == nil && v > 0 {
            page = v
        }
    }
    if ps := c.QueryParam("pageSize"); ps != "" {
        if v, err := strconv.Atoi(ps); err == nil && v > 0 {
            pageSize = v
        }
    }

    // Paginate non-free models only; free models are returned in full
    total := len(nonFree)
    start := (page - 1) * pageSize
    if start > total {
        start = total
    }
    end := start + pageSize
    if end > total {
        end = total
    }
    pagedNonFree := nonFree[start:end]
    hasMore := end < total

    var responseData struct {
        Free     []openrouter.Model `json:"free"`
        NonFree  []openrouter.Model `json:"nonFree"`
        HasMore  bool               `json:"hasMore"`
        Page     int                `json:"page"`
        PageSize int                `json:"pageSize"`
    }

    responseData.Free = free
    responseData.NonFree = pagedNonFree
    responseData.HasMore = hasMore
    responseData.Page = page
    responseData.PageSize = pageSize

	return c.JSON(http.StatusOK, ModelsResponse{
		Success: true,
		Data:    responseData,
	})
}
