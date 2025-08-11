package main

import (
	"botanic/internal/auth"
	"botanic/internal/db"
	"botanic/internal/handlers"
	"botanic/internal/litellm" // <-- CHANGED
	"botanic/internal/middleware"
	"os"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	emiddleware "github.com/labstack/echo/v4/middleware"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	if err := db.InitializeRedis(); err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer db.CloseRedis()

	if err := auth.Initialize(); err != nil {
		log.Fatalf("Failed to initialize auth: %v", err)
	}

	// Initialize LiteLLM client
	liteLLMClient := litellm.NewClient() // <-- CHANGED

	e := echo.New()

	e.Use(emiddleware.Logger())
	e.Use(emiddleware.Recover())
    // CORS: allow localhost and optional ALLOWED_ORIGINS (comma-separated)
    allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")
    allowedOrigins := []string{"http://localhost:5173"}
    if allowedOriginsEnv != "" {
        // split by comma
        tmp := []string{}
        start := 0
        for i := 0; i <= len(allowedOriginsEnv); i++ {
            if i == len(allowedOriginsEnv) || allowedOriginsEnv[i] == ',' {
                if start < i {
                    tmp = append(tmp, allowedOriginsEnv[start:i])
                }
                start = i + 1
            }
        }
        if len(tmp) > 0 {
            allowedOrigins = append(allowedOrigins, tmp...)
        }
    }

    e.Use(emiddleware.CORSWithConfig(emiddleware.CORSConfig{
        AllowOrigins:     allowedOrigins,
        AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
        AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, echo.HeaderCookie, "X-CSRF-Token"},
        AllowCredentials: true,
        MaxAge:           300,
        ExposeHeaders:    []string{"Set-Cookie", "Authorization"},
    }))
	// Auth routes
	e.POST("/api/auth/register", handlers.Register)
	e.POST("/api/auth/login", handlers.Login)
	e.POST("/api/auth/verify", handlers.VerifyToken)
	e.POST("/api/auth/refresh", handlers.RefreshToken)
	e.POST("/api/auth/logout", handlers.Logout)
	e.GET("/api/auth/google", handlers.HandleGoogleAuth)
	e.GET("/api/auth/github", handlers.HandleGithubAuth)
	e.GET("/api/auth/:provider/callback", handlers.OAuthCallback)
	e.GET("/api/auth/profile", handlers.GetProfile, middleware.Auth)
	e.PUT("/api/auth/profile", handlers.UpdateProfile, middleware.Auth)
	e.PUT("/api/auth/preferences", handlers.UpdatePreferences, middleware.Auth)
	e.POST("/api/auth/avatar", handlers.UploadAvatar, middleware.Auth)

    // Public routes
    e.GET("/api/models", handlers.GetModels)
    e.POST("/api/demo/message", handlers.DemoMessage)

	// Chat routes
	chat := e.Group("/api/chat")
	chat.Use(middleware.Auth)
	chat.POST("/sessions", handlers.CreateSession)
	chat.GET("/sessions", handlers.GetSessions)
	chat.GET("/sessions/:id", handlers.GetSession)
	chat.DELETE("/sessions/:id", handlers.DeleteSession)
	chat.POST("/sessions/:id/messages", handlers.CreateMessage)

	// WebSocket endpoint
	e.GET("/ws", handlers.NewWSHandler(liteLLMClient).HandleWebSocket) // <-- CHANGED

	e.Logger.Fatal(e.Start(":8000"))
}
