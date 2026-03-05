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
	"context"
)

type AuthService interface {
	Register(ctx context.Context, req *model.RegisterRequest) (*domain.User, error)
	Login(ctx context.Context, req *model.LoginRequest) (*domain.User, error)
	Logout(ctx context.Context, req *model.LogoutRequest) error
	RefreshAuth(ctx context.Context, req *model.RefreshTokenRequest) (*domain.Token, error)
	ResetPassword(ctx context.Context, req *model.ResetPasswordRequest) error
	VerifyEmail(ctx context.Context, query *model.VerifyEmailRequest) error
}

type AuthServiceImpl struct {
	Conf         *config.Config           `inject:"config"`
	DB           database.DatabaseAdapter `inject:"database"`
	TokenService TokenService             `inject:"tokenService"`
	UserService  UserService              `inject:"userService"`
	Validate     validator.Validator      `inject:"validator"`
}

func (s *AuthServiceImpl) Register(ctx context.Context, req *model.RegisterRequest) (*domain.User, error) {
	if err := s.Validate.Validate(ctx, req); err != nil {
		return nil, err
	}

	newUser, err := s.UserService.CreateUser(ctx, &model.CreateUserRequest{
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

func (s *AuthServiceImpl) Login(ctx context.Context, req *model.LoginRequest) (*domain.User, error) {
	if err := s.Validate.Validate(ctx, req); err != nil {
		return nil, err
	}

	user, err := s.UserService.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if !crypto.CheckPasswordHash(req.Password, user.Password) {
		return nil, myerrors.ErrInvalidEmailOrPassword
	}

	return user, nil
}

func (s *AuthServiceImpl) Logout(ctx context.Context, req *model.LogoutRequest) error {
	if err := s.Validate.Validate(ctx, req); err != nil {
		return err
	}

	token, err := s.TokenService.GetTokenByRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return err
	}

	return s.TokenService.DeleteToken(ctx, domain.TokenTypeRefresh, token.UserID.String())
}

func (s *AuthServiceImpl) RefreshAuth(ctx context.Context, req *model.RefreshTokenRequest) (*domain.Token, error) {
	if err := s.Validate.Validate(ctx, req); err != nil {
		return nil, err
	}

	token, err := s.TokenService.GetTokenByRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	user, err := s.UserService.GetUserByID(ctx, token.UserID.String())
	if err != nil {
		return nil, err
	}

	accessToken, err := s.TokenService.GenerateAccessToken(ctx, user.ID.String())
	if err != nil {
		return nil, err
	}

	return accessToken, nil
}

func (s *AuthServiceImpl) ResetPassword(ctx context.Context, req *model.ResetPasswordRequest) error {
	if err := s.Validate.Validate(ctx, req); err != nil {
		return err
	}

	userID, err := token.VerifyToken(req.Token, s.Conf.JWT.Secret, domain.TokenTypeResetPassword.String())
	if err != nil {
		return err
	}

	user, err := s.UserService.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if errUpdate := s.UserService.UpdatePassOrVerify(ctx, &model.UpdatePassOrVerifyRequest{
		Password:      req.Password,
		VerifiedEmail: user.VerifiedEmail,
	}, user.ID.String()); errUpdate != nil {
		return errUpdate
	}

	if errToken := s.TokenService.DeleteToken(ctx, domain.TokenTypeResetPassword, user.ID.String()); errToken != nil {
		return errToken
	}

	return nil
}

func (s *AuthServiceImpl) VerifyEmail(ctx context.Context, req *model.VerifyEmailRequest) error {
	if err := s.Validate.Validate(ctx, req); err != nil {
		return err
	}

	userID, err := token.VerifyToken(req.Token, s.Conf.JWT.Secret, domain.TokenTypeVerifyEmail.String())
	if err != nil {
		return err
	}

	user, err := s.UserService.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if delErr := s.TokenService.DeleteToken(ctx, domain.TokenTypeVerifyEmail, user.ID.String()); delErr != nil {
		return delErr
	}

	return nil
}
