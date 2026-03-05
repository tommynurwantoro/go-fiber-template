package service

import (
	"app/internal/application/model"
	"app/internal/domain"
	"app/internal/domain/myerrors"
	"app/internal/pkg/crypto"
	mockRepository "app/internal/adapter/database/repository/mocks"
	mockValidator "app/internal/pkg/validator/mocks"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type userServiceTestSuite struct {
	suite.Suite
	mockCtrl        *gomock.Controller
	mockUserRepo    *mockRepository.MockUserRepository
	mockValidator   *mockValidator.MockValidator
	userService     *UserServiceImpl
	ctx             context.Context
	testUUID        uuid.UUID
	testUUID2       uuid.UUID
	hashedPass      string
}

func TestUserService(t *testing.T) {
	suite.Run(t, new(userServiceTestSuite))
}

func (s *userServiceTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockUserRepo = mockRepository.NewMockUserRepository(s.mockCtrl)
	s.mockValidator = mockValidator.NewMockValidator(s.mockCtrl)

	s.userService = &UserServiceImpl{
		UserRepository: s.mockUserRepo,
		Validator:      s.mockValidator,
	}

	s.ctx = context.Background()
	s.testUUID = uuid.Must(uuid.NewV7())
	s.testUUID2 = uuid.Must(uuid.NewV7())

	var err error
	s.hashedPass, err = crypto.HashPassword("password123")
	s.Require().NoError(err)
}

func (s *userServiceTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

// Helper to create test user
func (s *userServiceTestSuite) createTestUser() *domain.User {
	return &domain.User{
		ID:            s.testUUID,
		Name:          "Test User",
		Email:         "test@example.com",
		Password:      s.hashedPass,
		Role:          "user",
		VerifiedEmail: false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// Helper to create test user list
func (s *userServiceTestSuite) createTestUserList() []domain.User {
	return []domain.User{
		{
			ID:            s.testUUID,
			Name:          "User One",
			Email:         "user1@example.com",
			Password:      s.hashedPass,
			Role:          "user",
			VerifiedEmail: false,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			ID:            s.testUUID2,
			Name:          "User Two",
			Email:         "user2@example.com",
			Password:      s.hashedPass,
			Role:          "admin",
			VerifiedEmail: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}
}

// ==================== GetUsers Tests ====================

func (s *userServiceTestSuite) TestGetUsers_Success() {
	req := &model.GetUserRequest{
		Page:   1,
		Limit:  10,
		Search: "",
	}

	testUsers := s.createTestUserList()
	var totalResults int64 = 2

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserRepo.EXPECT().
		GetAll(s.ctx, 10, 0, "").
		Return(testUsers, totalResults, nil)

	result, total, err := s.userService.GetUsers(s.ctx, req)

	s.NoError(err)
	s.Equal(totalResults, total)
	s.Len(result, 2)
	s.Equal("User One", result[0].Name)
	s.Equal("User Two", result[1].Name)
}

func (s *userServiceTestSuite) TestGetUsers_WithPagination() {
	req := &model.GetUserRequest{
		Page:   2,
		Limit:  5,
		Search: "test",
	}

	testUsers := s.createTestUserList()[:1]
	var totalResults int64 = 7

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	// offset = (2-1) * 5 = 5
	s.mockUserRepo.EXPECT().
		GetAll(s.ctx, 5, 5, "test").
		Return(testUsers, totalResults, nil)

	result, total, err := s.userService.GetUsers(s.ctx, req)

	s.NoError(err)
	s.Equal(totalResults, total)
	s.Len(result, 1)
}

func (s *userServiceTestSuite) TestGetUsers_ValidationError() {
	req := &model.GetUserRequest{
		Page:  -1,
		Limit: -1,
	}

	validationErr := errors.New("validation failed")

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(validationErr)

	result, total, err := s.userService.GetUsers(s.ctx, req)

	s.Error(err)
	s.Equal(myerrors.ErrInvalidRequest, err)
	s.Nil(result)
	s.Equal(int64(0), total)
}

func (s *userServiceTestSuite) TestGetUsers_RepositoryError() {
	req := &model.GetUserRequest{
		Page:  1,
		Limit: 10,
	}

	repoErr := myerrors.ErrGetUserFailed

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserRepo.EXPECT().
		GetAll(s.ctx, 10, 0, "").
		Return(nil, int64(0), repoErr)

	result, total, err := s.userService.GetUsers(s.ctx, req)

	s.Error(err)
	s.Equal(repoErr, err)
	s.Nil(result)
	s.Equal(int64(0), total)
}

// ==================== GetUserByID Tests ====================

func (s *userServiceTestSuite) TestGetUserByID_Success() {
	testUser := s.createTestUser()
	userID := s.testUUID.String()

	s.mockUserRepo.EXPECT().
		GetByID(s.ctx, userID).
		Return(testUser, nil)

	result, err := s.userService.GetUserByID(s.ctx, userID)

	s.NoError(err)
	s.Equal(testUser, result)
	s.Equal("Test User", result.Name)
	s.Equal("test@example.com", result.Email)
}

func (s *userServiceTestSuite) TestGetUserByID_UserNotFound() {
	userID := s.testUUID.String()

	s.mockUserRepo.EXPECT().
		GetByID(s.ctx, userID).
		Return(nil, myerrors.ErrUserNotFound)

	result, err := s.userService.GetUserByID(s.ctx, userID)

	s.Error(err)
	s.Equal(myerrors.ErrUserNotFound, err)
	s.Nil(result)
}

func (s *userServiceTestSuite) TestGetUserByID_RepositoryError() {
	userID := s.testUUID.String()
	repoErr := myerrors.ErrGetUserFailed

	s.mockUserRepo.EXPECT().
		GetByID(s.ctx, userID).
		Return(nil, repoErr)

	result, err := s.userService.GetUserByID(s.ctx, userID)

	s.Error(err)
	s.Equal(repoErr, err)
	s.Nil(result)
}

// ==================== GetUserByEmail Tests ====================

func (s *userServiceTestSuite) TestGetUserByEmail_Success() {
	testUser := s.createTestUser()
	email := "test@example.com"

	s.mockUserRepo.EXPECT().
		GetByEmail(s.ctx, email).
		Return(testUser, nil)

	result, err := s.userService.GetUserByEmail(s.ctx, email)

	s.NoError(err)
	s.Equal(testUser, result)
	s.Equal("test@example.com", result.Email)
}

func (s *userServiceTestSuite) TestGetUserByEmail_UserNotFound() {
	email := "nonexistent@example.com"

	s.mockUserRepo.EXPECT().
		GetByEmail(s.ctx, email).
		Return(nil, myerrors.ErrUserNotFound)

	result, err := s.userService.GetUserByEmail(s.ctx, email)

	s.Error(err)
	s.Equal(myerrors.ErrUserNotFound, err)
	s.Nil(result)
}

func (s *userServiceTestSuite) TestGetUserByEmail_RepositoryError() {
	email := "test@example.com"
	repoErr := myerrors.ErrGetUserFailed

	s.mockUserRepo.EXPECT().
		GetByEmail(s.ctx, email).
		Return(nil, repoErr)

	result, err := s.userService.GetUserByEmail(s.ctx, email)

	s.Error(err)
	s.Equal(repoErr, err)
	s.Nil(result)
}

// ==================== CreateUser Tests ====================

func (s *userServiceTestSuite) TestCreateUser_Success() {
	req := &model.CreateUserRequest{
		Name:     "New User",
		Email:    "newuser@example.com",
		Password: "password123",
		Role:     "user",
	}

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserRepo.EXPECT().
		Create(s.ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, user *domain.User) (*domain.User, error) {
			user.ID = s.testUUID
			return user, nil
		})

	result, err := s.userService.CreateUser(s.ctx, req)

	s.NoError(err)
	s.Equal(req.Name, result.Name)
	s.Equal(req.Email, result.Email)
	s.Equal(req.Role, result.Role)
	s.NotEqual("password123", result.Password) // Should be hashed
}

func (s *userServiceTestSuite) TestCreateUser_ValidationError() {
	req := &model.CreateUserRequest{
		Name:     "",
		Email:    "invalid-email",
		Password: "short",
		Role:     "",
	}

	validationErr := errors.New("validation failed")

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(validationErr)

	result, err := s.userService.CreateUser(s.ctx, req)

	s.Error(err)
	s.Equal(myerrors.ErrInvalidRequest, err)
	s.Nil(result)
}

func (s *userServiceTestSuite) TestCreateUser_EmailAlreadyInUse() {
	req := &model.CreateUserRequest{
		Name:     "New User",
		Email:    "existing@example.com",
		Password: "password123",
		Role:     "user",
	}

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserRepo.EXPECT().
		Create(s.ctx, gomock.Any()).
		Return(nil, myerrors.ErrEmailAlreadyInUse)

	result, err := s.userService.CreateUser(s.ctx, req)

	s.Error(err)
	s.Equal(myerrors.ErrEmailAlreadyInUse, err)
	s.Nil(result)
}

func (s *userServiceTestSuite) TestCreateUser_RepositoryError() {
	req := &model.CreateUserRequest{
		Name:     "New User",
		Email:    "newuser@example.com",
		Password: "password123",
		Role:     "user",
	}

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserRepo.EXPECT().
		Create(s.ctx, gomock.Any()).
		Return(nil, myerrors.ErrCreateUserFailed)

	result, err := s.userService.CreateUser(s.ctx, req)

	s.Error(err)
	s.Equal(myerrors.ErrCreateUserFailed, err)
	s.Nil(result)
}

// ==================== UpdateUser Tests ====================

func (s *userServiceTestSuite) TestUpdateUser_Success_NameOnly() {
	req := &model.UpdateUserRequest{
		UserID: s.testUUID.String(),
		Name:   "Updated Name",
	}

	updatedUser := &domain.User{
		ID:            s.testUUID,
		Name:          "Updated Name",
		Email:         "test@example.com",
		Password:      s.hashedPass,
		Role:          "user",
		VerifiedEmail: false,
	}

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserRepo.EXPECT().
		Update(s.ctx, gomock.Any()).
		Return(updatedUser, nil)

	result, err := s.userService.UpdateUser(s.ctx, req)

	s.NoError(err)
	s.Equal("Updated Name", result.Name)
}

func (s *userServiceTestSuite) TestUpdateUser_Success_EmailOnly() {
	req := &model.UpdateUserRequest{
		UserID: s.testUUID.String(),
		Email:  "newemail@example.com",
	}

	updatedUser := &domain.User{
		ID:            s.testUUID,
		Name:          "Test User",
		Email:         "newemail@example.com",
		Password:      s.hashedPass,
		Role:          "user",
		VerifiedEmail: false,
	}

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserRepo.EXPECT().
		Update(s.ctx, gomock.Any()).
		Return(updatedUser, nil)

	result, err := s.userService.UpdateUser(s.ctx, req)

	s.NoError(err)
	s.Equal("newemail@example.com", result.Email)
}

func (s *userServiceTestSuite) TestUpdateUser_Success_WithPassword() {
	req := &model.UpdateUserRequest{
		UserID:   s.testUUID.String(),
		Name:     "Updated Name",
		Password: "newpassword123",
	}

	updatedUser := &domain.User{
		ID:            s.testUUID,
		Name:          "Updated Name",
		Email:         "test@example.com",
		Password:      s.hashedPass,
		Role:          "user",
		VerifiedEmail: false,
	}

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserRepo.EXPECT().
		Update(s.ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, user *domain.User) (*domain.User, error) {
			// Verify password was hashed
			s.NotEqual("newpassword123", user.Password)
			return updatedUser, nil
		})

	result, err := s.userService.UpdateUser(s.ctx, req)

	s.NoError(err)
	s.Equal("Updated Name", result.Name)
}

func (s *userServiceTestSuite) TestUpdateUser_Success_AllFields() {
	req := &model.UpdateUserRequest{
		UserID:   s.testUUID.String(),
		Name:     "New Name",
		Email:    "newemail@example.com",
		Password: "newpassword123",
	}

	updatedUser := &domain.User{
		ID:            s.testUUID,
		Name:          "New Name",
		Email:         "newemail@example.com",
		Password:      s.hashedPass,
		Role:          "user",
		VerifiedEmail: false,
	}

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserRepo.EXPECT().
		Update(s.ctx, gomock.Any()).
		Return(updatedUser, nil)

	result, err := s.userService.UpdateUser(s.ctx, req)

	s.NoError(err)
	s.Equal("New Name", result.Name)
	s.Equal("newemail@example.com", result.Email)
}

func (s *userServiceTestSuite) TestUpdateUser_ValidationError() {
	req := &model.UpdateUserRequest{
		UserID: "",
	}

	validationErr := errors.New("validation failed")

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(validationErr)

	result, err := s.userService.UpdateUser(s.ctx, req)

	s.Error(err)
	s.Equal(myerrors.ErrInvalidRequest, err)
	s.Nil(result)
}

func (s *userServiceTestSuite) TestUpdateUser_EmptyUpdateFields() {
	req := &model.UpdateUserRequest{
		UserID:   s.testUUID.String(),
		Name:     "",
		Email:    "",
		Password: "",
	}

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	result, err := s.userService.UpdateUser(s.ctx, req)

	s.Error(err)
	s.Equal(myerrors.ErrInvalidRequest, err)
	s.Nil(result)
}

func (s *userServiceTestSuite) TestUpdateUser_EmailAlreadyInUse() {
	req := &model.UpdateUserRequest{
		UserID: s.testUUID.String(),
		Email:  "existing@example.com",
	}

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserRepo.EXPECT().
		Update(s.ctx, gomock.Any()).
		Return(nil, myerrors.ErrEmailAlreadyInUse)

	result, err := s.userService.UpdateUser(s.ctx, req)

	s.Error(err)
	s.Equal(myerrors.ErrEmailAlreadyInUse, err)
	s.Nil(result)
}

func (s *userServiceTestSuite) TestUpdateUser_UserNotFound() {
	req := &model.UpdateUserRequest{
		UserID: s.testUUID.String(),
		Name:   "Updated Name",
	}

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserRepo.EXPECT().
		Update(s.ctx, gomock.Any()).
		Return(nil, myerrors.ErrUserNotFound)

	result, err := s.userService.UpdateUser(s.ctx, req)

	s.Error(err)
	s.Equal(myerrors.ErrUserNotFound, err)
	s.Nil(result)
}

// ==================== UpdatePassOrVerify Tests ====================

func (s *userServiceTestSuite) TestUpdatePassOrVerify_Success_PasswordOnly() {
	req := &model.UpdatePassOrVerifyRequest{
		Password: "newpassword123",
	}

	userID := s.testUUID.String()

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserRepo.EXPECT().
		UpdatePassOrVerify(s.ctx, gomock.Any(), userID).
		DoAndReturn(func(_ context.Context, user *domain.User, _ string) error {
			// Verify password was hashed
			s.NotEqual("newpassword123", user.Password)
			return nil
		})

	err := s.userService.UpdatePassOrVerify(s.ctx, req, userID)

	s.NoError(err)
}

func (s *userServiceTestSuite) TestUpdatePassOrVerify_Success_VerifiedEmailOnly() {
	req := &model.UpdatePassOrVerifyRequest{
		VerifiedEmail: true,
	}

	userID := s.testUUID.String()

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserRepo.EXPECT().
		UpdatePassOrVerify(s.ctx, gomock.Any(), userID).
		Return(nil)

	err := s.userService.UpdatePassOrVerify(s.ctx, req, userID)

	s.NoError(err)
}

func (s *userServiceTestSuite) TestUpdatePassOrVerify_Success_BothFields() {
	req := &model.UpdatePassOrVerifyRequest{
		Password:      "newpassword123",
		VerifiedEmail: true,
	}

	userID := s.testUUID.String()

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserRepo.EXPECT().
		UpdatePassOrVerify(s.ctx, gomock.Any(), userID).
		Return(nil)

	err := s.userService.UpdatePassOrVerify(s.ctx, req, userID)

	s.NoError(err)
}

func (s *userServiceTestSuite) TestUpdatePassOrVerify_ValidationError() {
	req := &model.UpdatePassOrVerifyRequest{}

	validationErr := errors.New("validation failed")

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(validationErr)

	err := s.userService.UpdatePassOrVerify(s.ctx, req, s.testUUID.String())

	s.Error(err)
	s.Equal(myerrors.ErrInvalidRequest, err)
}

func (s *userServiceTestSuite) TestUpdatePassOrVerify_EmptyFields() {
	req := &model.UpdatePassOrVerifyRequest{
		Password:      "",
		VerifiedEmail: false,
	}

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	err := s.userService.UpdatePassOrVerify(s.ctx, req, s.testUUID.String())

	s.Error(err)
	s.Equal(myerrors.ErrInvalidRequest, err)
}

func (s *userServiceTestSuite) TestUpdatePassOrVerify_RepositoryError() {
	req := &model.UpdatePassOrVerifyRequest{
		Password: "newpassword123",
	}

	userID := s.testUUID.String()

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserRepo.EXPECT().
		UpdatePassOrVerify(s.ctx, gomock.Any(), userID).
		Return(myerrors.ErrUpdatePassOrVerifyFailed)

	err := s.userService.UpdatePassOrVerify(s.ctx, req, userID)

	s.Error(err)
	s.Equal(myerrors.ErrUpdatePassOrVerifyFailed, err)
}

// ==================== DeleteUser Tests ====================

func (s *userServiceTestSuite) TestDeleteUser_Success() {
	userID := s.testUUID.String()

	s.mockUserRepo.EXPECT().
		Delete(s.ctx, userID).
		Return(nil)

	err := s.userService.DeleteUser(s.ctx, userID)

	s.NoError(err)
}

func (s *userServiceTestSuite) TestDeleteUser_UserNotFound() {
	userID := s.testUUID.String()

	s.mockUserRepo.EXPECT().
		Delete(s.ctx, userID).
		Return(myerrors.ErrUserNotFound)

	err := s.userService.DeleteUser(s.ctx, userID)

	s.Error(err)
	s.Equal(myerrors.ErrUserNotFound, err)
}

func (s *userServiceTestSuite) TestDeleteUser_RepositoryError() {
	userID := s.testUUID.String()

	s.mockUserRepo.EXPECT().
		Delete(s.ctx, userID).
		Return(myerrors.ErrDeleteUserFailed)

	err := s.userService.DeleteUser(s.ctx, userID)

	s.Error(err)
	s.Equal(myerrors.ErrDeleteUserFailed, err)
}

// ==================== CreateGoogleUser Tests ====================

func (s *userServiceTestSuite) TestCreateGoogleUser_Success_NewUser() {
	req := &model.CreateGoogleUserRequest{
		Name:          "Google User",
		Email:         "google@example.com",
		VerifiedEmail: true,
	}

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserRepo.EXPECT().
		GetByEmail(s.ctx, req.Email).
		Return(nil, myerrors.ErrUserNotFound)

	s.mockUserRepo.EXPECT().
		Create(s.ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, user *domain.User) (*domain.User, error) {
			user.ID = s.testUUID
			return user, nil
		})

	result, err := s.userService.CreateGoogleUser(s.ctx, req)

	s.NoError(err)
	s.Equal(req.Name, result.Name)
	s.Equal(req.Email, result.Email)
	s.Equal(req.VerifiedEmail, result.VerifiedEmail)
}

func (s *userServiceTestSuite) TestCreateGoogleUser_Success_ExistingUser() {
	req := &model.CreateGoogleUserRequest{
		Name:          "Google User",
		Email:         "existing@example.com",
		VerifiedEmail: true,
	}

	existingUser := &domain.User{
		ID:            s.testUUID,
		Name:          "Existing User",
		Email:         "existing@example.com",
		Password:      s.hashedPass,
		Role:          "user",
		VerifiedEmail: false,
		CreatedAt:     time.Now().Add(-24 * time.Hour),
		UpdatedAt:     time.Now().Add(-24 * time.Hour),
	}

	updatedUser := &domain.User{
		ID:            s.testUUID,
		Name:          "Existing User",
		Email:         "existing@example.com",
		Password:      s.hashedPass,
		Role:          "user",
		VerifiedEmail: true,
		CreatedAt:     existingUser.CreatedAt,
		UpdatedAt:     time.Now(),
	}

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserRepo.EXPECT().
		GetByEmail(s.ctx, req.Email).
		Return(existingUser, nil)

	s.mockUserRepo.EXPECT().
		Update(s.ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, user *domain.User) (*domain.User, error) {
			s.True(user.VerifiedEmail)
			return updatedUser, nil
		})

	result, err := s.userService.CreateGoogleUser(s.ctx, req)

	s.NoError(err)
	s.Equal(existingUser.Email, result.Email)
	s.True(result.VerifiedEmail)
}

func (s *userServiceTestSuite) TestCreateGoogleUser_ValidationError() {
	req := &model.CreateGoogleUserRequest{
		Name:  "",
		Email: "invalid-email",
	}

	validationErr := errors.New("validation failed")

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(validationErr)

	result, err := s.userService.CreateGoogleUser(s.ctx, req)

	s.Error(err)
	s.Equal(myerrors.ErrInvalidRequest, err)
	s.Nil(result)
}

func (s *userServiceTestSuite) TestCreateGoogleUser_GetUserError() {
	req := &model.CreateGoogleUserRequest{
		Name:          "Google User",
		Email:         "google@example.com",
		VerifiedEmail: true,
	}

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserRepo.EXPECT().
		GetByEmail(s.ctx, req.Email).
		Return(nil, myerrors.ErrGetUserFailed)

	result, err := s.userService.CreateGoogleUser(s.ctx, req)

	s.Error(err)
	s.Equal(myerrors.ErrGetUserFailed, err)
	s.Nil(result)
}

func (s *userServiceTestSuite) TestCreateGoogleUser_CreateError() {
	req := &model.CreateGoogleUserRequest{
		Name:          "Google User",
		Email:         "google@example.com",
		VerifiedEmail: true,
	}

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserRepo.EXPECT().
		GetByEmail(s.ctx, req.Email).
		Return(nil, myerrors.ErrUserNotFound)

	s.mockUserRepo.EXPECT().
		Create(s.ctx, gomock.Any()).
		Return(nil, myerrors.ErrCreateUserFailed)

	result, err := s.userService.CreateGoogleUser(s.ctx, req)

	s.Error(err)
	s.Equal(myerrors.ErrCreateUserFailed, err)
	s.Nil(result)
}

func (s *userServiceTestSuite) TestCreateGoogleUser_UpdateError() {
	req := &model.CreateGoogleUserRequest{
		Name:          "Google User",
		Email:         "existing@example.com",
		VerifiedEmail: true,
	}

	existingUser := &domain.User{
		ID:            s.testUUID,
		Name:          "Existing User",
		Email:         "existing@example.com",
		Password:      s.hashedPass,
		Role:          "user",
		VerifiedEmail: false,
	}

	s.mockValidator.EXPECT().
		Validate(s.ctx, req).
		Return(nil)

	s.mockUserRepo.EXPECT().
		GetByEmail(s.ctx, req.Email).
		Return(existingUser, nil)

	s.mockUserRepo.EXPECT().
		Update(s.ctx, gomock.Any()).
		Return(nil, myerrors.ErrUpdateUserFailed)

	result, err := s.userService.CreateGoogleUser(s.ctx, req)

	s.Error(err)
	s.Equal(myerrors.ErrUpdateUserFailed, err)
	s.Nil(result)
}
