package handlers

import (
	"errors"
	"log"
	"net/http"
	"time"

	"botanic/internal/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

type CreateSessionRequest struct {
	Title string `json:"title"`
	Model string `json:"model"`
}

type CreateSessionResponse struct {
	Session *models.ChatSession `json:"session"`
	Message *models.Message     `json:"message"`
}

type CreateMessageRequest struct {
	Content string `json:"content"`
}

// CreateSession creates a new chat session with an optional initial message
func CreateSession(c echo.Context) error {
	userID, err := GetUserID(c)
	if err != nil {
		return err
	}

	var req CreateSessionRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Set default model if not provided
	if req.Model == "" {
		req.Model = "deepseek/deepseek-chat:free"
	}

	// Create session
	session, err := models.CreateChatSession(userID, req.Title, req.Model)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create session")
	}

	return c.JSON(http.StatusCreated, CreateSessionResponse{
		Session: session,
		Message: nil,
	})
}

// GetSession retrieves a chat session by ID
func GetSession(c echo.Context) error {
	userID, err := GetUserID(c)
	if err != nil {
		return err
	}

	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid session ID")
	}

	session, err := models.GetChatSession(sessionID.String())
	if err != nil {
		// Specifically check if the error is `redis: nil` (key not found)
		// and return a proper 404 Not Found error.
		if errors.Is(err, redis.Nil) {
			return echo.NewHTTPError(http.StatusNotFound, "session not found")
		}

		// For all other unexpected errors, log them and return a generic 500 error.
		log.Printf("ERROR getting chat session %s from models: %v", sessionID.String(), err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get session")
	}

	if session == nil {
		return echo.NewHTTPError(http.StatusNotFound, "session not found")
	}

	if session.UserID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "not authorized to access this session")
	}

	// Get messages for the session
	messages, err := models.GetSessionMessages(sessionID.String())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get messages")
	}

	// Create response with session and messages
	response := struct {
		ID        string            `json:"id"`
		UserID    string            `json:"user_id"`
		Title     string            `json:"title"`
		CreatedAt time.Time         `json:"created_at"`
		UpdatedAt time.Time         `json:"updated_at"`
		Messages  []*models.Message `json:"messages"`
	}{
		ID:        session.ID,
		UserID:    session.UserID,
		Title:     session.Title,
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
		Messages:  messages,
	}

	return c.JSON(http.StatusOK, response)
}

// GetUserID retrieves the user ID from the Echo context
func GetUserID(c echo.Context) (string, error) {
	userID, ok := c.Get("userID").(string)
	if !ok {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "user not authenticated")
	}
	return userID, nil
}

// GetSessions retrieves all chat sessions for the authenticated user
func GetSessions(c echo.Context) error {
	userID, err := GetUserID(c)
	if err != nil {
		return err
	}

	sessions, err := models.GetUserSessions(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get sessions")
	}

	// Create response with sessions and their messages
	var response []struct {
		ID        string            `json:"id"`
		UserID    string            `json:"user_id"`
		Title     string            `json:"title"`
		Model     string            `json:"model"`
		CreatedAt time.Time         `json:"created_at"`
		UpdatedAt time.Time         `json:"updated_at"`
		Messages  []*models.Message `json:"messages"`
	}

	for _, session := range sessions {
		messages, err := models.GetSessionMessages(session.ID)
		if err != nil {
			log.Printf("Failed to get messages for session %s: %v", session.ID, err)
			messages = []*models.Message{}
		}

		response = append(response, struct {
			ID        string            `json:"id"`
			UserID    string            `json:"user_id"`
			Title     string            `json:"title"`
			Model     string            `json:"model"`
			CreatedAt time.Time         `json:"created_at"`
			UpdatedAt time.Time         `json:"updated_at"`
			Messages  []*models.Message `json:"messages"`
		}{
			ID:        session.ID,
			UserID:    session.UserID,
			Title:     session.Title,
			Model:     "default", // Default model if not specified
			CreatedAt: session.CreatedAt,
			UpdatedAt: session.UpdatedAt,
			Messages:  messages,
		})
	}

	return c.JSON(http.StatusOK, response)
}

// DeleteSession deletes a chat session
func DeleteSession(c echo.Context) error {
	userID, err := GetUserID(c)
	if err != nil {
		return err
	}

	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid session ID")
	}

	session, err := models.GetChatSession(sessionID.String())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get session")
	}

	if session == nil {
		return echo.NewHTTPError(http.StatusNotFound, "session not found")
	}

	if session.UserID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "not authorized to delete this session")
	}

	if err := models.DeleteChatSession(sessionID.String()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete session")
	}

	return c.NoContent(http.StatusNoContent)
}

// CreateMessage creates a new message in a chat session
func CreateMessage(c echo.Context) error {
	userID, err := GetUserID(c)
	if err != nil {
		return err
	}

	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid session ID")
	}

	session, err := models.GetChatSession(sessionID.String())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get session")
	}

	if session == nil {
		return echo.NewHTTPError(http.StatusNotFound, "session not found")
	}

	if session.UserID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "not authorized to access this session")
	}

	var req CreateMessageRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	message, err := models.CreateMessage(sessionID.String(), "user", req.Content)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create message")
	}

	return c.JSON(http.StatusCreated, message)
}
