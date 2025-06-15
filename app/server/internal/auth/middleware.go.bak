package auth

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	echo "github.com/labstack/echo/v4"
)

func JWTAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
			}

			claims, err := ValidateToken(parts[1])
			if err != nil {
				if err == ErrExpiredToken {
					return echo.NewHTTPError(http.StatusUnauthorized, "token has expired")
				}
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			c.Set("user_id", claims.UserID)
			c.Set("email", claims.Email)

			return next(c)
		}
	}
}

type tokenBucket struct {
	tokens     int
	lastRefill time.Time
	mu         sync.Mutex
}

var (
	buckets      = make(map[string]*tokenBucket)
	bucketsMutex sync.Mutex
)

const (
	rateLimitCapacity = 20 // max tokens (requests) per window
	rateLimitRefill   = 20 // refill tokens per window
	rateLimitWindow   = 1 * time.Minute
)

func RateLimiter() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := getIP(c.Request())

			bucketsMutex.Lock()
			bucket, exists := buckets[ip]
			if !exists {
				bucket = &tokenBucket{tokens: rateLimitCapacity, lastRefill: time.Now()}
				buckets[ip] = bucket
			}
			bucketsMutex.Unlock()

			bucket.mu.Lock()

			if time.Since(bucket.lastRefill) > rateLimitWindow {
				bucket.tokens = rateLimitRefill
				bucket.lastRefill = time.Now()
			}
			if bucket.tokens > 0 {
				bucket.tokens--
				bucket.mu.Unlock()
				return next(c)
			}
			bucket.mu.Unlock()
			return echo.NewHTTPError(http.StatusTooManyRequests, "rate limit exceeded")
		}
	}
}

func getIP(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	return ip
}

// CORS and WebSocket Origin Check
func IsAllowedOrigin(origin string) bool {
	allowed := []string{"http://localhost:5173"}
	for _, o := range allowed {
		if origin == o {
			return true
		}
	}
	return false
}

func GetUserID(c echo.Context) (string, error) {
	userID, ok := c.Get("userID").(string)
	if !ok {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "user not authenticated")
	}
	return userID, nil
}
