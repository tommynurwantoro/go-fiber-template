package middleware

import (
	"time"

	"app/internal/pkg/formatter"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func LimiterConfig() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        20,
		Expiration: 15 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(
				formatter.NewErrorResponse(
					formatter.TooManyRequest,
					"Too many requests, please try again later",
					c.Get("traceId"),
				),
			)
		},
		SkipSuccessfulRequests: true,
	})
}
