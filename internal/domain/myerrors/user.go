package myerrors

import "errors"

var (
	ErrUserNotFound             = errors.New("user not found")
	ErrEmailAlreadyInUse        = errors.New("email already in use")
	ErrGetUserFailed            = errors.New("failed to get user")
	ErrCreateUserFailed         = errors.New("failed to create user")
	ErrUpdateUserFailed         = errors.New("failed to update user")
	ErrDeleteUserFailed         = errors.New("failed to delete user")
	ErrUpdatePassOrVerifyFailed = errors.New("failed to update user password or verifiedEmail")
	ErrInvalidEmailOrPassword   = errors.New("invalid email or password")
	ErrInvalidPassword          = errors.New("invalid password")
)
