package middleware

import (
	"API/internal/utils"
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
)

func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")

		// Check token in Redis first (faster than JWT verification)
		tokenKey := "token:" + authHeader
		ctx := context.Background()

		// Try to get from Redis cache first
		if cachedClaims, err := utils.RedisClient.Get(ctx, tokenKey).Result(); err == nil {
			// Token found in cache, set user context
			c.Locals("user", cachedClaims)
			return c.Next()
		}

		// Not in cache, verify JWT
		claims, err := utils.ExtractTokenFromHeader(authHeader)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":  "Invalid token",
				"status": fiber.StatusUnauthorized,
			})
		}

		// Store in Redis for future requests
		utils.RedisClient.Set(ctx, tokenKey, claims, 24*time.Hour)

		c.Locals("user", claims)
		return c.Next()
	}
}
