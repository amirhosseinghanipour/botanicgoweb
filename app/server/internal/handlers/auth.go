package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"crypto/rand"
	"encoding/base64"

	"botanic/internal/auth"
	"botanic/internal/models"
	"net/url"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

var (
	googleOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_CALLBACK_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	githubOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GITHUB_CALLBACK_URL"),
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	RememberMe bool   `json:"remember_me"`
}

type AuthResponse struct {
	Token   string      `json:"token"`
	User    models.User `json:"user"`
	Session struct {
		ID        string    `json:"id"`
		ExpiresAt time.Time `json:"expires_at"`
	} `json:"session"`
}

type UpdateProfileRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50"`
	AvatarURL   string `json:"avatar_url"`
	Preferences struct {
		Theme string `json:"theme" binding:"required,oneof=light dark system"`
	} `json:"preferences"`
}

type UpdatePreferencesRequest struct {
	Theme         string `json:"theme"`
	Language      string `json:"language"`
	Timezone      string `json:"timezone"`
	Notifications bool   `json:"notifications"`
}

// SessionInfo represents a user's session information
type SessionInfo struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Device    string    `json:"device"`
	Location  string    `json:"location"`
}

type VerifyTokenRequest struct {
	Token string `json:"token"`
}

type VerifyTokenResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message,omitempty"`
}

type RefreshTokenRequest struct {
	Token string `json:"token"`
}

// Register handles user registration
func Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Check if user already exists
	existingUser, _ := models.GetUserByEmail(req.Email)
	if existingUser != nil {
		return echo.NewHTTPError(http.StatusConflict, "user already exists")
	}

	// Create new user
	user, err := models.CreateUser(req.Email, req.Password, "email", "", "", "")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create user")
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate token")
	}

	resp := AuthResponse{
		Token: token,
		User:  *user,
	}

	return c.JSON(http.StatusCreated, resp)
}

// Login handles user login
func Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Get user by email
	user, err := models.GetUserByEmail(req.Email)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}

	// Verify password
	if !user.VerifyPassword(req.Password) {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(30 * 24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}

	// Create a session
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	session, err := models.CreateUserSession(user.ID, expiresAt)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create session")
	}

	resp := AuthResponse{
		Token: tokenString,
		User:  *user,
	}
	resp.Session.ID = session.SessionID
	resp.Session.ExpiresAt = session.ExpiresAt

	return c.JSON(http.StatusOK, resp)
}

// RefreshToken handles token refresh
func RefreshToken(c echo.Context) error {
	var req RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.Token == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "missing token")
	}

	// Try to verify token and get user ID
	userID, err := auth.VerifyToken(req.Token)
	if err != nil {
		// If token is expired, try to extract user ID from claims without validation
		if err == auth.ErrExpiredToken {
			claims := &auth.Claims{}
			token, _ := jwt.ParseWithClaims(req.Token, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(os.Getenv("JWT_SECRET")), nil
			})
			if token != nil && claims.UserID != "" {
				userID = claims.UserID
			} else {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}
		} else {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}
	}

	// Get user from database
	user, err := models.GetUserByID(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "user not found")
	}

	// Generate new token
	newToken, err := auth.GenerateToken(user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate new token")
	}

	// Create a new session
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	session, err := models.CreateUserSession(user.ID, expiresAt)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create session")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"token": newToken,
		"session": map[string]interface{}{
			"id":         session.SessionID,
			"expires_at": session.ExpiresAt,
		},
	})
}

