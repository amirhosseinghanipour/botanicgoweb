package db

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	ctx         = context.Background()
)

// InitializeRedis sets up the Redis client
func InitializeRedis() error {
	addr := getEnvOrDefault("REDIS_ADDR", "localhost:6379")
	password := getEnvOrDefault("REDIS_PASSWORD", "")
	dbStr := getEnvOrDefault("REDIS_DB", "0")
	db, err := strconv.Atoi(dbStr)
	if err != nil {
		return fmt.Errorf("invalid REDIS_DB value: %v", err)
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test the connection
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return nil
}

// getEnvOrDefault returns the environment variable value or a default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// CloseRedis closes the Redis connection
func CloseRedis() error {
	if redisClient != nil {
		return redisClient.Close()
	}
	return nil
}

// Set stores a value in Redis with an expiration time
func Set(key string, value interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return redisClient.Set(ctx, key, jsonData, expiration).Err()
}

// Get retrieves a value from Redis
func Get(key string, dest interface{}) error {
	val, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// Delete removes a key from Redis
func Delete(key string) error {
	return redisClient.Del(ctx, key).Err()
}

// Exists checks if a key exists in Redis
func Exists(key string) (bool, error) {
	n, err := redisClient.Exists(ctx, key).Result()
	return n > 0, err
}

// List operations
func LPush(key string, value interface{}) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return redisClient.LPush(ctx, key, jsonData).Err()
}

func LRange(key string, start, stop int64) ([]string, error) {
	return redisClient.LRange(ctx, key, start, stop).Result()
}

// Hash operations
func HSet(key string, field string, value interface{}) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return redisClient.HSet(ctx, key, field, jsonData).Err()
}

func HGet(key string, field string, dest interface{}) error {
	val, err := redisClient.HGet(ctx, key, field).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// Sorted Set operations
func ZAdd(key string, score float64, member interface{}) error {
	jsonData, err := json.Marshal(member)
	if err != nil {
		return err
	}
	return redisClient.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: jsonData,
	}).Err()
}

// ZRange retrieves members from a sorted set
func ZRange(key string, start, stop int64) ([]string, error) {
	vals, err := redisClient.ZRange(ctx, key, start, stop).Result()
	if err != nil {
		return nil, err
	}

	// Unmarshal each value from JSON
	var result []string
	for _, val := range vals {
		var unmarshaled string
		if err := json.Unmarshal([]byte(val), &unmarshaled); err != nil {
			// If unmarshaling fails, use the raw value
			result = append(result, val)
		} else {
			result = append(result, unmarshaled)
		}
	}

	return result, nil
}

func ZRem(key string, member interface{}) error {
	return redisClient.ZRem(ctx, key, member).Err()
}
