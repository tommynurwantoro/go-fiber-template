package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Token struct {
	ID        uuid.UUID `gorm:"primaryKey;not null" json:"id"`
	Token     string    `gorm:"not null" json:"token"`
	UserID    uuid.UUID `gorm:"not null"`
	Type      TokenType `gorm:"not null" json:"type"`
	Expires   time.Time `gorm:"not null" json:"expires"`
	CreatedAt time.Time `gorm:"autoCreateTime:milli"`
	UpdatedAt time.Time `gorm:"autoCreateTime:milli;autoUpdateTime:milli"`
	User      *User     `gorm:"foreignKey:user_id;references:id"`
}

func (token *Token) BeforeCreate(_ *gorm.DB) error {
	token.ID = uuid.Must(uuid.NewV7())
	return nil
}

type TokenType string

const (
	TokenTypeAccess        TokenType = "access"
	TokenTypeRefresh       TokenType = "refresh"
	TokenTypeResetPassword TokenType = "resetPassword"
	TokenTypeVerifyEmail   TokenType = "verifyEmail"
)

func (t TokenType) String() string {
	return string(t)
}
