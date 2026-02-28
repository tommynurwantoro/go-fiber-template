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
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tommynurwantoro/golog"
	"gorm.io/gorm"
)

type TokenService interface {
	DeleteToken(c *fiber.Ctx, tokenType domain.TokenType, userID string) error
	DeleteAllToken(c *fiber.Ctx, userID string) error
	GetTokenByRefreshToken(c *fiber.Ctx, refreshToken string) (*domain.Token, error)
	GenerateAuthTokens(c *fiber.Ctx, userID string) (*domain.Token, *domain.Token, error)
	GenerateAccessToken(c *fiber.Ctx, userID string) (*domain.Token, error)
	GenerateResetPasswordToken(c *fiber.Ctx, req *model.ForgotPasswordRequest) (*domain.Token, error)
	GenerateVerifyEmailToken(c *fiber.Ctx, userID string) (*domain.Token, error)
}

type TokenServiceImpl struct {
	Conf        *config.Config           `inject:"config"`
	DB          database.DatabaseAdapter `inject:"database"`
	UserService UserService              `inject:"userService"`
	Validator   validator.Validator      `inject:"validator"`
}

func (s *TokenServiceImpl) DeleteToken(c *fiber.Ctx, tokenType domain.TokenType, userID string) error {
	tokenDoc := new(domain.Token)

	result := s.DB.GetDB().WithContext(c.Context()).
		Where("type = ? AND user_id = ?", tokenType.String(), userID).
		Delete(tokenDoc)

	if result.Error != nil {
		golog.Error("Error deleting token", result.Error)
		return myerrors.ErrDeleteTokenFailed
	}

	return nil
}

func (s *TokenServiceImpl) DeleteAllToken(c *fiber.Ctx, userID string) error {
	tokenDoc := new(domain.Token)

	result := s.DB.GetDB().WithContext(c.Context()).Where("user_id = ?", userID).Delete(tokenDoc)

	if result.Error != nil {
		golog.Error("Error deleting all token", result.Error)
		return myerrors.ErrDeleteAllTokenFailed
	}

	return nil
}

func (s *TokenServiceImpl) GetTokenByRefreshToken(c *fiber.Ctx, refreshToken string) (*domain.Token, error) {
	userID, err := token.VerifyToken(refreshToken, s.Conf.JWT.Secret, domain.TokenTypeRefresh.String())
	if err != nil {
		return nil, err
	}

	tokenDoc := new(domain.Token)

	result := s.DB.GetDB().WithContext(c.Context()).
		Where("token = ? AND user_id = ?", refreshToken, userID).
		First(tokenDoc)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, myerrors.ErrTokenNotFound
		}
		golog.Error("Error getting token by refresh token", result.Error)
		return nil, myerrors.ErrGetTokenByUserIDFailed
	}

	return tokenDoc, nil
}

func (s *TokenServiceImpl) GenerateAuthTokens(c *fiber.Ctx, userID string) (*domain.Token, *domain.Token, error) {
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

	refreshTokenDomain, err := s.saveToken(c, refreshToken, userID, domain.TokenTypeRefresh, refreshTokenExpires)
	if err != nil {
		return nil, nil, err
	}

	return accessTokenDomain, refreshTokenDomain, nil
}

func (s *TokenServiceImpl) GenerateAccessToken(_ *fiber.Ctx, userID string) (*domain.Token, error) {
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
	c *fiber.Ctx, req *model.ForgotPasswordRequest,
) (*domain.Token, error) {
	if err := s.Validator.Validate(c.Context(), req); err != nil {
		golog.Error("Error validating reset password request", err)
		return nil, myerrors.ErrInvalidRequest
	}

	user, err := s.UserService.GetUserByEmail(c, req.Email)
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
		c, resetPasswordToken, user.ID.String(), domain.TokenTypeResetPassword, expires,
	)
	if err != nil {
		return nil, err
	}

	return resetPasswordTokenDomain, nil
}

func (s *TokenServiceImpl) GenerateVerifyEmailToken(c *fiber.Ctx, userID string) (*domain.Token, error) {
	expires := time.Now().UTC().Add(s.Conf.JWT.VerifyEmailExpire)
	verifyEmailToken, err := s.generateToken(userID, expires, domain.TokenTypeVerifyEmail)
	if err != nil {
		golog.Error("Error generating verify email token", err)
		return nil, myerrors.ErrGenerateTokenFailed
	}

	verifyEmailTokenDomain, err := s.saveToken(c, verifyEmailToken, userID, domain.TokenTypeVerifyEmail, expires)
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
	c *fiber.Ctx, token, userID string, tokenType domain.TokenType, expires time.Time,
) (*domain.Token, error) {
	if err := s.DeleteToken(c, tokenType, userID); err != nil {
		golog.Error("Error deleting token", err)
		return nil, myerrors.ErrDeleteTokenFailed
	}

	tokenDoc := &domain.Token{
		Token:   token,
		UserID:  uuid.MustParse(userID),
		Type:    tokenType,
		Expires: expires,
	}

	result := s.DB.GetDB().WithContext(c.Context()).Create(tokenDoc)

	if result.Error != nil {
		golog.Error("Error saving token", result.Error)
		return nil, myerrors.ErrSaveTokenFailed
	}

	return tokenDoc, nil
}