// Generate a random state string
func generateState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// HandleGoogleAuth initiates Google OAuth flow with state
func HandleGoogleAuth(c echo.Context) error {
	// Log OAuth configuration
	log.Printf("OAuth Configuration:")
	log.Printf("ClientID: %s", googleOAuthConfig.ClientID)
	log.Printf("RedirectURL: %s", googleOAuthConfig.RedirectURL)
	log.Printf("Scopes: %v", googleOAuthConfig.Scopes)

	state, err := generateState()
	if err != nil {
		log.Printf("Failed to generate state: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate state")
	}
	log.Printf("Generated OAuth state: %s", state)

	cookie := new(http.Cookie)
	cookie.Name = "oauth_state"
	cookie.Value = state
	cookie.HttpOnly = true
	cookie.Secure = true
	cookie.Path = "/"
	cookie.SameSite = http.SameSiteLaxMode
	cookie.MaxAge = 300 // 5 minutes
	c.SetCookie(cookie)
	log.Printf("Set OAuth state cookie")

	// Add additional parameters for Google OAuth
	opts := []oauth2.AuthCodeOption{
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce,
	}
	url := googleOAuthConfig.AuthCodeURL(state, opts...)
	log.Printf("Redirecting to Google OAuth URL: %s", url)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

// HandleGoogleCallback processes Google OAuth callback
func HandleGoogleCallback(c echo.Context) error {
	// Handle OAuth errors
	if err := c.QueryParam("error"); err != "" {
		return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=%s", os.Getenv("FRONTEND_URL"), url.QueryEscape(err)))
	}

	// Validate state parameter
	state := c.QueryParam("state")
	if state == "" {
		return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=%s", os.Getenv("FRONTEND_URL"), url.QueryEscape("missing_state")))
	}

	// Get the authorization code
	code := c.QueryParam("code")
	if code == "" {
		return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=%s", os.Getenv("FRONTEND_URL"), url.QueryEscape("missing_code")))
	}

	// Exchange code for token
	token, err := googleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=%s", os.Getenv("FRONTEND_URL"), url.QueryEscape("token_exchange_failed")))
	}

	// Get user info
	client := googleOAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=%s", os.Getenv("FRONTEND_URL"), url.QueryEscape("failed_to_get_user_info")))
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=%s", os.Getenv("FRONTEND_URL"), url.QueryEscape("failed_to_parse_user_info")))
	}

	// Check if user exists by provider ID
	existingUser, err := models.GetUserByProviderID("google", userInfo.ID)
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=%s", os.Getenv("FRONTEND_URL"), url.QueryEscape("failed_to_check_user")))
	}

	var user *models.User
	if existingUser != nil {
		user = existingUser
	} else {
		// Check if user exists by email
		existingUser, err = models.GetUserByEmail(userInfo.Email)
		if err != nil {
			return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=%s", os.Getenv("FRONTEND_URL"), url.QueryEscape("failed_to_check_user")))
		}

		if existingUser != nil {
			// Link provider to existing user
			if err := models.LinkProviderToUser(existingUser.ID, "google", userInfo.ID); err != nil {
				return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=%s", os.Getenv("FRONTEND_URL"), url.QueryEscape("failed_to_link_provider")))
			}
			user = existingUser
		} else {
			// Create new user
			user, err = models.CreateUser(userInfo.Email, "", "google", userInfo.ID, userInfo.Name, userInfo.Picture)
			if err != nil {
				return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=%s", os.Getenv("FRONTEND_URL"), url.QueryEscape("failed_to_create_user")))
			}
		}
	}

	// Generate JWT token
	tokenString, err := auth.GenerateToken(user.ID)
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=%s", os.Getenv("FRONTEND_URL"), url.QueryEscape("failed_to_generate_token")))
	}

	// Create auth response
	authResponse := AuthResponse{
		Token: tokenString,
		User:  *user,
	}

	// Encode the response data
	responseData, err := json.Marshal(authResponse)
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=%s", os.Getenv("FRONTEND_URL"), url.QueryEscape("failed_to_encode_response")))
	}

	// Base64 encode the data to safely pass it in URL
	encodedData := base64.StdEncoding.EncodeToString(responseData)

	// Redirect to frontend with the encoded data
	return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/auth/callback/complete?data=%s", os.Getenv("FRONTEND_URL"), url.QueryEscape(encodedData)))
}

// HandleGithubAuth initiates GitHub OAuth flow with state
func HandleGithubAuth(c echo.Context) error {
	state, err := generateState()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate state")
	}
	cookie := new(http.Cookie)
	cookie.Name = "oauth_state"
	cookie.Value = state
	cookie.HttpOnly = true
	cookie.Secure = true
	cookie.Path = "/"
	cookie.SameSite = http.SameSiteLaxMode
	cookie.MaxAge = 300 // 5 minutes
	c.SetCookie(cookie)
	url := githubOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

