package bootstrap

import (
	"app/internal/application/handler"
	"app/internal/application/router"
	"app/internal/application/service"
	"app/internal/pkg/middleware"
)

func RegisterServices() {
	appContainer.RegisterService("healthCheckService", new(service.HealthCheckServiceImpl))
	appContainer.RegisterService("authService", new(service.AuthServiceImpl))
	appContainer.RegisterService("userService", new(service.UserServiceImpl))
	appContainer.RegisterService("tokenService", new(service.TokenServiceImpl))
}

func RegisterMiddleware() {
	appContainer.RegisterService("authMiddleware", new(middleware.AuthImpl))
}

func RegisterHandlers() {
	appContainer.RegisterService("healthCheckHandler", new(handler.HealthCheckHandlerImpl))
	appContainer.RegisterService("authHandler", new(handler.AuthHandlerImpl))
	appContainer.RegisterService("userHandler", new(handler.UserHandlerImpl))
	appContainer.RegisterService("router", new(router.Router))
}
