package myerrors

import "errors"

var (
	ErrDeleteTokenFailed      = errors.New("failed to delete token")
	ErrDeleteAllTokenFailed   = errors.New("failed to delete all token")
	ErrSaveTokenFailed        = errors.New("failed to save token")
	ErrGetTokenByUserIDFailed = errors.New("failed to get token by user id")
	ErrTokenNotFound          = errors.New("token not found")
	ErrGenerateTokenFailed    = errors.New("failed to generate token")
	ErrInvalidToken           = errors.New("invalid token")
	ErrInvalidTokenClaims     = errors.New("invalid token claims")
	ErrInvalidTokenType       = errors.New("invalid token type")
	ErrInvalidTokenUserID     = errors.New("invalid token user id")
)
