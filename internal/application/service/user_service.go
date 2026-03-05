package service

import (
	"app/internal/adapter/database"
	"app/internal/adapter/database/repository"
	"app/internal/application/model"
	"app/internal/domain"
	"app/internal/domain/myerrors"
	"app/internal/pkg/crypto"
	"app/internal/pkg/validator"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/tommynurwantoro/golog"
)

//go:generate mockgen -source=user_service.go -destination=mocks/user_service.go -package=mocks
type UserService interface {
	GetUsers(c *fiber.Ctx, params *model.GetUserRequest) ([]domain.User, int64, error)
	GetUserByID(c *fiber.Ctx, id string) (*domain.User, error)
	GetUserByEmail(c *fiber.Ctx, email string) (*domain.User, error)
	CreateUser(c *fiber.Ctx, req *model.CreateUserRequest) (*domain.User, error)
	UpdatePassOrVerify(c *fiber.Ctx, req *model.UpdatePassOrVerifyRequest, id string) error
	UpdateUser(c *fiber.Ctx, req *model.UpdateUserRequest) (*domain.User, error)
	DeleteUser(c *fiber.Ctx, id string) error
	CreateGoogleUser(c *fiber.Ctx, req *model.CreateGoogleUserRequest) (*domain.User, error)
}

type UserServiceImpl struct {
	DB             database.DatabaseAdapter  `inject:"database"`
	UserRepository repository.UserRepository `inject:"userRepository"`
	Validator      validator.Validator       `inject:"validator"`
}

func (u *UserServiceImpl) GetUsers(c *fiber.Ctx, req *model.GetUserRequest) ([]domain.User, int64, error) {
	if err := u.Validator.Validate(c.Context(), req); err != nil {
		golog.Error("Error validating get users request", err)
		return nil, 0, myerrors.ErrInvalidRequest
	}

	offset := (req.Page - 1) * req.Limit
	users, totalResults, err := u.UserRepository.GetAll(c.Context(), req.Limit, offset, req.Search)
	if err != nil {
		return nil, 0, err
	}

	return users, totalResults, nil
}

func (u *UserServiceImpl) GetUserByID(c *fiber.Ctx, id string) (*domain.User, error) {
	user, err := u.UserRepository.GetByID(c.Context(), id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UserServiceImpl) GetUserByEmail(c *fiber.Ctx, email string) (*domain.User, error) {
	user, err := u.UserRepository.GetByEmail(c.Context(), email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UserServiceImpl) CreateUser(c *fiber.Ctx, req *model.CreateUserRequest) (*domain.User, error) {
	if err := u.Validator.Validate(c.Context(), req); err != nil {
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

	newUser, err := u.UserRepository.Create(c.Context(), user)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}

func (u *UserServiceImpl) UpdateUser(c *fiber.Ctx, req *model.UpdateUserRequest) (*domain.User, error) {
	if err := u.Validator.Validate(c.Context(), req); err != nil {
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

	updatedUser, err := u.UserRepository.Update(c.Context(), updateBody)
	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (u *UserServiceImpl) UpdatePassOrVerify(c *fiber.Ctx, req *model.UpdatePassOrVerifyRequest, id string) error {
	if err := u.Validator.Validate(c.Context(), req); err != nil {
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

	err := u.UserRepository.UpdatePassOrVerify(c.Context(), updateBody, id)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserServiceImpl) DeleteUser(c *fiber.Ctx, id string) error {
	err := u.UserRepository.Delete(c.Context(), id)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserServiceImpl) CreateGoogleUser(c *fiber.Ctx, req *model.CreateGoogleUserRequest) (*domain.User, error) {
	if err := u.Validator.Validate(c.Context(), req); err != nil {
		golog.Error("Error validating create google user request", err)
		return nil, myerrors.ErrInvalidRequest
	}

	userFromDB, err := u.UserRepository.GetByEmail(c.Context(), req.Email)
	if err != nil {
		if errors.Is(err, myerrors.ErrUserNotFound) {
			newUser, err := u.UserRepository.Create(c.Context(), &domain.User{
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
	updatedUser, err := u.UserRepository.Update(c.Context(), userFromDB)
	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}
