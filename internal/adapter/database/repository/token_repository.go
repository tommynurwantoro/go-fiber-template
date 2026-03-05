package repository

import (
	"app/internal/adapter/database"
	"app/internal/domain"
	"app/internal/domain/myerrors"
	"context"
	"errors"

	"github.com/tommynurwantoro/golog"
	"gorm.io/gorm"
)

//go:generate mockgen -source=token_repository.go -destination=mocks/token_repository.go -package=mocks
type TokenRepository interface {
	Create(ctx context.Context, token *domain.Token) (*domain.Token, error)
	Delete(ctx context.Context, tokenType domain.TokenType, userID string) error
	DeleteAll(ctx context.Context, userID string) error
	GetByTokenAndUserID(ctx context.Context, token, userID string) (*domain.Token, error)
}

type TokenRepositoryImpl struct {
	DB database.DatabaseAdapter `inject:"database"`
}

func (r *TokenRepositoryImpl) Create(ctx context.Context, token *domain.Token) (*domain.Token, error) {
	result := r.DB.GetDB().WithContext(ctx).Create(token)
	if result.Error != nil {
		golog.Error("Error creating token", result.Error)
		return nil, myerrors.ErrSaveTokenFailed
	}
	return token, nil
}

func (r *TokenRepositoryImpl) Delete(ctx context.Context, tokenType domain.TokenType, userID string) error {
	result := r.DB.GetDB().WithContext(ctx).
		Delete(&domain.Token{}, "type = ? AND user_id = ?", tokenType.String(), userID)

	if result.Error != nil {
		golog.Error("Error deleting token", result.Error)
		return myerrors.ErrDeleteTokenFailed
	}

	return nil
}

func (r *TokenRepositoryImpl) DeleteAll(ctx context.Context, userID string) error {
	result := r.DB.GetDB().WithContext(ctx).Delete(&domain.Token{}, "user_id = ?", userID)

	if result.Error != nil {
		golog.Error("Error deleting all token", result.Error)
		return myerrors.ErrDeleteAllTokenFailed
	}

	return nil
}

func (r *TokenRepositoryImpl) GetByTokenAndUserID(ctx context.Context, token, userID string) (*domain.Token, error) {
	var tokenDoc domain.Token

	result := r.DB.GetDB().WithContext(ctx).
		First(tokenDoc, "token = ? AND user_id = ?", token, userID)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, myerrors.ErrTokenNotFound
		}
		golog.Error("Error getting token by token and user id", result.Error)
		return nil, myerrors.ErrGetTokenByUserIDFailed
	}

	return &tokenDoc, nil
}
