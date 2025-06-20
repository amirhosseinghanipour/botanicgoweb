package auth

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	echo "github.com/labstack/echo/v4"
)

// JWTAuth middleware (remains unchanged)
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
	lastSeen   time.Time // <-- ADDED: Track last access time for cleanup
	mu         sync.Mutex
}

var (
	buckets      = make(map[string]*tokenBucket)
	bucketsMutex sync.Mutex
)

const (
	rateLimitCapacity = 20
	rateLimitRefill   = 20
	rateLimitWindow   = 1 * time.Minute
	cleanupInterval   = 5 * time.Minute  // <-- ADDED: How often to run cleanup
	bucketMaxAge      = 15 * time.Minute // <-- ADDED: Max age of an inactive bucket
)

// Initialize a background goroutine to clean up old buckets
func init() {
	go cleanupBuckets()
}

func cleanupBuckets() {
	// Run this function on a ticker
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		<-ticker.C
		bucketsMutex.Lock()
		for ip, bucket := range buckets {
			// Check if the bucket has been inactive for longer than the max age
			if time.Since(bucket.lastSeen) > bucketMaxAge {
				delete(buckets, ip)
			}
		}
		bucketsMutex.Unlock()
	}
}

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
			// Update last seen time on every request
			bucket.lastSeen = time.Now()
			bucketsMutex.Unlock()

			bucket.mu.Lock()
			defer bucket.mu.Unlock()

			if time.Since(bucket.lastRefill) > rateLimitWindow {
				bucket.tokens = rateLimitRefill
				bucket.lastRefill = time.Now()
			}

			if bucket.tokens > 0 {
				bucket.tokens--
				return next(c)
			}

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

// IsAllowedOrigin (remains unchanged)
func IsAllowedOrigin(origin string) bool {
	allowed := []string{"http://localhost:5173"}
	for _, o := range allowed {
		if origin == o {
			return true
		}
	}
	return false
}
