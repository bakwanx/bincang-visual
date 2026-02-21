package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
)

func OAuthMiddleware(redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from JWT (set by JWT middleware)
		userID, ok := c.Locals("userID").(string)
		if !ok || userID == "" {
			return c.Next()
		}

		token, err := getOAuthToken(redisClient, userID)
		if err == nil && token.Valid() {

			c.Locals("oauth_token", token)
		}

		return c.Next()
	}
}

func getOAuthToken(redisClient *redis.Client, userID string) (*oauth2.Token, error) {
	key := fmt.Sprintf("oauth_token:%s", userID)

	ctx := context.Background()
	tokenJSON, err := redisClient.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var token oauth2.Token
	if err := json.Unmarshal(tokenJSON, &token); err != nil {
		return nil, err
	}

	if token.AccessToken == "" {
		return nil, fmt.Errorf("invalid token: missing access token")
	}

	if !token.Expiry.IsZero() && token.Expiry.Before(time.Now()) {
		return nil, fmt.Errorf("token expired")
	}

	return &token, nil
}

func OAuthMiddlewareFromCookie() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenCookie := c.Cookies("oauth_token")
		if tokenCookie != "" {
			tokenJSON, err := base64.StdEncoding.DecodeString(tokenCookie)
			if err == nil {
				var token oauth2.Token
				if json.Unmarshal(tokenJSON, &token) == nil {
					c.Locals("oauth_token", &token)
				}
			}
		}
		return c.Next()
	}
}
