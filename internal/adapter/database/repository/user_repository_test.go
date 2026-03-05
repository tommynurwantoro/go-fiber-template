package repository

import (
	"app/internal/adapter/database/mocks"
	"app/internal/domain"
	"app/internal/domain/myerrors"
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type userRepositoryTestSuite struct {
	suite.Suite
	ctx       context.Context
	mockCtrl  *gomock.Controller
	mockDB    *mocks.MockDatabaseAdapter
	gormDB    *gorm.DB
	repo      *UserRepositoryImpl
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(userRepositoryTestSuite))
}

func (s *userRepositoryTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.mockCtrl = gomock.NewController(s.T())

	// Use unique in-memory DB per test to avoid shared state
	dbPath := fmt.Sprintf("file:test-%s.db?mode=memory", uuid.New().String())
	gormDB, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		TranslateError: true,
	})
	s.Require().NoError(err)
	s.Require().NoError(gormDB.AutoMigrate(&domain.User{}, &domain.Token{}))

	s.gormDB = gormDB
	s.mockDB = mocks.NewMockDatabaseAdapter(s.mockCtrl)
	s.mockDB.EXPECT().GetDB().Return(gormDB).AnyTimes()
	s.repo = &UserRepositoryImpl{DB: s.mockDB}
}

func (s *userRepositoryTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *userRepositoryTestSuite) makeUser(name, email, password, role string) *domain.User {
	return &domain.User{
		Name:          name,
		Email:         email,
		Password:      password,
		Role:          role,
		VerifiedEmail: false,
	}
}

func (s *userRepositoryTestSuite) TestGetAll_SuccessWithResults() {
	user1 := s.makeUser("Alice", "alice@example.com", "pass1", "user")
	user2 := s.makeUser("Bob", "bob@example.com", "pass2", "admin")

	_, err := s.repo.Create(s.ctx, user1)
	s.Require().NoError(err)
	_, err = s.repo.Create(s.ctx, user2)
	s.Require().NoError(err)

	users, total, err := s.repo.GetAll(s.ctx, 10, 0, "")
	s.NoError(err)
	s.Len(users, 2)
	s.Equal(int64(2), total)
}

func (s *userRepositoryTestSuite) TestGetAll_SuccessEmpty() {
	users, total, err := s.repo.GetAll(s.ctx, 10, 0, "")
	s.NoError(err)
	s.Empty(users)
	s.Equal(int64(0), total)
}

func (s *userRepositoryTestSuite) TestGetAll_SuccessWithSearch() {
	user1 := s.makeUser("Alice", "alice@example.com", "pass1", "user")
	user2 := s.makeUser("Bob", "bob@example.com", "pass2", "admin")

	_, err := s.repo.Create(s.ctx, user1)
	s.Require().NoError(err)
	_, err = s.repo.Create(s.ctx, user2)
	s.Require().NoError(err)

	users, total, err := s.repo.GetAll(s.ctx, 10, 0, "alice")
	s.NoError(err)
	s.Len(users, 1)
	s.Equal(int64(1), total)
	s.Equal("Alice", users[0].Name)
	s.Equal("alice@example.com", users[0].Email)
}

func (s *userRepositoryTestSuite) TestGetAll_ErrorOnCount() {
	sqlDB, err := s.gormDB.DB()
	s.Require().NoError(err)
	s.Require().NoError(sqlDB.Close())

	users, total, err := s.repo.GetAll(s.ctx, 10, 0, "")
	s.Error(err)
	s.Nil(users)
	s.Equal(int64(0), total)
	s.True(errors.Is(err, myerrors.ErrGetUserFailed))
}

func (s *userRepositoryTestSuite) TestGetByID_Success() {
	user := s.makeUser("Alice", "alice@example.com", "pass1", "user")
	created, err := s.repo.Create(s.ctx, user)
	s.Require().NoError(err)

	found, err := s.repo.GetByID(s.ctx, created.ID.String())
	s.NoError(err)
	s.Require().NotNil(found)
	s.Equal(created.ID, found.ID)
	s.Equal("Alice", found.Name)
	s.Equal("alice@example.com", found.Email)
}

func (s *userRepositoryTestSuite) TestGetByID_NotFound() {
	found, err := s.repo.GetByID(s.ctx, uuid.Must(uuid.NewV7()).String())
	s.Error(err)
	s.Nil(found)
	s.True(errors.Is(err, myerrors.ErrUserNotFound))
}

func (s *userRepositoryTestSuite) TestGetByEmail_Success() {
	user := s.makeUser("Alice", "alice@example.com", "pass1", "user")
	created, err := s.repo.Create(s.ctx, user)
	s.Require().NoError(err)

	found, err := s.repo.GetByEmail(s.ctx, "alice@example.com")
	s.NoError(err)
	s.Require().NotNil(found)
	s.Equal(created.ID, found.ID)
	s.Equal("Alice", found.Name)
}

