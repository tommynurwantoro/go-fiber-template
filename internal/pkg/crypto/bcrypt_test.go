package crypto

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type bcryptTestSuite struct {
	suite.Suite
}

func TestBcrypt(t *testing.T) {
	suite.Run(t, new(bcryptTestSuite))
}

// ==================== HashPassword Tests ====================

func (s *bcryptTestSuite) TestHashPassword_Success() {
	password := "password123"

	hash, err := HashPassword(password)

	s.NoError(err)
	s.NotEmpty(hash)
	s.NotEqual(password, hash)
	s.Len(hash, 60) // bcrypt hashes are always 60 characters
}

func (s *bcryptTestSuite) TestHashPassword_DifferentPasswordsProduceDifferentHashes() {
	password1 := "password123"
	password2 := "password456"

	hash1, err1 := HashPassword(password1)
	hash2, err2 := HashPassword(password2)

	s.NoError(err1)
	s.NoError(err2)
	s.NotEqual(hash1, hash2)
}

func (s *bcryptTestSuite) TestHashPassword_SamePasswordProducesDifferentHashes() {
	password := "password123"

	hash1, err1 := HashPassword(password)
	hash2, err2 := HashPassword(password)

	s.NoError(err1)
	s.NoError(err2)
	s.NotEqual(hash1, hash2) // bcrypt uses salt, so same password = different hash
}

func (s *bcryptTestSuite) TestHashPassword_EmptyPassword() {
	password := ""

	hash, err := HashPassword(password)

	s.NoError(err)
	s.NotEmpty(hash)
}

func (s *bcryptTestSuite) TestHashPassword_LongPassword_Exceeds72Bytes() {
	// bcrypt has a maximum password length of 72 bytes
	password := "this_is_a_very_long_password_that_exceeds_72_bytes_because_bcrypt_only_uses_first_72_bytes_of_password_anyway_so_this_extra_text_should_be_ignored"

	hash, err := HashPassword(password)

	// bcrypt returns an error for passwords exceeding 72 bytes
	s.Error(err)
	s.Empty(hash)
	s.Contains(err.Error(), "password length exceeds 72 bytes")
}

func (s *bcryptTestSuite) TestHashPassword_Exactly72Bytes() {
	// bcrypt accepts passwords up to exactly 72 bytes
	password := "123456789012345678901234567890123456789012345678901234567890123456789012" // exactly 72 chars

	hash, err := HashPassword(password)

	s.NoError(err)
	s.NotEmpty(hash)
	s.Len(hash, 60)
}

// ==================== CheckPasswordHash Tests ====================

func (s *bcryptTestSuite) TestCheckPasswordHash_CorrectPassword() {
	password := "password123"
	hash, err := HashPassword(password)
	s.Require().NoError(err)

	result := CheckPasswordHash(password, hash)

	s.True(result)
}

func (s *bcryptTestSuite) TestCheckPasswordHash_IncorrectPassword() {
	password := "password123"
	hash, err := HashPassword(password)
	s.Require().NoError(err)

	result := CheckPasswordHash("wrongpassword", hash)

	s.False(result)
}

func (s *bcryptTestSuite) TestCheckPasswordHash_EmptyPassword() {
	password := ""
	hash, err := HashPassword(password)
	s.Require().NoError(err)

	result := CheckPasswordHash("", hash)

	s.True(result)
}

func (s *bcryptTestSuite) TestCheckPasswordHash_EmptyPasswordAgainstNonEmptyHash() {
	password := "password123"
	hash, err := HashPassword(password)
	s.Require().NoError(err)

	result := CheckPasswordHash("", hash)

	s.False(result)
}

func (s *bcryptTestSuite) TestCheckPasswordHash_InvalidHash() {
	password := "password123"
	invalidHash := "not-a-valid-bcrypt-hash"

	result := CheckPasswordHash(password, invalidHash)

	s.False(result)
}

func (s *bcryptTestSuite) TestCheckPasswordHash_EmptyHash() {
	password := "password123"

	result := CheckPasswordHash(password, "")

	s.False(result)
}

func (s *bcryptTestSuite) TestCheckPasswordHash_CaseSensitive() {
	password := "Password123"
	hash, err := HashPassword(password)
	s.Require().NoError(err)

	resultLower := CheckPasswordHash("password123", hash)
	resultUpper := CheckPasswordHash("PASSWORD123", hash)
	resultCorrect := CheckPasswordHash("Password123", hash)

	s.False(resultLower)
	s.False(resultUpper)
	s.True(resultCorrect)
}

func (s *bcryptTestSuite) TestCheckPasswordHash_SpecialCharacters() {
	password := "p@$$w0rd!#$%^&*()_+-=[]{}|;':\",./<>?"
	hash, err := HashPassword(password)
	s.Require().NoError(err)

	result := CheckPasswordHash(password, hash)

	s.True(result)
}

func (s *bcryptTestSuite) TestCheckPasswordHash_UnicodeCharacters() {
	password := "密码123パスワード"
	hash, err := HashPassword(password)
	s.Require().NoError(err)

	result := CheckPasswordHash(password, hash)

	s.True(result)
}
