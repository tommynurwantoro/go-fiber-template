package middleware

import (
	"strings"
	"time"

	"app/internal/pkg/formatter"

	"github.com/gofiber/fiber/v2"
	"github.com/tommynurwantoro/golog"
)

func Log(codeMap map[error]formatter.Status, statusMap map[error]int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		startTime := time.Now()
		req := c.Request()
		resp := c.Response()
		reqBody := c.Body()
		reqHeader := req.Header.Header()
		correlationID := c.Get("X-Correlation-ID")
		if correlationID != "" {
			c.Locals("traceId", correlationID)
		}

		// Set context value
		c.Locals("srcIP", c.Get("x-forwarded-for"))
		c.Locals("port", c.Port())
		c.Locals("path", c.Path())

		var err error
		if _err := c.Next(); _err != nil {
			err = _err
		}

		if err != nil && strings.HasPrefix(err.Error(), "Cannot ") {
			return nil
		}

		if c.Path() == "/ping" || c.Path() == "/ready" {
			return nil
		}

		statusCode := formatter.Success
		httpStatus := resp.StatusCode()
		if err != nil {
			httpStatus = gethttpstatus(err, statusMap)
			statusCode = getcode(err, codeMap)
		}

		logMsg := golog.LogModel{
			Header:       reqHeader,
			Request:      reqBody,
			HttpStatus:   uint64(httpStatus),
			StatusCode:   statusCode.String(),
			Response:     string(resp.Body()),
			ResponseTime: time.Since(startTime),
			Error:        err,
		}
		golog.TDR(logMsg)

		return err
	}
}
