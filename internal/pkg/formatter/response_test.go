package formatter_test

import (
	"testing"

	"app/internal/pkg/formatter"

	"github.com/stretchr/testify/assert"
)

func TestResponse(t *testing.T) {
	t.Run("new success response", func(t *testing.T) {
		res := formatter.NewSuccessResponse(formatter.Success, "success", "data")

		assert.Equal(t, res.Status, "success")
	})

	t.Run("new error response", func(t *testing.T) {
		res := formatter.NewErrorResponse(formatter.InternalServerError, "unexpected", "12345")

		assert.Equal(t, res.Status, "APP05")
		assert.Equal(t, res.Message, "unexpected")
	})

	t.Run("new error response list", func(t *testing.T) {
		res := formatter.NewErrorResponseList(formatter.InternalServerError, "unexpected", "12345", "error")

		assert.Equal(t, res.Status, "APP05")
		assert.Equal(t, res.ErrorList, "error")
	})

	t.Run("new custom success response", func(t *testing.T) {
		res := formatter.NewSuccessResponse(formatter.Success, "success", "data")

		assert.Equal(t, res.Status, "success")
	})

	t.Run("new success response with metadata", func(t *testing.T) {
		res := formatter.NewSuccessResponseWithMetadata(formatter.Success, "success", "data", formatter.Metadata{
			Page:         1,
			Limit:        10,
			TotalPages:   int64(100),
			TotalResults: int64(1000),
		})

		assert.Equal(t, res.Status, "success")
		assert.Equal(t, res.Data, "data")
		assert.Equal(t, res.Metadata.Page, 1)
		assert.Equal(t, res.Metadata.Limit, 10)
		assert.Equal(t, res.Metadata.TotalPages, int64(100))
		assert.Equal(t, res.Metadata.TotalResults, int64(1000))
	})
}