// HandleGithubCallback processes GitHub OAuth callback
func HandleGithubCallback(c echo.Context) error {
	// Handle OAuth errors
	if err := c.QueryParam("error"); err != "" {
		return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=%s", os.Getenv("FRONTEND_URL"), url.QueryEscape(err)))
	}

	// Validate state parameter
	state := c.QueryParam("state")
	if state == "" {
		return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=%s", os.Getenv("FRONTEND_URL"), url.QueryEscape("missing_state")))
	}

	// Get the authorization code
	code := c.QueryParam("code")
	if code == "" {
		return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=%s", os.Getenv("FRONTEND_URL"), url.QueryEscape("missing_code")))
	}

	// Exchange code for token
	token, err := githubOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=%s", os.Getenv("FRONTEND_URL"), url.QueryEscape("token_exchange_failed")))
	}

	// Get user info
	client := githubOAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=%s", os.Getenv("FRONTEND_URL"), url.QueryEscape("failed_to_get_user_info")))
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=%s", os.Getenv("FRONTEND_URL"), url.QueryEscape("failed_to_parse_user_info")))
	}

	// Check if user exists by provider ID
	existingUser, err := models.GetUserByProviderID("github", fmt.Sprintf("%d", userInfo.ID))
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=%s", os.Getenv("FRONTEND_URL"), url.QueryEscape("failed_to_check_user")))
	}

	var user *models.User
	if existingUser != nil {
		user = existingUser
	} else {
		// Check if user exists by email
		existingUser, err = models.GetUserByEmail(userInfo.Email)
		if err != nil {
			return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=%s", os.Getenv("FRONTEND_URL"), url.QueryEscape("failed_to_check_user")))
		}

		if existingUser != nil {
			// Link provider to existing user
			if err := models.LinkProviderToUser(existingUser.ID, "github", fmt.Sprintf("%d", userInfo.ID)); err != nil {
				return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=%s", os.Getenv("FRONTEND_URL"), url.QueryEscape("failed_to_link_provider")))
			}
			user = existingUser
		} else {
			// Create new user
			user, err = models.CreateUser(userInfo.Email, "", "github", fmt.Sprintf("%d", userInfo.ID), userInfo.Name, userInfo.AvatarURL)
			if err != nil {
				return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=%s", os.Getenv("FRONTEND_URL"), url.QueryEscape("failed_to_create_user")))
			}
		}
	}

	// Generate JWT token
	tokenString, err := auth.GenerateToken(user.ID)
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=%s", os.Getenv("FRONTEND_URL"), url.QueryEscape("failed_to_generate_token")))
	}

	// Create auth response
	authResponse := AuthResponse{
		Token: tokenString,
		User:  *user,
	}

	// Encode the response data
	responseData, err := json.Marshal(authResponse)
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=%s", os.Getenv("FRONTEND_URL"), url.QueryEscape("failed_to_encode_response")))
	}

	// Base64 encode the data to safely pass it in URL
	encodedData := base64.StdEncoding.EncodeToString(responseData)

	// Redirect to frontend with the encoded data
	return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/auth/callback/complete?data=%s", os.Getenv("FRONTEND_URL"), url.QueryEscape(encodedData)))
}

// GetProfile returns the user's profile information
func GetProfile(c echo.Context) error {
	// Get user ID from context (set by auth middleware)
	userID := c.Get("userID").(string)
	if userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "user not authenticated")
	}

	// Get user from database
	user, err := models.GetUserByID(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	return c.JSON(http.StatusOK, user)
}

// UpdateProfile updates the user's profile information
func UpdateProfile(c echo.Context) error {
	// Get user ID from context (set by auth middleware)
	userID := c.Get("userID").(string)
	if userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "user not authenticated")
	}

	var req UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Get user from database
	user, err := models.GetUserByID(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	// Update profile
	if err := user.UpdateProfile(req.Name, req.AvatarURL); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update profile")
	}

	// Update preferences
	user.Preferences.Theme = req.Preferences.Theme
	if err := user.UpdatePreferences(user.Preferences); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update preferences")
	}

	return c.JSON(http.StatusOK, user)
}

// UpdatePreferences updates the user's preferences
func UpdatePreferences(c echo.Context) error {
	// Get user ID from context (set by auth middleware)
	userID := c.Get("userID").(string)
	if userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "user not authenticated")
	}

	var req UpdatePreferencesRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Get user from database
	user, err := models.GetUserByID(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	// Update preferences
	user.Preferences.Theme = req.Theme
	user.Preferences.Language = req.Language
	user.Preferences.Timezone = req.Timezone
	user.Preferences.Notifications = req.Notifications

	if err := user.UpdatePreferences(user.Preferences); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update preferences")
	}

	return c.JSON(http.StatusOK, user.Preferences)
}

