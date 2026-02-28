package rest

import (
	"errors"
	"strings"

	"app/internal/pkg/formatter"
	"app/internal/pkg/validator"

	"github.com/gofiber/fiber/v2"
)

func ErrorHandler(codeMap map[error]formatter.Status, statusMap map[error]int) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		message := err.Error()
		errList := make(map[string]any, 0)

		// Status code defaults to 500
		httpStatus := fiber.StatusInternalServerError

		// if error is a validator.ErrorMap
		if _err, ok := err.(*validator.ErrorMap); ok {
			message, errList = makeErrorMap(_err.Error())
			err = fiber.ErrBadRequest
		}

		// Retrieve the custom status code if it's a *fiber.Error
		var e *fiber.Error
		if errors.As(err, &e) {
			httpStatus = e.Code
		} else {
			httpStatus = gethttpstatus(err, statusMap)
		}

		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		c.Status(httpStatus)

		code := getcode(err, codeMap)
		traceID, ok := c.Locals("traceId").(string)
		if !ok {
			traceID = ""
		}

		if len(errList) > 0 {
			return c.JSON(formatter.NewErrorResponseList(code, message, traceID, errList))
		}

		return c.JSON(formatter.NewErrorResponse(code, message, traceID))
	}
}

func makeErrorMap(er string) (string, map[string]any) {
	err := make(map[string]any, 0)
	message := ""
	errorMsg := strings.Split(er, ";")
	for _, msg := range errorMsg {
		errorList := strings.Split(msg, ":")

		message = strings.Join(errorList[1:], ":")
		if len(errorList) > 2 {
			err[errorList[0]] = strings.Join(errorList[1:], ":")
		} else {
			err[errorList[0]] = errorList[1]
		}
	}

	return message, err
}

func getcode(err error, codeMap map[error]formatter.Status) formatter.Status {
	if _, ok := err.(*validator.ErrorMap); ok {
		return formatter.InvalidRequest
	}
	for key, val := range codeMap {
		if errors.Is(err, key) {
			return val
		}
	}

	return formatter.InternalServerError
}

func gethttpstatus(err error, statusMap map[error]int) int {
	if _, ok := err.(*validator.ErrorMap); ok {
		return fiber.StatusBadRequest
	}
	if errFiber, ok := err.(*fiber.Error); ok {
		return errFiber.Code
	}
	for key, val := range statusMap {
		if errors.Is(err, key) {
			return val
		}
	}

	return fiber.StatusInternalServerError
}