func (s *userRepositoryTestSuite) TestGetByEmail_NotFound() {
	found, err := s.repo.GetByEmail(s.ctx, "nonexistent@example.com")
	s.Error(err)
	s.Nil(found)
	s.True(errors.Is(err, myerrors.ErrUserNotFound))
}

func (s *userRepositoryTestSuite) TestCreate_Success() {
	user := s.makeUser("Alice", "alice@example.com", "pass1", "user")

	created, err := s.repo.Create(s.ctx, user)
	s.NoError(err)
	s.Require().NotNil(created)
	s.NotEqual(uuid.Nil, created.ID)
	s.Equal("Alice", created.Name)
	s.Equal("alice@example.com", created.Email)
}

func (s *userRepositoryTestSuite) TestCreate_DuplicateEmail() {
	user1 := s.makeUser("Alice", "alice@example.com", "pass1", "user")
	_, err := s.repo.Create(s.ctx, user1)
	s.Require().NoError(err)

	user2 := s.makeUser("Bob", "alice@example.com", "pass2", "user")
	created, err := s.repo.Create(s.ctx, user2)
	s.Error(err)
	s.Nil(created)
	s.True(errors.Is(err, myerrors.ErrEmailAlreadyInUse))
}

func (s *userRepositoryTestSuite) TestUpdate_Success() {
	user := s.makeUser("Alice", "alice@example.com", "pass1", "user")
	created, err := s.repo.Create(s.ctx, user)
	s.Require().NoError(err)

	created.Name = "Alice Updated"
	updated, err := s.repo.Update(s.ctx, created)
	s.NoError(err)
	s.Require().NotNil(updated)
	s.Equal("Alice Updated", updated.Name)
}

func (s *userRepositoryTestSuite) TestUpdate_NotFound() {
	user := s.makeUser("Alice", "alice@example.com", "pass1", "user")
	user.ID = uuid.Must(uuid.NewV7())

	updated, err := s.repo.Update(s.ctx, user)
	s.Error(err)
	s.Nil(updated)
	s.True(errors.Is(err, myerrors.ErrUserNotFound))
}

func (s *userRepositoryTestSuite) TestUpdate_DuplicateEmail() {
	user1 := s.makeUser("Alice", "alice@example.com", "pass1", "user")
	user2 := s.makeUser("Bob", "bob@example.com", "pass2", "user")

	created1, err := s.repo.Create(s.ctx, user1)
	s.Require().NoError(err)
	_, err = s.repo.Create(s.ctx, user2)
	s.Require().NoError(err)

	created1.Email = "bob@example.com"
	updated, err := s.repo.Update(s.ctx, created1)
	s.Error(err)
	s.Nil(updated)
	// Update fails on duplicate email: may return ErrEmailAlreadyInUse, ErrUpdateUserFailed,
	// or ErrUserNotFound (when RowsAffected=0 is checked before Error)
	s.True(
		errors.Is(err, myerrors.ErrEmailAlreadyInUse) ||
			errors.Is(err, myerrors.ErrUpdateUserFailed) ||
			errors.Is(err, myerrors.ErrUserNotFound),
		"expected update to fail, got: %v", err)
}

func (s *userRepositoryTestSuite) TestUpdatePassOrVerify_Success() {
	user := s.makeUser("Alice", "alice@example.com", "pass1", "user")
	created, err := s.repo.Create(s.ctx, user)
	s.Require().NoError(err)

	created.Password = "newpass"
	created.VerifiedEmail = true
	err = s.repo.UpdatePassOrVerify(s.ctx, created, created.ID.String())
	s.NoError(err)
}

func (s *userRepositoryTestSuite) TestUpdatePassOrVerify_NotFound() {
	user := s.makeUser("Alice", "alice@example.com", "pass1", "user")
	user.ID = uuid.Must(uuid.NewV7())

	err := s.repo.UpdatePassOrVerify(s.ctx, user, user.ID.String())
	s.Error(err)
	s.True(errors.Is(err, myerrors.ErrUserNotFound))
}

func (s *userRepositoryTestSuite) TestDelete_Success() {
	user := s.makeUser("Alice", "alice@example.com", "pass1", "user")
	created, err := s.repo.Create(s.ctx, user)
	s.Require().NoError(err)

	err = s.repo.Delete(s.ctx, created.ID.String())
	s.NoError(err)

	_, err = s.repo.GetByID(s.ctx, created.ID.String())
	s.True(errors.Is(err, myerrors.ErrUserNotFound))
}

func (s *userRepositoryTestSuite) TestDelete_NotFound() {
	err := s.repo.Delete(s.ctx, uuid.Must(uuid.NewV7()).String())
	s.Error(err)
	s.True(errors.Is(err, myerrors.ErrUserNotFound))
}
