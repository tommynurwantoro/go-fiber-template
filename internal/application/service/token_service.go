package service

import (
	"app/config"
	"app/internal/application/model"
	"app/internal/domain"
	"app/internal/domain/myerrors"
	"app/internal/domain/repository"
	"app/internal/pkg/crypto"
	"app/internal/pkg/token"
	"app/internal/pkg/validator"
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tommynurwantoro/golog"
)

//go:generate mockgen -source=token_service.go -destination=mocks/token_service.go -package=mocks
type TokenService interface {
	DeleteToken(ctx context.Context, tokenType domain.TokenType, userID string) error
	DeleteAllToken(ctx context.Context, userID string) error
	GetTokenByRefreshToken(ctx context.Context, refreshToken string) (*domain.Token, error)
	GenerateAuthTokens(ctx context.Context, userID string) (*domain.Token, *domain.Token, error)
	GenerateAccessToken(ctx context.Context, userID string) (*domain.Token, error)
	GenerateResetPasswordToken(ctx context.Context, req *model.ForgotPasswordRequest) (*domain.Token, error)
	GenerateVerifyEmailToken(ctx context.Context, userID string) (*domain.Token, error)
}

type TokenServiceImpl struct {
	Conf            *config.Config             `inject:"config"`
	TokenRepository repository.TokenRepository `inject:"tokenRepository"`
	UserService     UserService                `inject:"userService"`
	Validator       validator.Validator        `inject:"validator"`
}

func (s *TokenServiceImpl) DeleteToken(ctx context.Context, tokenType domain.TokenType, userID string) error {
	return s.TokenRepository.Delete(ctx, tokenType, userID)
}

func (s *TokenServiceImpl) DeleteAllToken(ctx context.Context, userID string) error {
	return s.TokenRepository.DeleteAll(ctx, userID)
}

func (s *TokenServiceImpl) GetTokenByRefreshToken(ctx context.Context, refreshToken string) (*domain.Token, error) {
	userID, err := token.VerifyToken(refreshToken, s.Conf.JWT.Secret, domain.TokenTypeRefresh.String())
	if err != nil {
		return nil, err
	}

	tokenDoc, err := s.TokenRepository.GetByTokenAndUserID(ctx, refreshToken, userID)
	if err != nil {
		return nil, err
	}

	return tokenDoc, nil
}

func (s *TokenServiceImpl) GenerateAuthTokens(ctx context.Context, userID string) (*domain.Token, *domain.Token, error) {
	accessTokenExpires := time.Now().UTC().Add(s.Conf.JWT.Expire)
	accessToken, err := s.generateToken(userID, accessTokenExpires, domain.TokenTypeAccess)
	if err != nil {
		golog.Error("Error generating access token", err)
		return nil, nil, myerrors.ErrGenerateTokenFailed
	}

	accessTokenDomain := &domain.Token{
		Token:   accessToken,
		UserID:  uuid.MustParse(userID),
		Type:    domain.TokenTypeAccess,
		Expires: accessTokenExpires,
	}

	refreshTokenExpires := time.Now().UTC().Add(time.Hour * 24 * time.Duration(s.Conf.JWT.RefreshExpire))
	refreshToken, err := s.generateToken(userID, refreshTokenExpires, domain.TokenTypeRefresh)
	if err != nil {
		golog.Error("Error generating refresh token", err)
		return nil, nil, myerrors.ErrGenerateTokenFailed
	}

	refreshTokenDomain, err := s.saveToken(ctx, refreshToken, userID, domain.TokenTypeRefresh, refreshTokenExpires)
	if err != nil {
		return nil, nil, err
	}

	return accessTokenDomain, refreshTokenDomain, nil
}

func (s *TokenServiceImpl) GenerateAccessToken(_ context.Context, userID string) (*domain.Token, error) {
	accessTokenExpires := time.Now().UTC().Add(s.Conf.JWT.Expire)
	accessToken, err := s.generateToken(userID, accessTokenExpires, domain.TokenTypeAccess)
	if err != nil {
		golog.Error("Error generating access token", err)
		return nil, myerrors.ErrGenerateTokenFailed
	}

	accessTokenDomain := &domain.Token{
		Token:   accessToken,
		UserID:  uuid.MustParse(userID),
		Type:    domain.TokenTypeAccess,
		Expires: accessTokenExpires,
	}

	return accessTokenDomain, nil
}

func (s *TokenServiceImpl) GenerateResetPasswordToken(
	ctx context.Context, req *model.ForgotPasswordRequest,
) (*domain.Token, error) {
	if err := s.Validator.Validate(ctx, req); err != nil {
		golog.Error("Error validating reset password request", err)
		return nil, myerrors.ErrInvalidRequest
	}

	user, err := s.UserService.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if !crypto.CheckPasswordHash(req.Password, user.Password) {
		return nil, myerrors.ErrInvalidPassword
	}

	expires := time.Now().UTC().Add(s.Conf.JWT.ResetPasswordExpire)
	resetPasswordToken, err := s.generateToken(user.ID.String(), expires, domain.TokenTypeResetPassword)
	if err != nil {
		golog.Error("Error signing reset password token", err)
		return nil, myerrors.ErrGenerateTokenFailed
	}

	resetPasswordTokenDomain, err := s.saveToken(
		ctx, resetPasswordToken, user.ID.String(), domain.TokenTypeResetPassword, expires,
	)
	if err != nil {
		return nil, err
	}

	return resetPasswordTokenDomain, nil
}

func (s *TokenServiceImpl) GenerateVerifyEmailToken(ctx context.Context, userID string) (*domain.Token, error) {
	expires := time.Now().UTC().Add(s.Conf.JWT.VerifyEmailExpire)
	verifyEmailToken, err := s.generateToken(userID, expires, domain.TokenTypeVerifyEmail)
	if err != nil {
		golog.Error("Error generating verify email token", err)
		return nil, myerrors.ErrGenerateTokenFailed
	}

	verifyEmailTokenDomain, err := s.saveToken(ctx, verifyEmailToken, userID, domain.TokenTypeVerifyEmail, expires)
	if err != nil {
		return nil, err
	}

	return verifyEmailTokenDomain, nil
}

func (s *TokenServiceImpl) generateToken(userID string, expires time.Time, tokenType domain.TokenType) (string, error) {
	claims := jwt.MapClaims{
		"user_id":    userID,
		"issued_at":  time.Now().Unix(),
		"expires_at": expires.Unix(),
		"token_type": tokenType.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(s.Conf.JWT.Secret))
}

func (s *TokenServiceImpl) saveToken(
	ctx context.Context, token, userID string, tokenType domain.TokenType, expires time.Time,
) (*domain.Token, error) {
	if err := s.TokenRepository.Delete(ctx, tokenType, userID); err != nil {
		golog.Error("Error deleting token", err)
		return nil, myerrors.ErrDeleteTokenFailed
	}

	tokenDoc := &domain.Token{
		Token:   token,
		UserID:  uuid.MustParse(userID),
		Type:    tokenType,
		Expires: expires,
	}

	savedToken, err := s.TokenRepository.Create(ctx, tokenDoc)
	if err != nil {
		return nil, err
	}

	return savedToken, nil
}
