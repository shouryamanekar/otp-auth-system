package cache

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func InitRedis() {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Fatal("REDIS_URL is not set")
	}

	// Parse Redis URL correctly
	parsedURL, err := url.Parse(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}

	// Extract password and host
	password, _ := parsedURL.User.Password()
	redisAddr := parsedURL.Host

	// Use TLS for secure connection (required for Heroku Redis)
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // Heroku Redis requires this sometimes
	}

	RDB = redis.NewClient(&redis.Options{
		Addr:      redisAddr,
		Password:  password,
		TLSConfig: tlsConfig, // Enable TLS for Heroku Redis
	})

	// Test Redis connection
	_, err = RDB.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	fmt.Println("Connected to Redis successfully!")
}

// BlacklistToken adds a JWT token to the blacklist with 24-hour expiry
func BlacklistToken(token string) error {
	return RDB.Set(context.Background(), "blacklist:"+token, "1", 24*time.Hour).Err()
}

// IsTokenBlacklisted checks if a token is in the blacklist
func IsTokenBlacklisted(token string) (bool, error) {
	exists, err := RDB.Exists(context.Background(), "blacklist:"+token).Result()
	return exists == 1, err
}

// RemoveExpiredTokens cleans up expired tokens (this runs periodically)
func RemoveExpiredTokens() {
	ctx := context.Background()

	// Get all keys that start with "blacklist:"
	iter := RDB.Scan(ctx, 0, "blacklist:*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()

		// Check if the key has expired
		ttl, err := RDB.TTL(ctx, key).Result()
		if err == nil && ttl <= 0 {
			fmt.Println("Removing expired token")
			RDB.Del(ctx, key)
		}
	}
}
