package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"botanic/internal/auth"
	"botanic/internal/litellm"

	"github.com/google/uuid" // New import for UUID generation
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

// Message defines the structure for websocket messages.
type Message struct {
	ID        string    `json:"id,omitempty"` // Added: Unique message ID
	Type      string    `json:"type"`
	SessionID string    `json:"sessionId,omitempty"`
	UserID    string    `json:"userId,omitempty"`
	Role      string    `json:"role,omitempty"`
	Content   string    `json:"content"` // Changed: from json.RawMessage to string
	Model     string    `json:"model,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	// Note: UpdatedAt is not in the JSON tags here, but is in frontend Message interface.
	// Ensure consistency if you need UpdatedAt to be sent over WS.
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte // Buffered channel of outbound messages.
	room string      // session_id
}

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	rooms      map[string]map[*Client]bool
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	llmClient  *litellm.Client
	// For cancelling in-flight AI requests
	aiRequests   map[string]context.CancelFunc
	aiRequestMux sync.Mutex
}

func newHub(llmClient *litellm.Client) *Hub {
	return &Hub{
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		rooms:      make(map[string]map[*Client]bool),
		aiRequests: make(map[string]context.CancelFunc),
		llmClient:  llmClient,
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.rooms[client.room] == nil {
				h.rooms[client.room] = make(map[*Client]bool)
			}
			h.rooms[client.room][client] = true
			log.Printf("Client registered to room %s. Total clients in room: %d", client.room, len(h.rooms[client.room]))
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.rooms[client.room]; ok {
				delete(h.rooms[client.room], client)
				close(client.send)
				if len(h.rooms[client.room]) == 0 {
					delete(h.rooms, client.room)
					log.Printf("Room %s closed.", client.room)
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			// Handle 'stop' message (command, not to be broadcasted to clients)
			if message.Type == "stop" {
				h.aiRequestMux.Lock()
				if cancel, exists := h.aiRequests[message.SessionID]; exists {
					cancel()
					delete(h.aiRequests, message.SessionID)
				}
				h.aiRequestMux.Unlock()
				continue // Do not broadcast stop messages to clients
			}

			// Only broadcast messages intended for display (assistant responses, typing indicators)
			// This prevents echoing user messages back to themselves.
			if message.Role == "assistant" || message.Type == "typing" {
				h.mu.RLock()
				clientsInRoom := h.rooms[message.SessionID]
				h.mu.RUnlock()

				marshalledMsg, err := json.Marshal(message)
				if err != nil {
					log.Printf("Error marshalling broadcast message: %v", err)
					continue
				}

				// Broadcast to all clients in the room
				for client := range clientsInRoom {
					select {
					case client.send <- marshalledMsg:
					default:
						close(client.send)
						delete(clientsInRoom, client)
					}
				}
			}

			// If it's a user message, process it to get an AI response
			if message.Role == "user" {
				// Send typing indicator immediately
				typingMsg, _ := json.Marshal(&Message{
					ID:        uuid.New().String(),
					Type:      "typing",
					SessionID: message.SessionID,
					Role:      "assistant",
					CreatedAt: time.Now(),
				})
				h.mu.RLock()
				clientsInRoom := h.rooms[message.SessionID]
				h.mu.RUnlock()
				for client := range clientsInRoom {
					// Using a non-blocking send with select to avoid blocking hub.run() if client.send is full
					select {
					case client.send <- typingMsg:
					default:
						log.Printf("Warning: Client send channel for typing message is full for room %s", message.SessionID)
					}
				}

				ctx, cancel := context.WithCancel(context.Background())
				h.aiRequestMux.Lock()
				h.aiRequests[message.SessionID] = cancel
				h.aiRequestMux.Unlock()

				go func(ctx context.Context, msg *Message) {
					defer func() {
						h.aiRequestMux.Lock()
						delete(h.aiRequests, msg.SessionID)
						h.aiRequestMux.Unlock()
					}()

					// The incoming user message 'Content' field is already a string
					// due to the struct change, so no need for json.Unmarshal here.
					contentStr := msg.Content
					log.Printf("LITELLM DEBUG Sending message to model : %q", contentStr)

					aiResp, err := h.llmClient.GetChatCompletion(ctx, []litellm.ChatMessage{{Role: "user", Content: contentStr}}, msg.Model, 0.7)
					if err != nil {
						if ctx.Err() == context.Canceled {
							log.Printf("AI request for session %s was cancelled.", msg.SessionID)
							// Optionally send a "stop" message to the frontend if needed
							// h.broadcast <- &Message{Type: "stop", SessionID: msg.SessionID}
							return
						}
						log.Printf("AI completion error: %v", err)
						// TODO: Send an error message back to the client
						// errorMsg, _ := json.Marshal(map[string]string{"error": "Failed to get AI response"})
						// h.broadcast <- &Message{Type: "error", SessionID: msg.SessionID, Content: string(errorMsg), Role: "system"}
						return
					}

					log.Printf("Received response from LiteLLM: %s", aiResp)

					// aiResp is already a string, and Message.Content is now string.
					// No need to json.Marshal(aiResp) again unless aiResp itself is expected to be JSON string.
					// If aiResp from litellm.Client.GetChatCompletion is a plain string,
					// assign it directly. If it's a JSON string, ensure it's still treated as string.
					// Assuming GetChatCompletion returns a plain string:
					assistantMessage := &Message{
						ID:        uuid.New().String(), // Generate a unique ID for the assistant's message
						Type:      "message",
						SessionID: msg.SessionID,
						UserID:    "assistant", // This represents the AI assistant
						Content:   aiResp,      // Directly assign the string content
						Model:     msg.Model,
						CreatedAt: time.Now(),
						Role:      "assistant", // Set role to assistant
					}
					h.broadcast <- assistantMessage

				}(ctx, message)
			}
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, rawMessage, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Unexpected close error: %v", err)
			}
			break
		}
		var msg Message
		// Unmarshal the incoming message from frontend.
		// If frontend sends user message content as a simple string, it will be unmarshaled into msg.Content.
		// If frontend sends user message content as a nested JSON string, it will attempt to unmarshal.
		// To be safe, if frontend sends nested, you might need to adapt here or ensure frontend sends plain string content.
		if err := json.Unmarshal(rawMessage, &msg); err != nil {
			log.Printf("Error unmarshalling message from client: %v", err)
			continue
		}
		msg.SessionID = c.room // Ensure session ID is always from the URL param
		c.hub.broadcast <- &msg
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, message)
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

type WSHandler struct {
	hub *Hub
}

func NewWSHandler(llmClient *litellm.Client) *WSHandler {
	hub := newHub(llmClient)
	go hub.run()
	return &WSHandler{hub: hub}
}

func (wh *WSHandler) HandleWebSocket(c echo.Context) error {
	sessionID := c.QueryParam("session_id")
	token := c.QueryParam("token")
	if sessionID == "" || token == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing session_id or token")
	}

	if _, err := auth.VerifyToken(token); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return err
	}

	client := &Client{hub: wh.hub, conn: conn, send: make(chan []byte, 256), room: sessionID}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
	return nil
}
