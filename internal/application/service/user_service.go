package service

import (
	"app/internal/adapter/database"
	"app/internal/application/model"
	"app/internal/domain"
	"app/internal/domain/myerrors"
	"app/internal/pkg/crypto"
	"app/internal/pkg/validator"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/tommynurwantoro/golog"
	"gorm.io/gorm"
)

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
	DB        database.DatabaseAdapter `inject:"database"`
	Validator validator.Validator      `inject:"validator"`
}

func (u *UserServiceImpl) GetUsers(c *fiber.Ctx, req *model.GetUserRequest) ([]domain.User, int64, error) {
	var users []domain.User
	var totalResults int64

	if err := u.Validator.Validate(c.Context(), req); err != nil {
		golog.Error("Error validating get users request", err)
		return nil, 0, myerrors.ErrInvalidRequest
	}

	offset := (req.Page - 1) * req.Limit
	query := u.DB.GetDB().WithContext(c.Context()).Order("created_at asc")

	if search := req.Search; search != "" {
		query = query.Where("name LIKE ? OR email LIKE ? OR role LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	result := query.Find(&users).Count(&totalResults)
	if result.Error != nil {
		golog.Error("Error counting users", result.Error)
		return nil, 0, myerrors.ErrGetUserFailed
	}

	result = query.Limit(req.Limit).Offset(offset).Find(&users)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, 0, myerrors.ErrUserNotFound
		}
		golog.Error("Error finding users", result.Error)
		return nil, 0, myerrors.ErrGetUserFailed
	}

	return users, totalResults, nil
}

func (u *UserServiceImpl) GetUserByID(c *fiber.Ctx, id string) (*domain.User, error) {
	user := new(domain.User)

	result := u.DB.GetDB().WithContext(c.Context()).First(user, "id = ?", id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, myerrors.ErrUserNotFound
	}

	if result.Error != nil {
		golog.Error("Error getting user by id", result.Error)
		return nil, myerrors.ErrGetUserFailed
	}

	return user, nil
}

func (u *UserServiceImpl) GetUserByEmail(c *fiber.Ctx, email string) (*domain.User, error) {
	user := new(domain.User)

	result := u.DB.GetDB().WithContext(c.Context()).Where("email = ?", email).First(user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, myerrors.ErrUserNotFound
	}

	if result.Error != nil {
		golog.Error("Error getting user by email", result.Error)
		return nil, myerrors.ErrGetUserFailed
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

	result := u.DB.GetDB().WithContext(c.Context()).Create(user)

	if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
		return nil, myerrors.ErrEmailAlreadyInUse
	}

	if result.Error != nil {
		golog.Error("Error creating user", result.Error)
		return nil, myerrors.ErrCreateUserFailed
	}

	return user, nil
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

	result := u.DB.GetDB().WithContext(c.Context()).Where("id = ?", req.UserID).Updates(updateBody)

	if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
		return nil, myerrors.ErrEmailAlreadyInUse
	}

	if result.RowsAffected == 0 {
		return nil, myerrors.ErrUserNotFound
	}

	if result.Error != nil {
		golog.Error("Error updating user", result.Error)
		return nil, myerrors.ErrUpdateUserFailed
	}

	user, err := u.GetUserByID(c, req.UserID)
	if err != nil {
		golog.Error("Error getting user by id", err)
		return nil, myerrors.ErrGetUserFailed
	}

	return user, nil
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

	result := u.DB.GetDB().WithContext(c.Context()).Where("id = ?", id).Updates(updateBody)

	if result.RowsAffected == 0 {
		return myerrors.ErrUserNotFound
	}

	if result.Error != nil {
		golog.Error("Error updating user password or verifiedEmail", result.Error)
		return myerrors.ErrUpdatePassOrVerifyFailed
	}

	return nil
}

func (u *UserServiceImpl) DeleteUser(c *fiber.Ctx, id string) error {
	user := new(domain.User)

	result := u.DB.GetDB().WithContext(c.Context()).Delete(user, "id = ?", id)

	if result.RowsAffected == 0 {
		return myerrors.ErrUserNotFound
	}

	if result.Error != nil {
		golog.Error("Error deleting user", result.Error)
		return myerrors.ErrDeleteUserFailed
	}

	return result.Error
}

func (u *UserServiceImpl) CreateGoogleUser(c *fiber.Ctx, req *model.CreateGoogleUserRequest) (*domain.User, error) {
	if err := u.Validator.Validate(c.Context(), req); err != nil {
		golog.Error("Error validating create google user request", err)
		return nil, myerrors.ErrInvalidRequest
	}

	userFromDB, err := u.GetUserByEmail(c, req.Email)
	if err != nil {
		if errors.Is(err, myerrors.ErrUserNotFound) {
			user := &domain.User{
				Name:          req.Name,
				Email:         req.Email,
				VerifiedEmail: req.VerifiedEmail,
			}

			result := u.DB.GetDB().WithContext(c.Context()).Create(user)
			if result.Error != nil {
				golog.Error("Error creating user", result.Error)
				return nil, myerrors.ErrCreateUserFailed
			}

			return user, nil
		}

		golog.Error("Error getting user by email", err)
		return nil, myerrors.ErrGetUserFailed
	}

	userFromDB.VerifiedEmail = req.VerifiedEmail
	result := u.DB.GetDB().WithContext(c.Context()).Save(userFromDB)
	if result.Error != nil {
		golog.Error("Error updating user", result.Error)
		return nil, myerrors.ErrUpdateUserFailed
	}

	return userFromDB, nil
}
