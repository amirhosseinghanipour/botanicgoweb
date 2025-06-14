package middleware

import (
	"net/http"
	"strings"

	"botanic/internal/auth"

	"github.com/labstack/echo/v4"
)

// Auth middleware checks for a valid JWT token in the Authorization header
func Auth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
		}

		// Check if the header has the Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
		}

		// Verify the token
		userID, err := auth.VerifyToken(parts[1])
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}

		// Set the user ID in the context
		c.Set("userID", userID)

		return next(c)
	}
}
