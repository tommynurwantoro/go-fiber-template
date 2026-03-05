package repository

import (
	"app/internal/domain"
	"context"
)

//go:generate mockgen -source=user_repository.go -destination=../../adapter/database/repository/mocks/user_repository.go -package=mocks
type UserRepository interface {
	GetAll(ctx context.Context, limit, offset int, search string) ([]domain.User, int64, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) (*domain.User, error)
	UpdatePassOrVerify(ctx context.Context, user *domain.User, id string) error
	Delete(ctx context.Context, id string) error
}