// UploadAvatar handles avatar file uploads
func UploadAvatar(c echo.Context) error {
	// Get user ID from context (set by auth middleware)
	userID := c.Get("userID").(string)
	if userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "user not authenticated")
	}

	// Get file from request
	file, err := c.FormFile("avatar")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid file upload")
	}

	// Validate file type
	if !strings.HasPrefix(file.Header.Get("Content-Type"), "image/") {
		return echo.NewHTTPError(http.StatusBadRequest, "file must be an image")
	}

	// Validate file size (max 5MB)
	if file.Size > 5*1024*1024 {
		return echo.NewHTTPError(http.StatusBadRequest, "file size must be less than 5MB")
	}

	// Create uploads directory if it doesn't exist
	uploadsDir := "uploads/avatars"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create uploads directory")
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filepath := filepath.Join(uploadsDir, filename)

	// Save file
	src, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to open uploaded file")
	}
	defer src.Close()

	dst, err := os.Create(filepath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save uploaded file")
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save uploaded file")
	}

	// Get user from database
	user, err := models.GetUserByID(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	// Delete old avatar if exists
	if user.AvatarURL != "" {
		oldPath := strings.TrimPrefix(user.AvatarURL, "/")
		os.Remove(oldPath)
	}

	// Update user's avatar URL
	avatarURL := fmt.Sprintf("/uploads/avatars/%s", filename)
	if err := user.UpdateProfile(user.Name, avatarURL); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update profile")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"avatar_url": avatarURL,
	})
}

// GetUserSessions returns a list of the user's active sessions
func GetUserSessions(c echo.Context) error {
	userID := c.Get("userID").(string)
	if userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "user not authenticated")
	}

	user, err := models.GetUserByID(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	sessions, err := models.GetUserActiveSessions(user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get sessions")
	}

	response := make([]SessionInfo, len(sessions))
	for i, session := range sessions {
		response[i] = SessionInfo{
			ID:        session.SessionID,
			CreatedAt: session.CreatedAt,
			ExpiresAt: session.ExpiresAt,
			Device:    c.Request().UserAgent(),
			Location:  c.Request().RemoteAddr,
		}
	}

	return c.JSON(http.StatusOK, response)
}

// DeleteUserSession deletes a user session
func DeleteUserSession(c echo.Context) error {
	userID := c.Get("userID").(string)
	if userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "user not authenticated")
	}

	sessionID := c.Param("id")
	if sessionID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing session ID")
	}

	if err := models.DeleteUserSession(userID, sessionID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete session")
	}

	return c.NoContent(http.StatusNoContent)
}

