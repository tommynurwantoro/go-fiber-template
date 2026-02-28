package validator_test

import (
	"app/internal/pkg/validator"
	"context"
	"testing"

	validatormocks "app/internal/pkg/validator/mocks"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type MockUser struct {
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
	Age   int    `validate:"required,gt=0"`
}

type validatorTestSuite struct {
	suite.Suite
	ctx           context.Context
	mockCtrl      *gomock.Controller
	mockValidator *validatormocks.MockValidator
	validator     validator.Validator
}

func TestValidatorTestSuite(t *testing.T) {
	suite.Run(t, new(validatorTestSuite))
}

func (t *validatorTestSuite) SetupTest() {
	t.ctx = context.Background()
	t.mockCtrl = gomock.NewController(t.T())
	t.mockValidator = validatormocks.NewMockValidator(t.mockCtrl)
	t.validator = validator.NewGoValidator()
}

func (t *validatorTestSuite) TearDownTest() {
	t.mockCtrl.Finish()
}

func (t *validatorTestSuite) TestValidateNoError() {
	mockUser := MockUser{Name: "Jony", Email: "joni@gmail.com", Age: 41}

	err := t.validator.Validate(t.ctx, mockUser)
	t.Nil(err)
}

func (t *validatorTestSuite) TestValidateError() {
	mockUser := MockUser{Name: "Jony"}

	err := t.validator.Validate(t.ctx, mockUser)
	t.Error(err)
}

func (t *validatorTestSuite) TestValidateStrongPassword() {
	type Register struct {
		Password string `validate:"strong-password"`
	}

	mockRegister := Register{Password: "!@#123Asd"}

	err := t.validator.Validate(t.ctx, mockRegister)
	t.Nil(err)
}

func (t *validatorTestSuite) TestValidateStrongPasswordError() {
	type Register struct {
		Password string `validate:"strong-password"`
	}

	mockRegister := Register{Password: "passwordkuat"}

	err := t.validator.Validate(t.ctx, mockRegister)
	t.Error(err)
}
