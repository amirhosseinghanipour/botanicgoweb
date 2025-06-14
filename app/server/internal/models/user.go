package models

import (
	"log"
	"net/http"
	"time"

	"botanic/internal/db"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

const (
	UserPrefix        = "user:"
	UserSessionPrefix = "user_session:"
	SessionPrefix     = "session:"
)

type User struct {
	ID           string          `json:"id"`
	Email        string          `json:"email"`
	PasswordHash string          `json:"-"`
	Provider     string          `json:"provider"`
	ProviderID   string          `json:"provider_id"`
	Name         string          `json:"name"`
	AvatarURL    string          `json:"avatar_url"`
	Preferences  UserPreferences `json:"preferences"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type UserPreferences struct {
	Theme         string `json:"theme"`
	Language      string `json:"language"`
	Timezone      string `json:"timezone"`
	Notifications bool   `json:"notifications"`
}

// CreateUser creates a new user in Redis
func CreateUser(email, password, provider, providerID, name, avatarURL string) (*User, error) {
	user := &User{
		ID:         uuid.New().String(),
		Email:      email,
		Provider:   provider,
		ProviderID: providerID,
		Name:       name,
		AvatarURL:  avatarURL,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	log.Printf("Creating user with ID: %s", user.ID)

	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Failed to hash password: %v", err)
			return nil, err
		}
		user.PasswordHash = string(hash)
	}

	user.Preferences = UserPreferences{
		Theme:         "system",
		Language:      "en",
		Timezone:      "UTC",
		Notifications: true,
	}

	userKey := UserPrefix + user.ID
	if err := db.Set(userKey, user, 0); err != nil {
		log.Printf("Failed to create user: %v", err)
		return nil, err
	}

	emailKey := UserPrefix + "email:" + email
	if err := db.Set(emailKey, user.ID, 0); err != nil {
		log.Printf("Failed to create email mapping: %v", err)
		return nil, err
	}

	if provider != "" && providerID != "" {
		providerKey := UserPrefix + "provider:" + provider + ":" + providerID
		if err := db.Set(providerKey, user.ID, 0); err != nil {
			log.Printf("Failed to create provider mapping: %v", err)
			return nil, err
		}
	}

	log.Printf("Successfully created user: %s", user.Email)
	return user, nil
}

// GetUserByEmail retrieves a user by email
func GetUserByEmail(email string) (*User, error) {
	emailKey := UserPrefix + "email:" + email
	var userID string
	if err := db.Get(emailKey, &userID); err != nil {
		return nil, err
	}

	userKey := UserPrefix + userID
	var user User
	if err := db.Get(userKey, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByProviderID retrieves a user by provider ID
func GetUserByProviderID(provider, providerID string) (*User, error) {
	providerKey := UserPrefix + "provider:" + provider + ":" + providerID
	var userID string
	if err := db.Get(providerKey, &userID); err != nil {
		return nil, err
	}

	userKey := UserPrefix + userID
	var user User
	if err := db.Get(userKey, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// VerifyPassword checks if the provided password matches the user's password hash
func (u *User) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// UpdateProfile updates the user's profile information
func (u *User) UpdateProfile(name, avatarURL string) error {
	u.Name = name
	u.AvatarURL = avatarURL
	u.UpdatedAt = time.Now()

	userKey := UserPrefix + u.ID
	return db.Set(userKey, u, 0)
}

// UpdatePreferences updates the user's preferences
func (u *User) UpdatePreferences(preferences UserPreferences) error {
	u.Preferences = preferences
	u.UpdatedAt = time.Now()

	userKey := UserPrefix + u.ID
	return db.Set(userKey, u, 0)
}

// LinkProviderToUser links an OAuth provider to an existing user
func LinkProviderToUser(userID, provider, providerID string) error {
	providerKey := UserPrefix + "provider:" + provider + ":" + providerID
	return db.Set(providerKey, userID, 0)
}

// GetUserByID retrieves a user by ID
func GetUserByID(id string) (*User, error) {
	userKey := UserPrefix + id
	var user User
	if err := db.Get(userKey, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// UserSession represents a user's active session
type UserSession struct {
	SessionID string    `json:"session_id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// GetUserActiveSessions retrieves all active sessions for a user
func GetUserActiveSessions(userID string) ([]UserSession, error) {
	sessionsKey := UserSessionPrefix + userID
	sessionIDs, err := db.ZRange(sessionsKey, 0, -1)
	if err != nil {
		return nil, err
	}

	var sessions []UserSession
	for _, sessionID := range sessionIDs {
		sessionKey := SessionPrefix + sessionID
		var session UserSession
		if err := db.Get(sessionKey, &session); err != nil {
			continue
		}
		if session.ExpiresAt.After(time.Now()) {
			sessions = append(sessions, session)
		}
	}

	return sessions, nil
}

// DeleteUserSession deletes a user session
func DeleteUserSession(userID, sessionID string) error {
	sessionsKey := UserSessionPrefix + userID
	sessionKey := SessionPrefix + sessionID

	if err := db.Delete(sessionsKey); err != nil {
		return err
	}

	return db.Delete(sessionKey)
}

// CreateUserSession creates a new user session
func CreateUserSession(userID string, expiresAt time.Time) (*UserSession, error) {
	session := &UserSession{
		SessionID: uuid.New().String(),
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
	}

	sessionKey := SessionPrefix + session.SessionID
	if err := db.Set(sessionKey, session, time.Until(expiresAt)); err != nil {
		return nil, err
	}

	sessionsKey := UserSessionPrefix + userID
	if err := db.ZAdd(sessionsKey, float64(session.CreatedAt.Unix()), session.SessionID); err != nil {
		return nil, err
	}

	return session, nil
}

// GetUserID retrieves the user ID from the Echo context
func GetUserID(c echo.Context) (string, error) {
	userID, ok := c.Get("userID").(string)
	if !ok {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "user not authenticated")
	}
	return userID, nil
}
