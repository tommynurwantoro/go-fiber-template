package repository

import (
	"app/internal/adapter/database/mocks"
	"app/internal/domain"
	"app/internal/domain/myerrors"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type tokenRepositoryTestSuite struct {
	suite.Suite
	ctx      context.Context
	mockCtrl *gomock.Controller
	mockDB   *mocks.MockDatabaseAdapter
	gormDB   *gorm.DB
	repo     *TokenRepositoryImpl
}

func TestTokenRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(tokenRepositoryTestSuite))
}

func (s *tokenRepositoryTestSuite) SetupTest() {
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
	s.repo = &TokenRepositoryImpl{DB: s.mockDB}
}

func (s *tokenRepositoryTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *tokenRepositoryTestSuite) makeToken(token string, userID uuid.UUID, tokenType domain.TokenType, expires time.Time) *domain.Token {
	return &domain.Token{
		Token:   token,
		UserID:  userID,
		Type:    tokenType,
		Expires: expires,
	}
}

func (s *tokenRepositoryTestSuite) TestCreate_Success() {
	userID := uuid.Must(uuid.NewV7())
	expires := time.Now().Add(time.Hour)
	token := s.makeToken("test-token", userID, domain.TokenTypeRefresh, expires)

	created, err := s.repo.Create(s.ctx, token)
	s.NoError(err)
	s.Require().NotNil(created)
	s.NotEqual(uuid.Nil, created.ID)
	s.Equal("test-token", created.Token)
	s.Equal(userID, created.UserID)
	s.Equal(domain.TokenTypeRefresh, created.Type)
}

func (s *tokenRepositoryTestSuite) TestCreate_Error() {
	sqlDB, err := s.gormDB.DB()
	s.Require().NoError(err)
	s.Require().NoError(sqlDB.Close())

	userID := uuid.Must(uuid.NewV7())
	token := s.makeToken("test-token", userID, domain.TokenTypeRefresh, time.Now())

	created, err := s.repo.Create(s.ctx, token)
	s.Error(err)
	s.Nil(created)
	s.True(errors.Is(err, myerrors.ErrSaveTokenFailed))
}

func (s *tokenRepositoryTestSuite) TestDelete_Success() {
	userID := uuid.Must(uuid.NewV7())
	expires := time.Now().Add(time.Hour)
	token := s.makeToken("test-token", userID, domain.TokenTypeRefresh, expires)

	_, err := s.repo.Create(s.ctx, token)
	s.Require().NoError(err)

	err = s.repo.Delete(s.ctx, domain.TokenTypeRefresh, userID.String())
	s.NoError(err)

	// Verify token is deleted
	_, err = s.repo.GetByTokenAndUserID(s.ctx, "test-token", userID.String())
	s.True(errors.Is(err, myerrors.ErrTokenNotFound))
}

func (s *tokenRepositoryTestSuite) TestDelete_Error() {
	sqlDB, err := s.gormDB.DB()
	s.Require().NoError(err)
	s.Require().NoError(sqlDB.Close())

	err = s.repo.Delete(s.ctx, domain.TokenTypeRefresh, uuid.Must(uuid.NewV7()).String())
	s.Error(err)
	s.True(errors.Is(err, myerrors.ErrDeleteTokenFailed))
}

func (s *tokenRepositoryTestSuite) TestDeleteAll_Success() {
	userID := uuid.Must(uuid.NewV7())
	expires := time.Now().Add(time.Hour)

	token1 := s.makeToken("token-1", userID, domain.TokenTypeRefresh, expires)
	token2 := s.makeToken("token-2", userID, domain.TokenTypeAccess, expires)

	_, err := s.repo.Create(s.ctx, token1)
	s.Require().NoError(err)
	_, err = s.repo.Create(s.ctx, token2)
	s.Require().NoError(err)

	err = s.repo.DeleteAll(s.ctx, userID.String())
	s.NoError(err)

	// Verify all tokens for user are deleted
	_, err = s.repo.GetByTokenAndUserID(s.ctx, "token-1", userID.String())
	s.True(errors.Is(err, myerrors.ErrTokenNotFound))
	_, err = s.repo.GetByTokenAndUserID(s.ctx, "token-2", userID.String())
	s.True(errors.Is(err, myerrors.ErrTokenNotFound))
}

func (s *tokenRepositoryTestSuite) TestDeleteAll_Error() {
	sqlDB, err := s.gormDB.DB()
	s.Require().NoError(err)
	s.Require().NoError(sqlDB.Close())

	err = s.repo.DeleteAll(s.ctx, uuid.Must(uuid.NewV7()).String())
	s.Error(err)
	s.True(errors.Is(err, myerrors.ErrDeleteAllTokenFailed))
}

func (s *tokenRepositoryTestSuite) TestGetByTokenAndUserID_Success() {
	userID := uuid.Must(uuid.NewV7())
	expires := time.Now().Add(time.Hour)
	token := s.makeToken("test-token", userID, domain.TokenTypeRefresh, expires)

	created, err := s.repo.Create(s.ctx, token)
	s.Require().NoError(err)

	found, err := s.repo.GetByTokenAndUserID(s.ctx, "test-token", userID.String())
	s.NoError(err)
	s.Require().NotNil(found)
	s.Equal(created.ID, found.ID)
	s.Equal("test-token", found.Token)
	s.Equal(userID, found.UserID)
	s.Equal(domain.TokenTypeRefresh, found.Type)
}

func (s *tokenRepositoryTestSuite) TestGetByTokenAndUserID_NotFound() {
	userID := uuid.Must(uuid.NewV7())

	found, err := s.repo.GetByTokenAndUserID(s.ctx, "nonexistent-token", userID.String())
	s.Error(err)
	s.Nil(found)
	s.True(errors.Is(err, myerrors.ErrTokenNotFound))
}

func (s *tokenRepositoryTestSuite) TestGetByTokenAndUserID_Error() {
	sqlDB, err := s.gormDB.DB()
	s.Require().NoError(err)
	s.Require().NoError(sqlDB.Close())

	found, err := s.repo.GetByTokenAndUserID(s.ctx, "test-token", uuid.Must(uuid.NewV7()).String())
	s.Error(err)
	s.Nil(found)
	s.True(errors.Is(err, myerrors.ErrGetTokenByUserIDFailed))
}
