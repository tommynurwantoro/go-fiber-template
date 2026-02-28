package validator

import "context"

//go:generate mockgen -destination=mocks/validator.go -package=mocks -source=validator.go
type Validator interface {
	Validate(ctx context.Context, data any) error
}
