package repository

import (
	"app/internal/adapter/database"
	"app/internal/domain"
	"app/internal/domain/myerrors"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/tommynurwantoro/golog"
	"gorm.io/gorm"
)

type UserRepositoryImpl struct {
	DB database.DatabaseAdapter `inject:"database"`
}

func (r *UserRepositoryImpl) GetAll(ctx context.Context, limit, offset int, search string) ([]domain.User, int64, error) {
	var users []domain.User
	var totalResults int64

	query := r.DB.GetDB().WithContext(ctx).Order("created_at asc")

	if search != "" {
		query = query.Where("name LIKE ? OR email LIKE ? OR role LIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	resultCount := query.Find(&users).Count(&totalResults)
	if resultCount.Error != nil {
		golog.Error("Error counting users", resultCount.Error)
		return nil, 0, myerrors.ErrGetUserFailed
	}

	result := query.Limit(limit).Offset(offset).Find(&users)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, 0, myerrors.ErrUserNotFound
		}
		golog.Error("Error getting users", result.Error)
		return nil, 0, myerrors.ErrGetUserFailed
	}

	return users, totalResults, nil
}

func (r *UserRepositoryImpl) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User

	result := r.DB.GetDB().WithContext(ctx).First(&user, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, myerrors.ErrUserNotFound
		}
		golog.Error("Error getting user by id", result.Error)
		return nil, myerrors.ErrGetUserFailed
	}

	return &user, nil
}

func (r *UserRepositoryImpl) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User

	result := r.DB.GetDB().WithContext(ctx).First(&user, "email = ?", email)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, myerrors.ErrUserNotFound
		}
		golog.Error("Error getting user by email", result.Error)
		return nil, myerrors.ErrGetUserFailed
	}
	return &user, nil
}

func (r *UserRepositoryImpl) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	user.ID = uuid.Must(uuid.NewV7())
	result := r.DB.GetDB().WithContext(ctx).Create(user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return nil, myerrors.ErrEmailAlreadyInUse
		}

		golog.Error("Error creating user", result.Error)
		return nil, myerrors.ErrCreateUserFailed
	}
	return user, nil
}

func (r *UserRepositoryImpl) Update(ctx context.Context, user *domain.User) (*domain.User, error) {
	result := r.DB.GetDB().WithContext(ctx).Where("id = ?", user.ID).Updates(user)

	if result.RowsAffected == 0 {
		return nil, myerrors.ErrUserNotFound
	}

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return nil, myerrors.ErrEmailAlreadyInUse
		}

		golog.Error("Error updating user", result.Error)
		return nil, myerrors.ErrUpdateUserFailed
	}
	return user, nil
}

func (r *UserRepositoryImpl) UpdatePassOrVerify(ctx context.Context, user *domain.User, id string) error {
	result := r.DB.GetDB().WithContext(ctx).Where("id = ?", id).Updates(user)

	if result.RowsAffected == 0 {
		return myerrors.ErrUserNotFound
	}

	if result.Error != nil {
		golog.Error("Error updating user password or verified email", result.Error)
		return myerrors.ErrUpdatePassOrVerifyFailed
	}

	return nil
}

func (r *UserRepositoryImpl) Delete(ctx context.Context, id string) error {
	result := r.DB.GetDB().WithContext(ctx).Delete(&domain.User{}, "id = ?", id)

	if result.RowsAffected == 0 {
		return myerrors.ErrUserNotFound
	}

	if result.Error != nil {
		golog.Error("Error deleting user", result.Error)
		return myerrors.ErrDeleteUserFailed
	}
	return nil
}
