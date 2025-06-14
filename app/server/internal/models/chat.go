package models

import (
	"time"

	"botanic/internal/db"

	"github.com/google/uuid"
)

// Key prefixes for Redis
const (
	ChatPrefix    = "chat:"
	MessagePrefix = "message:"
)

// ChatSession represents a chat session
type ChatSession struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Message represents a chat message
type Message struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// NewChatSession creates a new chat session
func NewChatSession(userID string, title string, model string) *ChatSession {
	now := time.Now()
	return &ChatSession{
		ID:        uuid.New().String(),
		UserID:    userID,
		Title:     title,
		Model:     model,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewMessage creates a new message
func NewMessage(sessionID string, role string, content string) *Message {
	return &Message{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      role,
		Content:   content,
		CreatedAt: time.Now(),
	}
}

// CreateChatSession creates a new chat session
func CreateChatSession(userID string, title string, model string) (*ChatSession, error) {
	session := NewChatSession(userID, title, model)

	// Store session data
	sessionKey := ChatPrefix + session.ID
	if err := db.Set(sessionKey, session, 0); err != nil {
		return nil, err
	}

	// Add session to user's sessions
	userSessionsKey := ChatPrefix + "user:" + userID
	if err := db.ZAdd(userSessionsKey, float64(session.CreatedAt.Unix()), session.ID); err != nil {
		return nil, err
	}

	return session, nil
}

// GetUserSessions retrieves all chat sessions for a user
func GetUserSessions(userID string) ([]*ChatSession, error) {
	userSessionsKey := ChatPrefix + "user:" + userID
	sessionIDs, err := db.ZRange(userSessionsKey, 0, -1)
	if err != nil {
		return nil, err
	}

	var sessions []*ChatSession
	for _, sessionID := range sessionIDs {
		var session ChatSession
		sessionKey := ChatPrefix + sessionID
		if err := db.Get(sessionKey, &session); err != nil {
			return nil, err
		}
		sessions = append(sessions, &session)
	}

	return sessions, nil
}

// GetChatSession retrieves a chat session by ID
func GetChatSession(sessionID string) (*ChatSession, error) {
	var session ChatSession
	sessionKey := ChatPrefix + sessionID
	if err := db.Get(sessionKey, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

// DeleteChatSession deletes a chat session and its messages
func DeleteChatSession(sessionID string) error {
	session, err := GetChatSession(sessionID)
	if err != nil {
		return err
	}

	// Delete session data
	sessionKey := ChatPrefix + sessionID
	if err := db.Delete(sessionKey); err != nil {
		return err
	}

	// Remove session from user's sessions
	userSessionsKey := ChatPrefix + "user:" + session.UserID
	if err := db.ZRem(userSessionsKey, sessionID); err != nil {
		return err
	}

	// Delete all messages in the session
	sessionMessagesKey := MessagePrefix + "session:" + sessionID
	messageIDs, err := db.ZRange(sessionMessagesKey, 0, -1)
	if err != nil {
		return err
	}

	for _, messageID := range messageIDs {
		messageKey := MessagePrefix + messageID
		if err := db.Delete(messageKey); err != nil {
			return err
		}
	}

	// Delete the session's messages set
	if err := db.Delete(sessionMessagesKey); err != nil {
		return err
	}

	return nil
}

// CreateMessage creates a new message in a chat session
func CreateMessage(sessionID string, role string, content string) (*Message, error) {
	message := NewMessage(sessionID, role, content)

	// Store message data
	messageKey := MessagePrefix + message.ID
	if err := db.Set(messageKey, message, 0); err != nil {
		return nil, err
	}

	// Add message to session's messages
	sessionMessagesKey := MessagePrefix + "session:" + sessionID
	if err := db.ZAdd(sessionMessagesKey, float64(message.CreatedAt.Unix()), message.ID); err != nil {
		return nil, err
	}

	return message, nil
}

// GetSessionMessages retrieves all messages in a chat session
func GetSessionMessages(sessionID string) ([]*Message, error) {
	sessionMessagesKey := MessagePrefix + "session:" + sessionID
	messageIDs, err := db.ZRange(sessionMessagesKey, 0, -1)
	if err != nil {
		return nil, err
	}

	var messages []*Message
	for _, messageID := range messageIDs {
		var message Message
		messageKey := MessagePrefix + messageID
		if err := db.Get(messageKey, &message); err != nil {
			return nil, err
		}
		messages = append(messages, &message)
	}

	return messages, nil
}

// DeleteMessage deletes a message from a chat session
func DeleteMessage(messageID string) error {
	message, err := GetMessage(messageID)
	if err != nil {
		return err
	}

	// Delete message data
	messageKey := MessagePrefix + messageID
	if err := db.Delete(messageKey); err != nil {
		return err
	}

	// Remove message from session's messages
	sessionMessagesKey := MessagePrefix + "session:" + message.SessionID
	if err := db.ZRem(sessionMessagesKey, messageID); err != nil {
		return err
	}

	return nil
}

// GetMessage retrieves a message by ID
func GetMessage(messageID string) (*Message, error) {
	var message Message
	messageKey := MessagePrefix + messageID
	if err := db.Get(messageKey, &message); err != nil {
		return nil, err
	}

	return &message, nil
}
