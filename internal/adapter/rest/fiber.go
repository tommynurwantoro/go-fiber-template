package rest

import (
	"app/config"
	"app/internal/pkg/middleware"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

type Fiber struct {
	*fiber.App
	Conf *config.Config `inject:"config"`
}

func (f *Fiber) Startup() error {
	f.App = fiber.New(fiber.Config{
		Prefork:       f.Conf.Environment == "production",
		CaseSensitive: true,
		ServerHeader:  "Fiber",
		AppName:       f.Conf.AppName,
		ErrorHandler:  ErrorHandler(CodeMap, StatusMap),
		JSONEncoder:   json.Marshal,
		JSONDecoder:   json.Unmarshal,
		ReadTimeout:   f.Conf.Http.ReadTimeout,
		WriteTimeout:  f.Conf.Http.WriteTimeout,
	})

	// Middleware setup
	f.Use(requestid.New(requestid.Config{
		ContextKey: "traceId",
	}))
	f.Use("/v1/auth", middleware.LimiterConfig())
	f.Use(helmet.New())
	f.Use(compress.New())
	f.Use(cors.New())
	f.Use(middleware.RecoverConfig())

	return nil
}

func (f *Fiber) Shutdown() error {
	return f.App.Shutdown()
}
