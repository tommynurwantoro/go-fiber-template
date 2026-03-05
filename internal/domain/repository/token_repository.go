package repository

import (
	"app/internal/domain"
	"context"
)

//go:generate mockgen -source=token_repository.go -destination=../../adapter/database/repository/mocks/token_repository.go -package=mocks
type TokenRepository interface {
	Create(ctx context.Context, token *domain.Token) (*domain.Token, error)
	Delete(ctx context.Context, tokenType domain.TokenType, userID string) error
	DeleteAll(ctx context.Context, userID string) error
	GetByTokenAndUserID(ctx context.Context, token, userID string) (*domain.Token, error)
}
