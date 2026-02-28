package middleware

import (
	"app/config"
	"app/internal/application/service"
	"app/internal/domain"
	"app/internal/domain/myerrors"
	"app/internal/pkg/token"
	"errors"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/tommynurwantoro/golog"
)

type Auth interface {
	JWTAuth(requiredRights ...string) fiber.Handler
}

type AuthImpl struct {
	Conf        *config.Config      `inject:"config"`
	UserService service.UserService `inject:"userService"`
}

func (a *AuthImpl) JWTAuth(requiredRights ...string) fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(a.Conf.JWT.Secret)},
		ErrorHandler: func(_ *fiber.Ctx, err error) error {
			golog.Error("Error verifying token", err)
			return myerrors.ErrInvalidToken
		},
		SuccessHandler: func(c *fiber.Ctx) error {
			user, ok := c.Locals("user").(*jwt.Token)
			if !ok {
				return myerrors.ErrInvalidToken
			}
			userID, err := token.VerifyToken(user.Raw, a.Conf.JWT.Secret, domain.TokenTypeAccess.String())
			if err != nil {
				return myerrors.ErrInvalidToken
			}

			_user, err := a.UserService.GetUserByID(c, userID)
			if err != nil && !errors.Is(err, myerrors.ErrUserNotFound) {
				golog.Error("Error getting user by id", err)
				return myerrors.ErrGetUserFailed
			}

			c.Locals("user", _user)

			if len(requiredRights) > 0 {
				userRights, hasRight := config.RoleRights[_user.Role]
				if (!hasRight || !hasAllRights(userRights, requiredRights)) && c.Params("userId") != userID {
					return fiber.NewError(fiber.StatusForbidden, "you don't have permission to access this resource")
				}
			}

			return c.Next()
		},
	})
}

func hasAllRights(userRights, requiredRights []string) bool {
	rightSet := make(map[string]struct{}, len(userRights))
	for _, right := range userRights {
		rightSet[right] = struct{}{}
	}

	for _, right := range requiredRights {
		if _, exists := rightSet[right]; !exists {
			return false
		}
	}
	return true
}
