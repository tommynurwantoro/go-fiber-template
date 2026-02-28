package rest

import (
	"app/internal/domain/myerrors"
	"app/internal/pkg/formatter"

	"github.com/gofiber/fiber/v2"
)

var CodeMap = map[error]formatter.Status{
	// Fiber errors
	myerrors.ErrInvalidRequest: formatter.InvalidRequest,

	// Token errors
	myerrors.ErrInvalidToken:       formatter.Unauthorized,
	myerrors.ErrInvalidTokenClaims: formatter.Unauthorized,
	myerrors.ErrInvalidTokenType:   formatter.Unauthorized,
	myerrors.ErrInvalidTokenUserID: formatter.Unauthorized,
	myerrors.ErrTokenNotFound:      formatter.DataNotFound,

	// User errors
	myerrors.ErrUserNotFound:           formatter.DataNotFound,
	myerrors.ErrEmailAlreadyInUse:      formatter.DataConflict,
	myerrors.ErrInvalidEmailOrPassword: formatter.Unauthorized,
	myerrors.ErrInvalidPassword:        formatter.Unauthorized,
}

var StatusMap = map[error]int{
	// Fiber errors
	myerrors.ErrInvalidRequest: fiber.StatusBadRequest,

	// Token errors
	myerrors.ErrInvalidToken:       fiber.StatusUnauthorized,
	myerrors.ErrInvalidTokenClaims: fiber.StatusUnauthorized,
	myerrors.ErrInvalidTokenType:   fiber.StatusUnauthorized,
	myerrors.ErrInvalidTokenUserID: fiber.StatusUnauthorized,
	myerrors.ErrTokenNotFound:      fiber.StatusNotFound,

	// User errors
	myerrors.ErrUserNotFound:           fiber.StatusNotFound,
	myerrors.ErrEmailAlreadyInUse:      fiber.StatusConflict,
	myerrors.ErrInvalidEmailOrPassword: fiber.StatusUnauthorized,
	myerrors.ErrInvalidPassword:        fiber.StatusUnauthorized,
}