func AuthenticateWithProvider(provider, code, state string) (string, *models.User, error) {
	var config *oauth2.Config
	switch provider {
	case "google":
		config = googleOAuthConfig
	case "github":
		config = githubOAuthConfig
	default:
		return "", nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		return "", nil, fmt.Errorf("failed to exchange token: %v", err)
	}

	var userInfo struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}

	client := config.Client(context.Background(), token)

	if provider == "google" {
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			return "", nil, fmt.Errorf("failed to get user info: %v", err)
		}
		defer resp.Body.Close()

		if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
			return "", nil, fmt.Errorf("failed to decode user info: %v", err)
		}

		if !userInfo.VerifiedEmail {
			return "", nil, fmt.Errorf("email not verified")
		}
	} else {
		// GitHub user info
		resp, err := client.Get("https://api.github.com/user")
		if err != nil {
			return "", nil, fmt.Errorf("failed to get user info: %v", err)
		}
		defer resp.Body.Close()

		var githubUser struct {
			ID    int    `json:"id"`
			Email string `json:"email"`
			Name  string `json:"name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
			return "", nil, fmt.Errorf("failed to decode user info: %v", err)
		}

		// Get primary email if not provided
		if githubUser.Email == "" {
			emailsResp, err := client.Get("https://api.github.com/user/emails")
			if err != nil {
				return "", nil, fmt.Errorf("failed to get user emails: %v", err)
			}
			defer emailsResp.Body.Close()

			var emails []struct {
				Email    string `json:"email"`
				Primary  bool   `json:"primary"`
				Verified bool   `json:"verified"`
			}
			if err := json.NewDecoder(emailsResp.Body).Decode(&emails); err != nil {
				return "", nil, fmt.Errorf("failed to decode emails: %v", err)
			}

			for _, email := range emails {
				if email.Primary && email.Verified {
					githubUser.Email = email.Email
					break
				}
			}
		}

		userInfo.ID = fmt.Sprintf("%d", githubUser.ID)
		userInfo.Email = githubUser.Email
		userInfo.Name = githubUser.Name
		userInfo.VerifiedEmail = true
	}

	// Try to find user by provider ID
	user, err := models.GetUserByProviderID(provider, userInfo.ID)
	if err != nil || user == nil {
		// If not found, try to find by email
		existingUser, _ := models.GetUserByEmail(userInfo.Email)
		if existingUser != nil {
			err = models.LinkProviderToUser(existingUser.ID, provider, userInfo.ID)
			if err != nil {
				return "", nil, fmt.Errorf("failed to link provider: %v", err)
			}
			user = existingUser
		} else {
			user, err = models.CreateUser(userInfo.Email, "", provider, userInfo.ID, userInfo.Name, userInfo.Picture)
			if err != nil {
				return "", nil, fmt.Errorf("failed to create user: %v", err)
			}
		}
	}

	// Generate JWT token
	jwtToken, err := auth.GenerateToken(user.ID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %v", err)
	}

	return jwtToken, user, nil
}

func OAuthCallback(c echo.Context) error {
	provider := c.Param("provider")
	code := c.QueryParam("code")
	state := c.QueryParam("state")
	oauthErr := c.QueryParam("error")

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	if oauthErr != "" {
		log.Printf("OAuth error: %s", oauthErr)
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("%s/login?error=%s", frontendURL, url.QueryEscape(oauthErr)))
	}

	if code == "" || state == "" {
		log.Printf("Missing code or state. Code: %s, State: %s", code, state)
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("%s/login?error=Missing+code+or+state", frontendURL))
	}

	token, user, err := AuthenticateWithProvider(provider, code, state)
	if err != nil {
		log.Printf("Authentication failed: %v", err)
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("%s/login?error=%s", frontendURL, url.QueryEscape(err.Error())))
	}

	log.Printf("Authentication successful for user %s", user.Email)

	// Create auth response
	authResponse := AuthResponse{
		Token: token,
		User:  *user,
	}

	// Encode the response data
	responseData, err := json.Marshal(authResponse)
	if err != nil {
		log.Printf("Failed to encode response: %v", err)
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("%s/login?error=%s", frontendURL, url.QueryEscape("failed_to_encode_response")))
	}

	// Base64 encode the data to safely pass it in URL
	encodedData := base64.StdEncoding.EncodeToString(responseData)

	// Redirect to frontend with the encoded data
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("%s/auth/callback/complete?data=%s", frontendURL, url.QueryEscape(encodedData)))
}

// VerifyToken handles token verification
func VerifyToken(c echo.Context) error {
	// Get token from Authorization header
	token := c.Request().Header.Get("Authorization")
	if token == "" {
		return c.JSON(http.StatusUnauthorized, VerifyTokenResponse{
			Valid:   false,
			Message: "No token provided",
		})
	}

	// Remove "Bearer " prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// Parse and validate the token
	claims := &jwt.RegisteredClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !parsedToken.Valid {
		return c.JSON(http.StatusUnauthorized, VerifyTokenResponse{
			Valid:   false,
			Message: "Invalid or expired token",
		})
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return c.JSON(http.StatusUnauthorized, VerifyTokenResponse{
			Valid:   false,
			Message: "Token has expired",
		})
	}

	return c.JSON(http.StatusOK, VerifyTokenResponse{
		Valid:   true,
		Message: "Token is valid",
	})
}

// Logout handles user logout
func Logout(c echo.Context) error {
	// Get token from Authorization header
	token := c.Request().Header.Get("Authorization")
	if token == "" {
		return c.NoContent(http.StatusOK)
	}

	// Remove "Bearer " prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// Parse token to get user ID
	claims := &jwt.RegisteredClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !parsedToken.Valid {
		return c.NoContent(http.StatusOK)
	}

	// Get user ID from token
	userID := claims.Subject
	if userID == "" {
		return c.NoContent(http.StatusOK)
	}

	// Delete user session
	if err := models.DeleteUserSession(userID, token); err != nil {
		log.Printf("Failed to delete user session: %v", err)
	}

	return c.NoContent(http.StatusOK)
}
