package service

import (
	"app/config"
	"app/internal/application/model"
	"app/internal/application/service/mocks"
	"app/internal/domain"
	"app/internal/domain/myerrors"
	"app/internal/pkg/crypto"
	mockRepository "app/internal/adapter/database/repository/mocks"
	mockValidator "app/internal/pkg/validator/mocks"
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type tokenServiceTestSuite struct {
	suite.Suite
	mockCtrl       *gomock.Controller
	mockTokenRepo  *mockRepository.MockTokenRepository
	mockUserSvc    *mocks.MockUserService
	mockValidator  *mockValidator.MockValidator
	tokenService   *TokenServiceImpl
	ctx            context.Context
	testUUID       uuid.UUID
	testSecret     string
	hashedPass     string
}

func TestTokenService(t *testing.T) {
	suite.Run(t, new(tokenServiceTestSuite))
}

func (s *tokenServiceTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockTokenRepo = mockRepository.NewMockTokenRepository(s.mockCtrl)
	s.mockUserSvc = mocks.NewMockUserService(s.mockCtrl)
	s.mockValidator = mockValidator.NewMockValidator(s.mockCtrl)

	s.testSecret = "test-secret-key-for-unit-testing"
	s.tokenService = &TokenServiceImpl{
		Conf: &config.Config{
			JWT: config.JWTConfig{
				Secret:              s.testSecret,
				Expire:              30 * time.Minute,
				RefreshExpire:       7,
				ResetPasswordExpire: 15 * time.Minute,
				VerifyEmailExpire:   24 * time.Hour,
			},
		},
		TokenRepository: s.mockTokenRepo,
		UserService:     s.mockUserSvc,
		Validator:       s.mockValidator,
	}

	s.ctx = context.Background()
	s.testUUID = uuid.Must(uuid.NewV7())

	var err error
	s.hashedPass, err = crypto.HashPassword("password123")
	s.Require().NoError(err)
}

func (s *tokenServiceTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

// Helper to create test token
func (s *tokenServiceTestSuite) createTestToken(tokenType domain.TokenType) *domain.Token {
	return &domain.Token{
		ID:        uuid.Must(uuid.NewV7()),
		Token:     "test-token-string",
		UserID:    s.testUUID,
		Type:      tokenType,
		Expires:   time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Helper to create a valid JWT token for testing
func (s *tokenServiceTestSuite) createTestJWTToken(userID string, tokenType domain.TokenType) string {
	claims := jwt.MapClaims{
		"user_id":    userID,
		"token_type": tokenType.String(),
		"exp":        time.Now().Add(1 * time.Hour).Unix(),
		"iat":        time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(s.testSecret))
	return tokenString
}

// ==================== DeleteToken Tests ====================

func (s *tokenServiceTestSuite) TestDeleteToken_Success() {
	userID := s.testUUID.String()

	s.mockTokenRepo.EXPECT().
		Delete(s.ctx, domain.TokenTypeRefresh, userID).
		Return(nil)

	err := s.tokenService.DeleteToken(s.ctx, domain.TokenTypeRefresh, userID)

	s.NoError(err)
}

func (s *tokenServiceTestSuite) TestDeleteToken_RepositoryError() {
	userID := s.testUUID.String()

	s.mockTokenRepo.EXPECT().
		Delete(s.ctx, domain.TokenTypeRefresh, userID).
		Return(myerrors.ErrDeleteTokenFailed)

	err := s.tokenService.DeleteToken(s.ctx, domain.TokenTypeRefresh, userID)

	s.Error(err)
	s.Equal(myerrors.ErrDeleteTokenFailed, err)
}

// ==================== DeleteAllToken Tests ====================

func (s *tokenServiceTestSuite) TestDeleteAllToken_Success() {
	userID := s.testUUID.String()

	s.mockTokenRepo.EXPECT().
		DeleteAll(s.ctx, userID).
		Return(nil)

	err := s.tokenService.DeleteAllToken(s.ctx, userID)

	s.NoError(err)
}

func (s *tokenServiceTestSuite) TestDeleteAllToken_RepositoryError() {
	userID := s.testUUID.String()

	s.mockTokenRepo.EXPECT().
		DeleteAll(s.ctx, userID).
		Return(myerrors.ErrDeleteAllTokenFailed)

	err := s.tokenService.DeleteAllToken(s.ctx, userID)

	s.Error(err)
	s.Equal(myerrors.ErrDeleteAllTokenFailed, err)
}

// ==================== GetTokenByRefreshToken Tests ====================

func (s *tokenServiceTestSuite) TestGetTokenByRefreshToken_Success() {
	userID := s.testUUID.String()
	refreshToken := s.createTestJWTToken(userID, domain.TokenTypeRefresh)
	testToken := s.createTestToken(domain.TokenTypeRefresh)
	testToken.Token = refreshToken

	s.mockTokenRepo.EXPECT().
		GetByTokenAndUserID(s.ctx, refreshToken, userID).
		Return(testToken, nil)

	result, err := s.tokenService.GetTokenByRefreshToken(s.ctx, refreshToken)

	s.NoError(err)
	s.Equal(testToken, result)
}

func (s *tokenServiceTestSuite) TestGetTokenByRefreshToken_InvalidToken() {
	invalidToken := "invalid-token-string"

	result, err := s.tokenService.GetTokenByRefreshToken(s.ctx, invalidToken)

	s.Error(err)
	s.Equal(myerrors.ErrInvalidToken, err)
	s.Nil(result)
}

func (s *tokenServiceTestSuite) TestGetTokenByRefreshToken_WrongTokenType() {
	userID := s.testUUID.String()
	// Create token with wrong type (access instead of refresh)
	wrongTypeToken := s.createTestJWTToken(userID, domain.TokenTypeAccess)

	result, err := s.tokenService.GetTokenByRefreshToken(s.ctx, wrongTypeToken)

	s.Error(err)
	s.Equal(myerrors.ErrInvalidTokenType, err)
	s.Nil(result)
}

func (s *tokenServiceTestSuite) TestGetTokenByRefreshToken_TokenNotFound() {
	userID := s.testUUID.String()
	refreshToken := s.createTestJWTToken(userID, domain.TokenTypeRefresh)

	s.mockTokenRepo.EXPECT().
		GetByTokenAndUserID(s.ctx, refreshToken, userID).
		Return(nil, myerrors.ErrTokenNotFound)

	result, err := s.tokenService.GetTokenByRefreshToken(s.ctx, refreshToken)

	s.Error(err)
	s.Equal(myerrors.ErrTokenNotFound, err)
	s.Nil(result)
}

func (s *tokenServiceTestSuite) TestGetTokenByRefreshToken_RepositoryError() {
	userID := s.testUUID.String()
	refreshToken := s.createTestJWTToken(userID, domain.TokenTypeRefresh)

	s.mockTokenRepo.EXPECT().
		GetByTokenAndUserID(s.ctx, refreshToken, userID).
		Return(nil, myerrors.ErrGetTokenByUserIDFailed)

	result, err := s.tokenService.GetTokenByRefreshToken(s.ctx, refreshToken)

	s.Error(err)
	s.Equal(myerrors.ErrGetTokenByUserIDFailed, err)
	s.Nil(result)
}

// ==================== GenerateAuthTokens Tests ====================

func (s *tokenServiceTestSuite) TestGenerateAuthTokens_Success() {
	userID := s.testUUID.String()

	s.mockTokenRepo.EXPECT().
		Delete(s.ctx, domain.TokenTypeRefresh, userID).
		Return(nil)

	s.mockTokenRepo.EXPECT().
		Create(s.ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, token *domain.Token) (*domain.Token, error) {
			token.ID = s.testUUID
			return token, nil
		})

	accessToken, refreshToken, err := s.tokenService.GenerateAuthTokens(s.ctx, userID)

	s.NoError(err)
	s.NotNil(accessToken)
	s.NotNil(refreshToken)
	s.Equal(domain.TokenTypeAccess, accessToken.Type)
	s.Equal(domain.TokenTypeRefresh, refreshToken.Type)
	s.Equal(uuid.MustParse(userID), accessToken.UserID)
	s.Equal(uuid.MustParse(userID), refreshToken.UserID)
}

func (s *tokenServiceTestSuite) TestGenerateAuthTokens_DeleteOldTokenError() {
	userID := s.testUUID.String()

	s.mockTokenRepo.EXPECT().
		Delete(s.ctx, domain.TokenTypeRefresh, userID).
		Return(myerrors.ErrDeleteTokenFailed)

	accessToken, refreshToken, err := s.tokenService.GenerateAuthTokens(s.ctx, userID)

	s.Error(err)
	s.Equal(myerrors.ErrDeleteTokenFailed, err)
	s.Nil(accessToken)
	s.Nil(refreshToken)
}

func (s *tokenServiceTestSuite) TestGenerateAuthTokens_CreateTokenError() {
	userID := s.testUUID.String()

	s.mockTokenRepo.EXPECT().
		Delete(s.ctx, domain.TokenTypeRefresh, userID).
		Return(nil)

	s.mockTokenRepo.EXPECT().
		Create(s.ctx, gomock.Any()).
		Return(nil, myerrors.ErrSaveTokenFailed)

	accessToken, refreshToken, err := s.tokenService.GenerateAuthTokens(s.ctx, userID)

	s.Error(err)
	s.Equal(myerrors.ErrSaveTokenFailed, err)
	s.Nil(accessToken)
	s.Nil(refreshToken)
}

// ==================== GenerateAccessToken Tests ====================

func (s *tokenServiceTestSuite) TestGenerateAccessToken_Success() {
	userID := s.testUUID.String()

	accessToken, err := s.tokenService.GenerateAccessToken(s.ctx, userID)

	s.NoError(err)
	s.NotNil(accessToken)
	s.Equal(domain.TokenTypeAccess, accessToken.Type)
	s.Equal(uuid.MustParse(userID), accessToken.UserID)
	s.NotEmpty(accessToken.Token)
}

// ==================== GenerateResetPasswordToken Tests ====================

func (s *tokenServiceTestSuite) TestGenerateResetPasswordToken_Success() {
	req := &model.ForgotPasswordRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	testUser := &domain.User{
		ID:            s.testUUID,
		Name:          "Test User",
		Email:         "test@example.com",
		Password:      s.hashedPass,
		Role:          "user",
		VerifiedEmail: true,
	}

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserSvc.EXPECT().
		GetUserByEmail(s.ctx, req.Email).
		Return(testUser, nil)

	s.mockTokenRepo.EXPECT().
		Delete(s.ctx, domain.TokenTypeResetPassword, testUser.ID.String()).
		Return(nil)

	s.mockTokenRepo.EXPECT().
		Create(s.ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, token *domain.Token) (*domain.Token, error) {
			token.ID = s.testUUID
			return token, nil
		})

	result, err := s.tokenService.GenerateResetPasswordToken(s.ctx, req)

	s.NoError(err)
	s.NotNil(result)
	s.Equal(domain.TokenTypeResetPassword, result.Type)
	s.Equal(testUser.ID, result.UserID)
}

func (s *tokenServiceTestSuite) TestGenerateResetPasswordToken_ValidationError() {
	req := &model.ForgotPasswordRequest{
		Email:    "invalid-email",
		Password: "",
	}

	validationErr := myerrors.ErrInvalidRequest

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(validationErr)

	result, err := s.tokenService.GenerateResetPasswordToken(s.ctx, req)

	s.Error(err)
	s.Equal(myerrors.ErrInvalidRequest, err)
	s.Nil(result)
}

func (s *tokenServiceTestSuite) TestGenerateResetPasswordToken_UserNotFound() {
	req := &model.ForgotPasswordRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserSvc.EXPECT().
		GetUserByEmail(s.ctx, req.Email).
		Return(nil, myerrors.ErrUserNotFound)

	result, err := s.tokenService.GenerateResetPasswordToken(s.ctx, req)

	s.Error(err)
	s.Equal(myerrors.ErrUserNotFound, err)
	s.Nil(result)
}

func (s *tokenServiceTestSuite) TestGenerateResetPasswordToken_InvalidPassword() {
	req := &model.ForgotPasswordRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	testUser := &domain.User{
		ID:            s.testUUID,
		Name:          "Test User",
		Email:         "test@example.com",
		Password:      s.hashedPass,
		Role:          "user",
		VerifiedEmail: true,
	}

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserSvc.EXPECT().
		GetUserByEmail(s.ctx, req.Email).
		Return(testUser, nil)

	result, err := s.tokenService.GenerateResetPasswordToken(s.ctx, req)

	s.Error(err)
	s.Equal(myerrors.ErrInvalidPassword, err)
	s.Nil(result)
}

func (s *tokenServiceTestSuite) TestGenerateResetPasswordToken_DeleteTokenError() {
	req := &model.ForgotPasswordRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	testUser := &domain.User{
		ID:            s.testUUID,
		Name:          "Test User",
		Email:         "test@example.com",
		Password:      s.hashedPass,
		Role:          "user",
		VerifiedEmail: true,
	}

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserSvc.EXPECT().
		GetUserByEmail(s.ctx, req.Email).
		Return(testUser, nil)

	s.mockTokenRepo.EXPECT().
		Delete(s.ctx, domain.TokenTypeResetPassword, testUser.ID.String()).
		Return(myerrors.ErrDeleteTokenFailed)

	result, err := s.tokenService.GenerateResetPasswordToken(s.ctx, req)

	s.Error(err)
	s.Equal(myerrors.ErrDeleteTokenFailed, err)
	s.Nil(result)
}

func (s *tokenServiceTestSuite) TestGenerateResetPasswordToken_CreateTokenError() {
	req := &model.ForgotPasswordRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	testUser := &domain.User{
		ID:            s.testUUID,
		Name:          "Test User",
		Email:         "test@example.com",
		Password:      s.hashedPass,
		Role:          "user",
		VerifiedEmail: true,
	}

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserSvc.EXPECT().
		GetUserByEmail(s.ctx, req.Email).
		Return(testUser, nil)

	s.mockTokenRepo.EXPECT().
		Delete(s.ctx, domain.TokenTypeResetPassword, testUser.ID.String()).
		Return(nil)

	s.mockTokenRepo.EXPECT().
		Create(s.ctx, gomock.Any()).
		Return(nil, myerrors.ErrSaveTokenFailed)

	result, err := s.tokenService.GenerateResetPasswordToken(s.ctx, req)

	s.Error(err)
	s.Equal(myerrors.ErrSaveTokenFailed, err)
	s.Nil(result)
}

// ==================== GenerateVerifyEmailToken Tests ====================

func (s *tokenServiceTestSuite) TestGenerateVerifyEmailToken_Success() {
	userID := s.testUUID.String()

	s.mockTokenRepo.EXPECT().
		Delete(s.ctx, domain.TokenTypeVerifyEmail, userID).
		Return(nil)

	s.mockTokenRepo.EXPECT().
		Create(s.ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, token *domain.Token) (*domain.Token, error) {
			token.ID = s.testUUID
			return token, nil
		})

	result, err := s.tokenService.GenerateVerifyEmailToken(s.ctx, userID)

	s.NoError(err)
	s.NotNil(result)
	s.Equal(domain.TokenTypeVerifyEmail, result.Type)
	s.Equal(uuid.MustParse(userID), result.UserID)
}

func (s *tokenServiceTestSuite) TestGenerateVerifyEmailToken_DeleteTokenError() {
	userID := s.testUUID.String()

	s.mockTokenRepo.EXPECT().
		Delete(s.ctx, domain.TokenTypeVerifyEmail, userID).
		Return(myerrors.ErrDeleteTokenFailed)

	result, err := s.tokenService.GenerateVerifyEmailToken(s.ctx, userID)

	s.Error(err)
	s.Equal(myerrors.ErrDeleteTokenFailed, err)
	s.Nil(result)
}

func (s *tokenServiceTestSuite) TestGenerateVerifyEmailToken_CreateTokenError() {
	userID := s.testUUID.String()

	s.mockTokenRepo.EXPECT().
		Delete(s.ctx, domain.TokenTypeVerifyEmail, userID).
		Return(nil)

	s.mockTokenRepo.EXPECT().
		Create(s.ctx, gomock.Any()).
		Return(nil, myerrors.ErrSaveTokenFailed)

	result, err := s.tokenService.GenerateVerifyEmailToken(s.ctx, userID)

	s.Error(err)
	s.Equal(myerrors.ErrSaveTokenFailed, err)
	s.Nil(result)
}

// ==================== generateToken Tests (via public methods) ====================

func (s *tokenServiceTestSuite) TestGenerateToken_TokenContainsCorrectClaims() {
	userID := s.testUUID.String()

	accessToken, err := s.tokenService.GenerateAccessToken(s.ctx, userID)

	s.NoError(err)

	// Verify the token can be parsed and contains correct claims
	token, err := jwt.Parse(accessToken.Token, func(_ *jwt.Token) (any, error) {
		return []byte(s.testSecret), nil
	})
	s.NoError(err)
	s.True(token.Valid)

	claims, ok := token.Claims.(jwt.MapClaims)
	s.True(ok)
	s.Equal(userID, claims["user_id"])
	s.Equal(domain.TokenTypeAccess.String(), claims["token_type"])
}

// ==================== saveToken Tests (via public methods) ====================

func (s *tokenServiceTestSuite) TestSaveToken_DeletesOldTokenFirst() {
	userID := s.testUUID.String()

	// Verify delete is called before create
	callOrder := []string{}

	s.mockTokenRepo.EXPECT().
		Delete(s.ctx, domain.TokenTypeRefresh, userID).
		DoAndReturn(func(_ context.Context, _ domain.TokenType, _ string) error {
			callOrder = append(callOrder, "delete")
			return nil
		})

	s.mockTokenRepo.EXPECT().
		Create(s.ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, token *domain.Token) (*domain.Token, error) {
			callOrder = append(callOrder, "create")
			token.ID = s.testUUID
			return token, nil
		})

	_, refreshToken, err := s.tokenService.GenerateAuthTokens(s.ctx, userID)

	s.NoError(err)
	s.NotNil(refreshToken)
	s.Equal([]string{"delete", "create"}, callOrder)
}
