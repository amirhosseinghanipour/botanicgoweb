package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"botanic/internal/auth"
	"botanic/internal/openrouter"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type Session struct {
	Conn         *websocket.Conn
	UserID       string
	LastActivity time.Time
}

type Message struct {
	Type      string    `json:"type"`
	SessionID string    `json:"sessionId"`
	UserID    string    `json:"userId"`
	Content   string    `json:"content"`
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"createdAt"`
}

type WSHandler struct {
	sessions     map[string]*Session
	sessionsLock sync.RWMutex
	client       *openrouter.Client
	upgrader     websocket.Upgrader
}

func NewWSHandler(client *openrouter.Client) *WSHandler {
	return &WSHandler{
		sessions: make(map[string]*Session),
		client:   client,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				// Allow all local development origins
				return origin == "" || // Allow requests with no origin
					origin == "http://localhost:5173" || // Vite dev server
					origin == "http://localhost:8000" || // Go dev server
					origin == "http://127.0.0.1:5173" ||
					origin == "http://127.0.0.1:8000" ||
					strings.HasPrefix(origin, "http://localhost:") || // Any localhost port
					strings.HasPrefix(origin, "http://127.0.0.1:") // Any loopback IP port
			},
		},
	}
}

func (h *WSHandler) HandleWebSocket(c echo.Context) error {
	sessionID := c.QueryParam("session_id")
	token := c.QueryParam("token")

	if sessionID == "" || token == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing session_id or token")
	}

	// The `auth.VerifyToken` function expects the raw token, without any "Bearer" prefix.
	userID, err := auth.VerifyToken(token)
	if err != nil {
		log.Printf("Token verification failed for WebSocket: %v", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	conn, err := h.upgrader.Upgrade(c.Response().Writer, c.Request(), nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return err
	}

	h.sessionsLock.Lock()
	if existingSession, exists := h.sessions[sessionID]; exists {
		existingSession.Conn.Close()
	}

	h.sessions[sessionID] = &Session{
		Conn:         conn,
		UserID:       userID,
		LastActivity: time.Now(),
	}
	h.sessionsLock.Unlock()

	go h.handleConnection(sessionID, conn)
	return nil
}

// THIS FUNCTION HAS BEEN UPDATED TO PREVENT THE UI FROM FREEZING
func (h *WSHandler) handleConnection(sessionID string, conn *websocket.Conn) {
	defer func() {
		conn.Close()
		h.sessionsLock.Lock()
		delete(h.sessions, sessionID)
		h.sessionsLock.Unlock()
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error reading message: %v", err)
			}
			break
		}

		// Update last activity time
		h.sessionsLock.Lock()
		if session, ok := h.sessions[sessionID]; ok {
			session.LastActivity = time.Now()
		}
		h.sessionsLock.Unlock()

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		// --- FIX: PROCESS AI REQUEST IN A NON-BLOCKING GOROUTINE ---

		// Send "typing..." indicator back to the UI immediately.
		typingMsg := Message{
			Type:      "typing",
			SessionID: sessionID,
			UserID:    msg.UserID, // Use the user ID from the incoming message
			Model:     msg.Model,
			CreatedAt: time.Now(),
		}
		// It's safe to write from the main loop.
		if err := conn.WriteJSON(typingMsg); err != nil {
			log.Printf("Error sending typing indicator: %v", err)
		}

		// Handle the potentially long API call in a separate goroutine
		// so it doesn't block the main ReadMessage loop.
		go func(incomingMsg Message) {
			// Get response from OpenRouter
			response, err := h.client.GetChatCompletion([]openrouter.ChatMessage{
				{
					Role:    "user",
					Content: incomingMsg.Content,
				},
			}, incomingMsg.Model, 0.7)

			// Check if the session still exists before sending a reply
			h.sessionsLock.RLock()
			_, exists := h.sessions[sessionID]
			h.sessionsLock.RUnlock()

			if !exists {
				log.Printf("Session %s closed before AI response could be sent.", sessionID)
				return // The user disconnected, so stop.
			}

			if err != nil {
				log.Printf("Error from OpenRouter: %v", err)
				errorMsg := Message{
					Type:      "error",
					SessionID: sessionID,
					UserID:    incomingMsg.UserID,
					Content:   "Failed to get response from AI model.",
					CreatedAt: time.Now(),
				}
				// The Gorilla WebSocket package handles concurrent writes, so this is safe.
				conn.WriteJSON(errorMsg)
				return // End the goroutine
			}

			// Send the AI's actual response
			assistantMsg := Message{
				Type:      "message",
				SessionID: sessionID,
				UserID:    "assistant",
				Content:   response,
				Model:     incomingMsg.Model,
				CreatedAt: time.Now(),
			}
			conn.WriteJSON(assistantMsg)
		}(msg) // Pass the `msg` by value to the goroutine to prevent data races
	}
}
