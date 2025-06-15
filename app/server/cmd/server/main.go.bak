package main

import (
	"botanic/internal/auth"
	"botanic/internal/db"
	"botanic/internal/handlers"
	"botanic/internal/middleware"
	"botanic/internal/openrouter"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	emiddleware "github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Initialize database
	if err := db.InitializeRedis(); err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer db.CloseRedis()

	// Initialize auth
	if err := auth.Initialize(); err != nil {
		log.Fatalf("Failed to initialize auth: %v", err)
	}

	// Initialize OpenRouter client
	openRouterClient := openrouter.NewClient()

	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(emiddleware.Logger())
	e.Use(emiddleware.Recover())
	e.Use(emiddleware.CORSWithConfig(emiddleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, echo.HeaderCookie, "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
		ExposeHeaders:    []string{"Set-Cookie", "Authorization"},
		AllowOriginFunc: func(origin string) (bool, error) {
			return origin == "http://localhost:5173", nil
		},
	}))

	/*
		e.Use(middleware.Logger()) e.Use(middleware.Recover()) e.Use(middleware.CORSWithConfig(middleware.CORSConfig{ AllowOrigins:     []string{"http://localhost:5173"}, AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions}, AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, echo.HeaderCookie, "X-CSRF-Token"}, AllowCredentials: true, MaxAge:           300, // Maximum value not ignored by any of major browsers ExposeHeaders:    []string{"Set-Cookie", "Authorization"}, AllowOriginFunc: func(origin string) (bool, error) { return origin == "http://localhost:5173", nil }, }))
	*/
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

	// Models routes
	e.GET("/api/models", handlers.GetModels)

	// Chat routes
	chat := e.Group("/api/chat")
	chat.Use(middleware.Auth)
	chat.POST("/sessions", handlers.CreateSession)
	chat.GET("/sessions", handlers.GetSessions)
	chat.GET("/sessions/:id", handlers.GetSession)
	chat.DELETE("/sessions/:id", handlers.DeleteSession)
	chat.POST("/sessions/:id/messages", handlers.CreateMessage)

	// WebSocket endpoint
	e.GET("/ws", handlers.NewWSHandler(openRouterClient).HandleWebSocket)

	// Start server
	e.Logger.Fatal(e.Start(":8000"))
}
