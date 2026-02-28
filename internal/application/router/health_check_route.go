package router

import (
	"app/internal/application/handler"

	"github.com/gofiber/fiber/v2"
)

func HealthCheckRoutes(v1 fiber.Router, h handler.HealthCheckHandler) fiber.Router {
	healthCheck := v1.Group("/health-check")
	healthCheck.Get("/", h.Check)
	return healthCheck
}
