package service

import (
	"app/internal/application/model"
	"app/internal/domain"
	"app/internal/domain/myerrors"
	"app/internal/domain/repository"
	"app/internal/pkg/crypto"
	"app/internal/pkg/validator"
	"context"
	"errors"

	"github.com/tommynurwantoro/golog"
)

//go:generate mockgen -source=user_service.go -destination=mocks/user_service.go -package=mocks
type UserService interface {
	GetUsers(ctx context.Context, params *model.GetUserRequest) ([]domain.User, int64, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	CreateUser(ctx context.Context, req *model.CreateUserRequest) (*domain.User, error)
	UpdatePassOrVerify(ctx context.Context, req *model.UpdatePassOrVerifyRequest, id string) error
	UpdateUser(ctx context.Context, req *model.UpdateUserRequest) (*domain.User, error)
	DeleteUser(ctx context.Context, id string) error
	CreateGoogleUser(ctx context.Context, req *model.CreateGoogleUserRequest) (*domain.User, error)
}

type UserServiceImpl struct {
	UserRepository repository.UserRepository `inject:"userRepository"`
	Validator      validator.Validator       `inject:"validator"`
}

func (u *UserServiceImpl) GetUsers(ctx context.Context, req *model.GetUserRequest) ([]domain.User, int64, error) {
	if err := u.Validator.Validate(ctx, req); err != nil {
		golog.Error("Error validating get users request", err)
		return nil, 0, myerrors.ErrInvalidRequest
	}

	offset := (req.Page - 1) * req.Limit
	users, totalResults, err := u.UserRepository.GetAll(ctx, req.Limit, offset, req.Search)
	if err != nil {
		return nil, 0, err
	}

	return users, totalResults, nil
}

func (u *UserServiceImpl) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	user, err := u.UserRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UserServiceImpl) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := u.UserRepository.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UserServiceImpl) CreateUser(ctx context.Context, req *model.CreateUserRequest) (*domain.User, error) {
	if err := u.Validator.Validate(ctx, req); err != nil {
		golog.Error("Error validating create user request", err)
		return nil, myerrors.ErrInvalidRequest
	}

	hashedPassword, err := crypto.HashPassword(req.Password)
	if err != nil {
		golog.Error("Error hashing password", err)
		return nil, myerrors.ErrHashPassword
	}

	user := &domain.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     req.Role,
	}

	newUser, err := u.UserRepository.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}

func (u *UserServiceImpl) UpdateUser(ctx context.Context, req *model.UpdateUserRequest) (*domain.User, error) {
	if err := u.Validator.Validate(ctx, req); err != nil {
		golog.Error("Error validating update user request", err)
		return nil, myerrors.ErrInvalidRequest
	}

	if req.Email == "" && req.Name == "" && req.Password == "" {
		return nil, myerrors.ErrInvalidRequest
	}

	if req.Password != "" {
		hashedPassword, err := crypto.HashPassword(req.Password)
		if err != nil {
			golog.Error("Error hashing password", err)
			return nil, myerrors.ErrHashPassword
		}
		req.Password = hashedPassword
	}

	updateBody := &domain.User{
		Name:     req.Name,
		Password: req.Password,
		Email:    req.Email,
	}

	updatedUser, err := u.UserRepository.Update(ctx, updateBody)
	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (u *UserServiceImpl) UpdatePassOrVerify(ctx context.Context, req *model.UpdatePassOrVerifyRequest, id string) error {
	if err := u.Validator.Validate(ctx, req); err != nil {
		golog.Error("Error validating update pass or verify request", err)
		return myerrors.ErrInvalidRequest
	}

	if req.Password == "" && !req.VerifiedEmail {
		return myerrors.ErrInvalidRequest
	}

	if req.Password != "" {
		hashedPassword, err := crypto.HashPassword(req.Password)
		if err != nil {
			golog.Error("Error hashing password", err)
			return myerrors.ErrHashPassword
		}
		req.Password = hashedPassword
	}

	updateBody := &domain.User{
		Password:      req.Password,
		VerifiedEmail: req.VerifiedEmail,
	}

	err := u.UserRepository.UpdatePassOrVerify(ctx, updateBody, id)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserServiceImpl) DeleteUser(ctx context.Context, id string) error {
	err := u.UserRepository.Delete(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserServiceImpl) CreateGoogleUser(ctx context.Context, req *model.CreateGoogleUserRequest) (*domain.User, error) {
	if err := u.Validator.Validate(ctx, req); err != nil {
		golog.Error("Error validating create google user request", err)
		return nil, myerrors.ErrInvalidRequest
	}

	userFromDB, err := u.UserRepository.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, myerrors.ErrUserNotFound) {
			newUser, err := u.UserRepository.Create(ctx, &domain.User{
				Name:          req.Name,
				Email:         req.Email,
				VerifiedEmail: req.VerifiedEmail,
			})
			if err != nil {
				return nil, err
			}

			return newUser, nil
		}

		golog.Error("Error getting user by email", err)
		return nil, myerrors.ErrGetUserFailed
	}

	userFromDB.VerifiedEmail = req.VerifiedEmail
	updatedUser, err := u.UserRepository.Update(ctx, userFromDB)
	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}
