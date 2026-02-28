package router

import (
	"app/config"

	_ "app/docs" // swagger init
	"app/internal/adapter/rest"
	"app/internal/application/handler"
	"app/internal/pkg/middleware"

	"github.com/gofiber/swagger"
)

type Router struct {
	App                *rest.Fiber                `inject:"rest"`
	Conf               *config.Config             `inject:"config"`
	HealthCheckHandler handler.HealthCheckHandler `inject:"healthCheckHandler"`
	AuthHandler        handler.AuthHandler        `inject:"authHandler"`
	UserHandler        handler.UserHandler        `inject:"userHandler"`
	AuthMiddleware     middleware.Auth            `inject:"authMiddleware"`
}

func (r *Router) Startup() error {
	healthCheck := r.App.Group("/health-check")
	healthCheck.Get("/", r.HealthCheckHandler.Check)

	auth := r.App.Group("/auth")
	auth.Post("/register", r.AuthHandler.Register)
	auth.Post("/login", r.AuthHandler.Login)
	auth.Post("/logout", r.AuthHandler.Logout)
	auth.Post("/refresh-tokens", r.AuthHandler.RefreshTokens)
	auth.Post("/forgot-password", r.AuthHandler.ForgotPassword)
	auth.Post("/reset-password", r.AuthHandler.ResetPassword)
	auth.Post("/send-verification-email", r.AuthMiddleware.JWTAuth(), r.AuthHandler.SendVerificationEmail)
	auth.Post("/verify-email", r.AuthHandler.VerifyEmail)
	auth.Get("/google", r.AuthHandler.GoogleLogin)
	auth.Get("/google-callback", r.AuthHandler.GoogleCallback)

	v1 := r.App.Group("/v1")

	if r.Conf.Environment == "development" {
		docs := v1.Group("/docs")
		docs.Get("/*", swagger.HandlerDefault)
	}

	user := v1.Group("/users")
	user.Get("/", r.AuthMiddleware.JWTAuth("getUsers"), r.UserHandler.GetUsers)
	user.Post("/", r.AuthMiddleware.JWTAuth("manageUsers"), r.UserHandler.CreateUser)
	user.Get("/:userId", r.AuthMiddleware.JWTAuth("getUsers"), r.UserHandler.GetUserByID)
	user.Patch("/:userId", r.AuthMiddleware.JWTAuth("manageUsers"), r.UserHandler.UpdateUser)
	user.Delete("/:userId", r.AuthMiddleware.JWTAuth("manageUsers"), r.UserHandler.DeleteUser)

	return nil
}

func (r *Router) Shutdown() error {
	return nil
}
