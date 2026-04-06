package middleware

import (
	"errors"
	"fmt"
	"testing"

	"app/internal/pkg/formatter"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// Test errors for use in tests
var (
	testErrNotFound     = errors.New("not found")
	testErrUnauthorized = errors.New("unauthorized")
	testErrBadRequest   = errors.New("bad request")
	testErrUnknown      = errors.New("unknown error")
)

func TestGetCode(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		codeMap map[error]formatter.Status
		want    formatter.Status
	}{
		{
			name: "error matches key in map",
			err:  testErrNotFound,
			codeMap: map[error]formatter.Status{
				testErrNotFound: formatter.DataNotFound,
			},
			want: formatter.DataNotFound,
		},
		{
			name: "error wrapped with fmt.Errorf using %w",
			err:  fmt.Errorf("wrapped: %w", testErrNotFound),
			codeMap: map[error]formatter.Status{
				testErrNotFound: formatter.DataNotFound,
			},
			want: formatter.DataNotFound,
		},
		{
			name: "error wrapped multiple times",
			err:  fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", testErrNotFound)),
			codeMap: map[error]formatter.Status{
				testErrNotFound: formatter.DataNotFound,
			},
			want: formatter.DataNotFound,
		},
		{
			name: "multiple entries in map, first matches",
			err:  testErrNotFound,
			codeMap: map[error]formatter.Status{
				testErrNotFound:     formatter.DataNotFound,
				testErrUnauthorized: formatter.Unauthorized,
			},
			want: formatter.DataNotFound,
		},
		{
			name: "multiple entries in map, second matches",
			err:  testErrUnauthorized,
			codeMap: map[error]formatter.Status{
				testErrNotFound:     formatter.DataNotFound,
				testErrUnauthorized: formatter.Unauthorized,
			},
			want: formatter.Unauthorized,
		},
		{
			name: "error does not match any key",
			err:  testErrUnknown,
			codeMap: map[error]formatter.Status{
				testErrNotFound:     formatter.DataNotFound,
				testErrUnauthorized: formatter.Unauthorized,
			},
			want: formatter.InternalServerError,
		},
		{
			name:    "empty map",
			err:     testErrNotFound,
			codeMap: map[error]formatter.Status{},
			want:    formatter.InternalServerError,
		},
		{
			name: "nil error",
			err:  nil,
			codeMap: map[error]formatter.Status{
				testErrNotFound: formatter.DataNotFound,
			},
			want: formatter.InternalServerError,
		},
		{
			name:    "nil error with empty map",
			err:     nil,
			codeMap: map[error]formatter.Status{},
			want:    formatter.InternalServerError,
		},
		{
			name: "error matches using errors.Is with wrapped error",
			err:  fmt.Errorf("outer: %w", testErrBadRequest),
			codeMap: map[error]formatter.Status{
				testErrBadRequest: formatter.InvalidRequest,
			},
			want: formatter.InvalidRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getcode(tt.err, tt.codeMap)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetHTTPStatus(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		statusMap map[error]int
		want      int
	}{
		{
			name: "error matches key in map",
			err:  testErrNotFound,
			statusMap: map[error]int{
				testErrNotFound: fiber.StatusNotFound,
			},
			want: fiber.StatusNotFound,
		},
		{
			name: "error wrapped with fmt.Errorf using %w",
			err:  fmt.Errorf("wrapped: %w", testErrNotFound),
			statusMap: map[error]int{
				testErrNotFound: fiber.StatusNotFound,
			},
			want: fiber.StatusNotFound,
		},
		{
			name: "error wrapped multiple times",
			err:  fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", testErrNotFound)),
			statusMap: map[error]int{
				testErrNotFound: fiber.StatusNotFound,
			},
			want: fiber.StatusNotFound,
		},
		{
			name: "multiple entries in map, first matches",
			err:  testErrNotFound,
			statusMap: map[error]int{
				testErrNotFound:     fiber.StatusNotFound,
				testErrUnauthorized: fiber.StatusUnauthorized,
			},
			want: fiber.StatusNotFound,
		},
		{
			name: "multiple entries in map, second matches",
			err:  testErrUnauthorized,
			statusMap: map[error]int{
				testErrNotFound:     fiber.StatusNotFound,
				testErrUnauthorized: fiber.StatusUnauthorized,
			},
			want: fiber.StatusUnauthorized,
		},
		{
			name: "error does not match any key",
			err:  testErrUnknown,
			statusMap: map[error]int{
				testErrNotFound:     fiber.StatusNotFound,
				testErrUnauthorized: fiber.StatusUnauthorized,
			},
			want: fiber.StatusInternalServerError,
		},
		{
			name:      "empty map",
			err:       testErrNotFound,
			statusMap: map[error]int{},
			want:      fiber.StatusInternalServerError,
		},
		{
			name: "nil error",
			err:  nil,
			statusMap: map[error]int{
				testErrNotFound: fiber.StatusNotFound,
			},
			want: fiber.StatusInternalServerError,
		},
		{
			name:      "nil error with empty map",
			err:       nil,
			statusMap: map[error]int{},
			want:      fiber.StatusInternalServerError,
		},
		{
			name: "error matches using errors.Is with wrapped error",
			err:  fmt.Errorf("outer: %w", testErrBadRequest),
			statusMap: map[error]int{
				testErrBadRequest: fiber.StatusBadRequest,
			},
			want: fiber.StatusBadRequest,
		},
		{
			name: "different HTTP status codes",
			err:  testErrBadRequest,
			statusMap: map[error]int{
				testErrBadRequest:   fiber.StatusBadRequest,
				testErrUnauthorized: fiber.StatusUnauthorized,
				testErrNotFound:     fiber.StatusNotFound,
			},
			want: fiber.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := gethttpstatus(tt.err, tt.statusMap)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Test with actual wrapped errors using fmt.Errorf
func TestGetCodeWithWrappedErrors(t *testing.T) {
	baseErr := errors.New("base error")
	wrappedErr := fmt.Errorf("wrapper: %w", baseErr)

	codeMap := map[error]formatter.Status{
		baseErr: formatter.DataNotFound,
	}

	// Test that wrapped error matches base error
	result := getcode(wrappedErr, codeMap)
	assert.Equal(t, formatter.DataNotFound, result)
}

func TestGetHTTPStatusWithWrappedErrors(t *testing.T) {
	baseErr := errors.New("base error")
	wrappedErr := fmt.Errorf("wrapper: %w", baseErr)

	statusMap := map[error]int{
		baseErr: fiber.StatusBadRequest,
	}

	// Test that wrapped error matches base error
	result := gethttpstatus(wrappedErr, statusMap)
	assert.Equal(t, fiber.StatusBadRequest, result)
}
