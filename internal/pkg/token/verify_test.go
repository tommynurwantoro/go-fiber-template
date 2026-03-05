package token

import (
	"app/internal/domain/myerrors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/suite"
)

type verifyTestSuite struct {
	suite.Suite
	testSecret string
}

func TestVerify(t *testing.T) {
	suite.Run(t, new(verifyTestSuite))
}

func (s *verifyTestSuite) SetupTest() {
	s.testSecret = "test-secret-key-for-unit-testing"
}

// Helper to create a valid JWT token for testing
func (s *verifyTestSuite) createTestToken(userID, tokenType, secret string, exp time.Time) string {
	claims := jwt.MapClaims{
		"user_id":    userID,
		"token_type": tokenType,
		"exp":        exp.Unix(),
		"iat":        time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

// ==================== VerifyToken Success Tests ====================

func (s *verifyTestSuite) TestVerifyToken_Success() {
	userID := "123e4567-e89b-12d3-a456-426614174000"
	tokenType := "access"
	token := s.createTestToken(userID, tokenType, s.testSecret, time.Now().Add(1*time.Hour))

	result, err := VerifyToken(token, s.testSecret, tokenType)

	s.NoError(err)
	s.Equal(userID, result)
}

func (s *verifyTestSuite) TestVerifyToken_ResetPasswordType() {
	userID := "123e4567-e89b-12d3-a456-426614174000"
	tokenType := "resetPassword"
	token := s.createTestToken(userID, tokenType, s.testSecret, time.Now().Add(15*time.Minute))

	result, err := VerifyToken(token, s.testSecret, tokenType)

	s.NoError(err)
	s.Equal(userID, result)
}

func (s *verifyTestSuite) TestVerifyToken_VerifyEmailType() {
	userID := "123e4567-e89b-12d3-a456-426614174000"
	tokenType := "verifyEmail"
	token := s.createTestToken(userID, tokenType, s.testSecret, time.Now().Add(24*time.Hour))

	result, err := VerifyToken(token, s.testSecret, tokenType)

	s.NoError(err)
	s.Equal(userID, result)
}

func (s *verifyTestSuite) TestVerifyToken_RefreshType() {
	userID := "123e4567-e89b-12d3-a456-426614174000"
	tokenType := "refresh"
	token := s.createTestToken(userID, tokenType, s.testSecret, time.Now().Add(7*24*time.Hour))

	result, err := VerifyToken(token, s.testSecret, tokenType)

	s.NoError(err)
	s.Equal(userID, result)
}

// ==================== VerifyToken Invalid Token Tests ====================

func (s *verifyTestSuite) TestVerifyToken_InvalidTokenFormat() {
	invalidToken := "not-a-valid-jwt-token"

	result, err := VerifyToken(invalidToken, s.testSecret, "access")

	s.Error(err)
	s.Equal(myerrors.ErrInvalidToken, err)
	s.Empty(result)
}

func (s *verifyTestSuite) TestVerifyToken_EmptyToken() {
	result, err := VerifyToken("", s.testSecret, "access")

	s.Error(err)
	s.Equal(myerrors.ErrInvalidToken, err)
	s.Empty(result)
}

func (s *verifyTestSuite) TestVerifyToken_ExpiredToken() {
	userID := "123e4567-e89b-12d3-a456-426614174000"
	tokenType := "access"
	token := s.createTestToken(userID, tokenType, s.testSecret, time.Now().Add(-1*time.Hour))

	result, err := VerifyToken(token, s.testSecret, tokenType)

	s.Error(err)
	s.Equal(myerrors.ErrInvalidToken, err)
	s.Empty(result)
}

func (s *verifyTestSuite) TestVerifyToken_WrongSecret() {
	userID := "123e4567-e89b-12d3-a456-426614174000"
	tokenType := "access"
	token := s.createTestToken(userID, tokenType, s.testSecret, time.Now().Add(1*time.Hour))

	result, err := VerifyToken(token, "wrong-secret", tokenType)

	s.Error(err)
	s.Equal(myerrors.ErrInvalidToken, err)
	s.Empty(result)
}

// ==================== VerifyToken Wrong Type Tests ====================

func (s *verifyTestSuite) TestVerifyToken_WrongTokenType() {
	userID := "123e4567-e89b-12d3-a456-426614174000"
	token := s.createTestToken(userID, "access", s.testSecret, time.Now().Add(1*time.Hour))

	result, err := VerifyToken(token, s.testSecret, "refresh")

	s.Error(err)
	s.Equal(myerrors.ErrInvalidTokenType, err)
	s.Empty(result)
}

func (s *verifyTestSuite) TestVerifyToken_EmptyTokenType() {
	userID := "123e4567-e89b-12d3-a456-426614174000"
	token := s.createTestToken(userID, "access", s.testSecret, time.Now().Add(1*time.Hour))

	result, err := VerifyToken(token, s.testSecret, "")

	s.Error(err)
	s.Equal(myerrors.ErrInvalidTokenType, err)
	s.Empty(result)
}

// ==================== VerifyToken Malformed Claims Tests ====================

func (s *verifyTestSuite) TestVerifyToken_MissingUserID() {
	claims := jwt.MapClaims{
		"token_type": "access",
		"exp":        time.Now().Add(1 * time.Hour).Unix(),
		"iat":        time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(s.testSecret))

	result, err := VerifyToken(tokenString, s.testSecret, "access")

	s.Error(err)
	s.Equal(myerrors.ErrInvalidTokenUserID, err)
	s.Empty(result)
}

func (s *verifyTestSuite) TestVerifyToken_MissingTokenType() {
	claims := jwt.MapClaims{
		"user_id": "123e4567-e89b-12d3-a456-426614174000",
		"exp":     time.Now().Add(1 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(s.testSecret))

	result, err := VerifyToken(tokenString, s.testSecret, "access")

	s.Error(err)
	s.Equal(myerrors.ErrInvalidTokenType, err)
	s.Empty(result)
}

func (s *verifyTestSuite) TestVerifyToken_UserIDNotString() {
	claims := jwt.MapClaims{
		"user_id":    12345, // integer instead of string
		"token_type": "access",
		"exp":        time.Now().Add(1 * time.Hour).Unix(),
		"iat":        time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(s.testSecret))

	result, err := VerifyToken(tokenString, s.testSecret, "access")

	s.Error(err)
	s.Equal(myerrors.ErrInvalidTokenUserID, err)
	s.Empty(result)
}

func (s *verifyTestSuite) TestVerifyToken_TokenTypeNotString() {
	claims := jwt.MapClaims{
		"user_id":    "123e4567-e89b-12d3-a456-426614174000",
		"token_type": 123, // integer instead of string
		"exp":        time.Now().Add(1 * time.Hour).Unix(),
		"iat":        time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(s.testSecret))

	result, err := VerifyToken(tokenString, s.testSecret, "access")

	s.Error(err)
	s.Equal(myerrors.ErrInvalidTokenType, err)
	s.Empty(result)
}

// ==================== VerifyToken Edge Cases ====================

func (s *verifyTestSuite) TestVerifyToken_TokenWithExtraClaims() {
	userID := "123e4567-e89b-12d3-a456-426614174000"
	tokenType := "access"
	claims := jwt.MapClaims{
		"user_id":    userID,
		"token_type": tokenType,
		"exp":        time.Now().Add(1 * time.Hour).Unix(),
		"iat":        time.Now().Unix(),
		"role":       "admin", // extra claim
		"email":      "test@example.com",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(s.testSecret))

	result, err := VerifyToken(tokenString, s.testSecret, tokenType)

	s.NoError(err)
	s.Equal(userID, result)
}

func (s *verifyTestSuite) TestVerifyToken_TokenJustExpired() {
	userID := "123e4567-e89b-12d3-a456-426614174000"
	tokenType := "access"
	// Token expired 1 second ago
	token := s.createTestToken(userID, tokenType, s.testSecret, time.Now().Add(-1*time.Second))

	result, err := VerifyToken(token, s.testSecret, tokenType)

	s.Error(err)
	s.Equal(myerrors.ErrInvalidToken, err)
	s.Empty(result)
}

func (s *verifyTestSuite) TestVerifyToken_TokenAboutToExpire() {
	userID := "123e4567-e89b-12d3-a456-426614174000"
	tokenType := "access"
	// Token expires in 1 second
	token := s.createTestToken(userID, tokenType, s.testSecret, time.Now().Add(1*time.Second))

	result, err := VerifyToken(token, s.testSecret, tokenType)

	s.NoError(err)
	s.Equal(userID, result)
}

func (s *verifyTestSuite) TestVerifyToken_DifferentSigningMethod() {
	// Create token with HS512 instead of HS256
	// Note: HS384/HS512 still work with the same secret, they just use different hash sizes
	// The VerifyToken function doesn't validate the specific algorithm, so this will succeed
	claims := jwt.MapClaims{
		"user_id":    "123e4567-e89b-12d3-a456-426614174000",
		"token_type": "access",
		"exp":        time.Now().Add(1 * time.Hour).Unix(),
		"iat":        time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenString, _ := token.SignedString([]byte(s.testSecret))

	// HS512 tokens are still valid with the same secret
	result, err := VerifyToken(tokenString, s.testSecret, "access")

	s.NoError(err)
	s.Equal("123e4567-e89b-12d3-a456-426614174000", result)
}
