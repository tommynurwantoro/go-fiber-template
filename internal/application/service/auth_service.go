package service

import (
	"app/config"
	"app/internal/adapter/database"
	"app/internal/application/model"
	"app/internal/domain"
	"app/internal/domain/myerrors"
	"app/internal/pkg/crypto"
	"app/internal/pkg/token"
	"app/internal/pkg/validator"

	"github.com/gofiber/fiber/v2"
)

type AuthService interface {
	Register(c *fiber.Ctx, req *model.RegisterRequest) (*domain.User, error)
	Login(c *fiber.Ctx, req *model.LoginRequest) (*domain.User, error)
	Logout(c *fiber.Ctx, req *model.LogoutRequest) error
	RefreshAuth(c *fiber.Ctx, req *model.RefreshTokenRequest) (*domain.Token, error)
	ResetPassword(c *fiber.Ctx, req *model.ResetPasswordRequest) error
	VerifyEmail(c *fiber.Ctx, query *model.VerifyEmailRequest) error
}

type AuthServiceImpl struct {
	Conf         *config.Config           `inject:"config"`
	DB           database.DatabaseAdapter `inject:"database"`
	TokenService TokenService             `inject:"tokenService"`
	UserService  UserService              `inject:"userService"`
	Validate     validator.Validator      `inject:"validator"`
}

func (s *AuthServiceImpl) Register(c *fiber.Ctx, req *model.RegisterRequest) (*domain.User, error) {
	if err := s.Validate.Validate(c.Context(), req); err != nil {
		return nil, err
	}

	newUser, err := s.UserService.CreateUser(c, &model.CreateUserRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     "user",
	})
	if err != nil {
		return nil, err
	}

	return newUser, nil
}

func (s *AuthServiceImpl) Login(c *fiber.Ctx, req *model.LoginRequest) (*domain.User, error) {
	if err := s.Validate.Validate(c.Context(), req); err != nil {
		return nil, err
	}

	user, err := s.UserService.GetUserByEmail(c, req.Email)
	if err != nil {
		return nil, err
	}

	if !crypto.CheckPasswordHash(req.Password, user.Password) {
		return nil, myerrors.ErrInvalidEmailOrPassword
	}

	return user, nil
}

func (s *AuthServiceImpl) Logout(c *fiber.Ctx, req *model.LogoutRequest) error {
	if err := s.Validate.Validate(c.Context(), req); err != nil {
		return err
	}

	token, err := s.TokenService.GetTokenByRefreshToken(c, req.RefreshToken)
	if err != nil {
		return err
	}

	return s.TokenService.DeleteToken(c, domain.TokenTypeRefresh, token.UserID.String())
}

func (s *AuthServiceImpl) RefreshAuth(c *fiber.Ctx, req *model.RefreshTokenRequest) (*domain.Token, error) {
	if err := s.Validate.Validate(c.Context(), req); err != nil {
		return nil, err
	}

	token, err := s.TokenService.GetTokenByRefreshToken(c, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	user, err := s.UserService.GetUserByID(c, token.UserID.String())
	if err != nil {
		return nil, err
	}

	accessToken, err := s.TokenService.GenerateAccessToken(c, user.ID.String())
	if err != nil {
		return nil, err
	}

	return accessToken, nil
}

func (s *AuthServiceImpl) ResetPassword(c *fiber.Ctx, req *model.ResetPasswordRequest) error {
	if err := s.Validate.Validate(c.Context(), req); err != nil {
		return err
	}

	userID, err := token.VerifyToken(req.Token, s.Conf.JWT.Secret, domain.TokenTypeResetPassword.String())
	if err != nil {
		return err
	}

	user, err := s.UserService.GetUserByID(c, userID)
	if err != nil {
		return err
	}

	if errUpdate := s.UserService.UpdatePassOrVerify(c, &model.UpdatePassOrVerifyRequest{
		Password:      req.Password,
		VerifiedEmail: user.VerifiedEmail,
	}, user.ID.String()); errUpdate != nil {
		return errUpdate
	}

	if errToken := s.TokenService.DeleteToken(c, domain.TokenTypeResetPassword, user.ID.String()); errToken != nil {
		return errToken
	}

	return nil
}

func (s *AuthServiceImpl) VerifyEmail(c *fiber.Ctx, req *model.VerifyEmailRequest) error {
	if err := s.Validate.Validate(c.Context(), req); err != nil {
		return err
	}

	userID, err := token.VerifyToken(req.Token, s.Conf.JWT.Secret, domain.TokenTypeVerifyEmail.String())
	if err != nil {
		return err
	}

	user, err := s.UserService.GetUserByID(c, userID)
	if err != nil {
		return err
	}

	if delErr := s.TokenService.DeleteToken(c, domain.TokenTypeVerifyEmail, user.ID.String()); delErr != nil {
		return delErr
	}

	return nil
}
