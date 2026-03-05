package service

import (
	"app/config"
	"app/internal/application/model"
	"app/internal/application/service/mocks"
	"app/internal/domain"
	"app/internal/domain/myerrors"
	"app/internal/pkg/crypto"
	mockValidator "app/internal/pkg/validator/mocks"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type authServiceTestSuite struct {
	suite.Suite
	mockCtrl      *gomock.Controller
	mockTokenSvc  *mocks.MockTokenService
	mockUserSvc   *mocks.MockUserService
	mockValidator *mockValidator.MockValidator
	authService   *AuthServiceImpl
	ctx           context.Context
	testUUID      uuid.UUID
	hashedPass    string
}

func TestAuthService(t *testing.T) {
	suite.Run(t, new(authServiceTestSuite))
}

func (s *authServiceTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockTokenSvc = mocks.NewMockTokenService(s.mockCtrl)
	s.mockUserSvc = mocks.NewMockUserService(s.mockCtrl)
	s.mockValidator = mockValidator.NewMockValidator(s.mockCtrl)

	s.authService = &AuthServiceImpl{
		Conf: &config.Config{
			JWT: config.JWTConfig{
				Secret: "test-secret-key-for-unit-testing",
			},
		},
		TokenService: s.mockTokenSvc,
		UserService:  s.mockUserSvc,
		Validate:     s.mockValidator,
	}

	s.ctx = context.Background()
	s.testUUID = uuid.Must(uuid.NewV7())

	var err error
	s.hashedPass, err = crypto.HashPassword("password123")
	s.Require().NoError(err)
}

func (s *authServiceTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

// Helper to create test user
func (s *authServiceTestSuite) createTestUser() *domain.User {
	return &domain.User{
		ID:            s.testUUID,
		Name:          "Test User",
		Email:         "test@example.com",
		Password:      s.hashedPass,
		Role:          "user",
		VerifiedEmail: false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// Helper to create test token
func (s *authServiceTestSuite) createTestToken(tokenType domain.TokenType) *domain.Token {
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

// ==================== Register Tests ====================

func (s *authServiceTestSuite) TestRegister_Success() {
	req := &model.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	expectedUser := s.createTestUser()

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserSvc.EXPECT().
		CreateUser(s.ctx, &model.CreateUserRequest{
			Name:     req.Name,
			Email:    req.Email,
			Password: req.Password,
			Role:     "user",
		}).
		Return(expectedUser, nil)

	result, err := s.authService.Register(s.ctx, req)

	s.NoError(err)
	s.Equal(expectedUser, result)
	s.Equal("Test User", result.Name)
	s.Equal("test@example.com", result.Email)
}

func (s *authServiceTestSuite) TestRegister_ValidationError() {
	req := &model.RegisterRequest{
		Name:     "",
		Email:    "invalid-email",
		Password: "short",
	}

	validationErr := errors.New("validation failed")

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(validationErr)

	result, err := s.authService.Register(s.ctx, req)

	s.Error(err)
	s.Equal(validationErr, err)
	s.Nil(result)
}

func (s *authServiceTestSuite) TestRegister_CreateUserError() {
	req := &model.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	createErr := myerrors.ErrEmailAlreadyInUse

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserSvc.EXPECT().
		CreateUser(s.ctx, gomock.Any()).
		Return(nil, createErr)

	result, err := s.authService.Register(s.ctx, req)

	s.Error(err)
	s.Equal(createErr, err)
	s.Nil(result)
}

// ==================== Login Tests ====================

func (s *authServiceTestSuite) TestLogin_Success() {
	req := &model.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	testUser := s.createTestUser()

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserSvc.EXPECT().
		GetUserByEmail(s.ctx, req.Email).
		Return(testUser, nil)

	result, err := s.authService.Login(s.ctx, req)

	s.NoError(err)
	s.Equal(testUser, result)
	s.Equal("test@example.com", result.Email)
}

func (s *authServiceTestSuite) TestLogin_ValidationError() {
	req := &model.LoginRequest{
		Email:    "",
		Password: "",
	}

	validationErr := errors.New("validation failed")

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(validationErr)

	result, err := s.authService.Login(s.ctx, req)

	s.Error(err)
	s.Equal(validationErr, err)
	s.Nil(result)
}

func (s *authServiceTestSuite) TestLogin_UserNotFound() {
	req := &model.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserSvc.EXPECT().
		GetUserByEmail(s.ctx, req.Email).
		Return(nil, myerrors.ErrUserNotFound)

	result, err := s.authService.Login(s.ctx, req)

	s.Error(err)
	s.Equal(myerrors.ErrUserNotFound, err)
	s.Nil(result)
}

func (s *authServiceTestSuite) TestLogin_InvalidPassword() {
	req := &model.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	testUser := s.createTestUser()

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserSvc.EXPECT().
		GetUserByEmail(s.ctx, req.Email).
		Return(testUser, nil)

	result, err := s.authService.Login(s.ctx, req)

	s.Error(err)
	s.Equal(myerrors.ErrInvalidEmailOrPassword, err)
	s.Nil(result)
}

// ==================== Logout Tests ====================

func (s *authServiceTestSuite) TestLogout_Success() {
	req := &model.LogoutRequest{
		RefreshToken: "valid-refresh-token",
	}

	testToken := s.createTestToken(domain.TokenTypeRefresh)

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockTokenSvc.EXPECT().
		GetTokenByRefreshToken(s.ctx, req.RefreshToken).
		Return(testToken, nil)

	s.mockTokenSvc.EXPECT().
		DeleteToken(s.ctx, domain.TokenTypeRefresh, testToken.UserID.String()).
		Return(nil)

	err := s.authService.Logout(s.ctx, req)

	s.NoError(err)
}

func (s *authServiceTestSuite) TestLogout_ValidationError() {
	req := &model.LogoutRequest{
		RefreshToken: "",
	}

	validationErr := errors.New("validation failed")

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(validationErr)

	err := s.authService.Logout(s.ctx, req)

	s.Error(err)
	s.Equal(validationErr, err)
}

func (s *authServiceTestSuite) TestLogout_TokenNotFound() {
	req := &model.LogoutRequest{
		RefreshToken: "invalid-refresh-token",
	}

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockTokenSvc.EXPECT().
		GetTokenByRefreshToken(s.ctx, req.RefreshToken).
		Return(nil, myerrors.ErrTokenNotFound)

	err := s.authService.Logout(s.ctx, req)

	s.Error(err)
	s.Equal(myerrors.ErrTokenNotFound, err)
}

func (s *authServiceTestSuite) TestLogout_DeleteTokenError() {
	req := &model.LogoutRequest{
		RefreshToken: "valid-refresh-token",
	}

	testToken := s.createTestToken(domain.TokenTypeRefresh)
	deleteErr := myerrors.ErrDeleteTokenFailed

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockTokenSvc.EXPECT().
		GetTokenByRefreshToken(s.ctx, req.RefreshToken).
		Return(testToken, nil)

	s.mockTokenSvc.EXPECT().
		DeleteToken(s.ctx, domain.TokenTypeRefresh, testToken.UserID.String()).
		Return(deleteErr)

	err := s.authService.Logout(s.ctx, req)

	s.Error(err)
	s.Equal(deleteErr, err)
}

// ==================== RefreshAuth Tests ====================

func (s *authServiceTestSuite) TestRefreshAuth_Success() {
	req := &model.RefreshTokenRequest{
		RefreshToken: "valid-refresh-token",
	}

	testToken := s.createTestToken(domain.TokenTypeRefresh)
	testUser := s.createTestUser()
	newAccessToken := s.createTestToken(domain.TokenTypeAccess)

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockTokenSvc.EXPECT().
		GetTokenByRefreshToken(s.ctx, req.RefreshToken).
		Return(testToken, nil)

	s.mockUserSvc.EXPECT().
		GetUserByID(s.ctx, testToken.UserID.String()).
		Return(testUser, nil)

	s.mockTokenSvc.EXPECT().
		GenerateAccessToken(s.ctx, testUser.ID.String()).
		Return(newAccessToken, nil)

	result, err := s.authService.RefreshAuth(s.ctx, req)

	s.NoError(err)
	s.Equal(newAccessToken, result)
}

func (s *authServiceTestSuite) TestRefreshAuth_ValidationError() {
	req := &model.RefreshTokenRequest{
		RefreshToken: "",
	}

	validationErr := errors.New("validation failed")

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(validationErr)

	result, err := s.authService.RefreshAuth(s.ctx, req)

	s.Error(err)
	s.Equal(validationErr, err)
	s.Nil(result)
}

func (s *authServiceTestSuite) TestRefreshAuth_TokenNotFound() {
	req := &model.RefreshTokenRequest{
		RefreshToken: "invalid-refresh-token",
	}

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockTokenSvc.EXPECT().
		GetTokenByRefreshToken(s.ctx, req.RefreshToken).
		Return(nil, myerrors.ErrTokenNotFound)

	result, err := s.authService.RefreshAuth(s.ctx, req)

	s.Error(err)
	s.Equal(myerrors.ErrTokenNotFound, err)
	s.Nil(result)
}

func (s *authServiceTestSuite) TestRefreshAuth_UserNotFound() {
	req := &model.RefreshTokenRequest{
		RefreshToken: "valid-refresh-token",
	}

	testToken := s.createTestToken(domain.TokenTypeRefresh)

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockTokenSvc.EXPECT().
		GetTokenByRefreshToken(s.ctx, req.RefreshToken).
		Return(testToken, nil)

	s.mockUserSvc.EXPECT().
		GetUserByID(s.ctx, testToken.UserID.String()).
		Return(nil, myerrors.ErrUserNotFound)

	result, err := s.authService.RefreshAuth(s.ctx, req)

	s.Error(err)
	s.Equal(myerrors.ErrUserNotFound, err)
	s.Nil(result)
}

func (s *authServiceTestSuite) TestRefreshAuth_GenerateAccessTokenError() {
	req := &model.RefreshTokenRequest{
		RefreshToken: "valid-refresh-token",
	}

	testToken := s.createTestToken(domain.TokenTypeRefresh)
	testUser := s.createTestUser()

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockTokenSvc.EXPECT().
		GetTokenByRefreshToken(s.ctx, req.RefreshToken).
		Return(testToken, nil)

	s.mockUserSvc.EXPECT().
		GetUserByID(s.ctx, testToken.UserID.String()).
		Return(testUser, nil)

	s.mockTokenSvc.EXPECT().
		GenerateAccessToken(s.ctx, testUser.ID.String()).
		Return(nil, myerrors.ErrGenerateTokenFailed)

	result, err := s.authService.RefreshAuth(s.ctx, req)

	s.Error(err)
	s.Equal(myerrors.ErrGenerateTokenFailed, err)
	s.Nil(result)
}

// ==================== ResetPassword Tests ====================

func (s *authServiceTestSuite) TestResetPassword_Success() {
	// Create a valid JWT token for reset password
	userID := s.testUUID.String()
	validToken := createTestJWTToken(userID, "resetPassword", s.authService.Conf.JWT.Secret)

	req := &model.ResetPasswordRequest{
		Token:    validToken,
		Password: "newpassword123",
	}

	testUser := s.createTestUser()

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserSvc.EXPECT().
		GetUserByID(s.ctx, userID).
		Return(testUser, nil)

	s.mockUserSvc.EXPECT().
		UpdatePassOrVerify(s.ctx, &model.UpdatePassOrVerifyRequest{
			Password:      req.Password,
			VerifiedEmail: testUser.VerifiedEmail,
		}, testUser.ID.String()).
		Return(nil)

	s.mockTokenSvc.EXPECT().
		DeleteToken(s.ctx, domain.TokenTypeResetPassword, testUser.ID.String()).
		Return(nil)

	err := s.authService.ResetPassword(s.ctx, req)

	s.NoError(err)
}

func (s *authServiceTestSuite) TestResetPassword_ValidationError() {
	req := &model.ResetPasswordRequest{
		Token:    "",
		Password: "",
	}

	validationErr := errors.New("validation failed")

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(validationErr)

	err := s.authService.ResetPassword(s.ctx, req)

	s.Error(err)
	s.Equal(validationErr, err)
}

func (s *authServiceTestSuite) TestResetPassword_InvalidToken() {
	req := &model.ResetPasswordRequest{
		Token:    "invalid-token",
		Password: "newpassword123",
	}

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	err := s.authService.ResetPassword(s.ctx, req)

	s.Error(err)
	s.Equal(myerrors.ErrInvalidToken, err)
}

func (s *authServiceTestSuite) TestResetPassword_UserNotFound() {
	userID := s.testUUID.String()
	validToken := createTestJWTToken(userID, "resetPassword", s.authService.Conf.JWT.Secret)

	req := &model.ResetPasswordRequest{
		Token:    validToken,
		Password: "newpassword123",
	}

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserSvc.EXPECT().
		GetUserByID(s.ctx, userID).
		Return(nil, myerrors.ErrUserNotFound)

	err := s.authService.ResetPassword(s.ctx, req)

	s.Error(err)
	s.Equal(myerrors.ErrUserNotFound, err)
}

func (s *authServiceTestSuite) TestResetPassword_UpdatePassError() {
	userID := s.testUUID.String()
	validToken := createTestJWTToken(userID, "resetPassword", s.authService.Conf.JWT.Secret)

	req := &model.ResetPasswordRequest{
		Token:    validToken,
		Password: "newpassword123",
	}

	testUser := s.createTestUser()

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserSvc.EXPECT().
		GetUserByID(s.ctx, userID).
		Return(testUser, nil)

	s.mockUserSvc.EXPECT().
		UpdatePassOrVerify(s.ctx, gomock.Any(), testUser.ID.String()).
		Return(myerrors.ErrUpdatePassOrVerifyFailed)

	err := s.authService.ResetPassword(s.ctx, req)

	s.Error(err)
	s.Equal(myerrors.ErrUpdatePassOrVerifyFailed, err)
}

func (s *authServiceTestSuite) TestResetPassword_DeleteTokenError() {
	userID := s.testUUID.String()
	validToken := createTestJWTToken(userID, "resetPassword", s.authService.Conf.JWT.Secret)

	req := &model.ResetPasswordRequest{
		Token:    validToken,
		Password: "newpassword123",
	}

	testUser := s.createTestUser()

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserSvc.EXPECT().
		GetUserByID(s.ctx, userID).
		Return(testUser, nil)

	s.mockUserSvc.EXPECT().
		UpdatePassOrVerify(s.ctx, gomock.Any(), testUser.ID.String()).
		Return(nil)

	s.mockTokenSvc.EXPECT().
		DeleteToken(s.ctx, domain.TokenTypeResetPassword, testUser.ID.String()).
		Return(myerrors.ErrDeleteTokenFailed)

	err := s.authService.ResetPassword(s.ctx, req)

	s.Error(err)
	s.Equal(myerrors.ErrDeleteTokenFailed, err)
}

// ==================== VerifyEmail Tests ====================

func (s *authServiceTestSuite) TestVerifyEmail_Success() {
	userID := s.testUUID.String()
	validToken := createTestJWTToken(userID, "verifyEmail", s.authService.Conf.JWT.Secret)

	req := &model.VerifyEmailRequest{
		Token: validToken,
	}

	testUser := s.createTestUser()

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserSvc.EXPECT().
		GetUserByID(s.ctx, userID).
		Return(testUser, nil)

	s.mockTokenSvc.EXPECT().
		DeleteToken(s.ctx, domain.TokenTypeVerifyEmail, testUser.ID.String()).
		Return(nil)

	err := s.authService.VerifyEmail(s.ctx, req)

	s.NoError(err)
}

func (s *authServiceTestSuite) TestVerifyEmail_ValidationError() {
	req := &model.VerifyEmailRequest{
		Token: "",
	}

	validationErr := errors.New("validation failed")

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(validationErr)

	err := s.authService.VerifyEmail(s.ctx, req)

	s.Error(err)
	s.Equal(validationErr, err)
}

func (s *authServiceTestSuite) TestVerifyEmail_InvalidToken() {
	req := &model.VerifyEmailRequest{
		Token: "invalid-token",
	}

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	err := s.authService.VerifyEmail(s.ctx, req)

	s.Error(err)
	s.Equal(myerrors.ErrInvalidToken, err)
}

func (s *authServiceTestSuite) TestVerifyEmail_UserNotFound() {
	userID := s.testUUID.String()
	validToken := createTestJWTToken(userID, "verifyEmail", s.authService.Conf.JWT.Secret)

	req := &model.VerifyEmailRequest{
		Token: validToken,
	}

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserSvc.EXPECT().
		GetUserByID(s.ctx, userID).
		Return(nil, myerrors.ErrUserNotFound)

	err := s.authService.VerifyEmail(s.ctx, req)

	s.Error(err)
	s.Equal(myerrors.ErrUserNotFound, err)
}

func (s *authServiceTestSuite) TestVerifyEmail_DeleteTokenError() {
	userID := s.testUUID.String()
	validToken := createTestJWTToken(userID, "verifyEmail", s.authService.Conf.JWT.Secret)

	req := &model.VerifyEmailRequest{
		Token: validToken,
	}

	testUser := s.createTestUser()

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserSvc.EXPECT().
		GetUserByID(s.ctx, userID).
		Return(testUser, nil)

	s.mockTokenSvc.EXPECT().
		DeleteToken(s.ctx, domain.TokenTypeVerifyEmail, testUser.ID.String()).
		Return(myerrors.ErrDeleteTokenFailed)

	err := s.authService.VerifyEmail(s.ctx, req)

	s.Error(err)
	s.Equal(myerrors.ErrDeleteTokenFailed, err)
}

func (s *authServiceTestSuite) TestVerifyEmail_WrongTokenType() {
	userID := s.testUUID.String()
	// Create token with wrong type
	wrongTypeToken := createTestJWTToken(userID, "access", s.authService.Conf.JWT.Secret)

	req := &model.VerifyEmailRequest{
		Token: wrongTypeToken,
	}

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	err := s.authService.VerifyEmail(s.ctx, req)

	s.Error(err)
	s.Equal(myerrors.ErrInvalidTokenType, err)
}

// ==================== Helper Functions ====================

// createTestJWTToken creates a valid JWT token for testing
func createTestJWTToken(userID, tokenType, secret string) string {
	claims := jwt.MapClaims{
		"user_id":    userID,
		"token_type": tokenType,
		"exp":        time.Now().Add(1 * time.Hour).Unix(),
		"iat":        time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}
