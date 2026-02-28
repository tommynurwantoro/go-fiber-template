package token

import (
	"app/internal/domain/myerrors"

	"github.com/golang-jwt/jwt/v5"
)

func VerifyToken(tokenStr, secret, tokenType string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(_ *jwt.Token) (any, error) {
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return "", myerrors.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", myerrors.ErrInvalidTokenClaims
	}

	jwtType, ok := claims["token_type"].(string)
	if !ok || jwtType != tokenType {
		return "", myerrors.ErrInvalidTokenType
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", myerrors.ErrInvalidTokenUserID
	}

	return userID, nil
}
