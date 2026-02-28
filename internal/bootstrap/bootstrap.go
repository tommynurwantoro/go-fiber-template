package bootstrap

import (
	"app/config"
	"app/internal/adapter/rest"
	"app/internal/pkg/validator"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/tommynurwantoro/golog"
	"github.com/tommynurwantoro/gontainer"
)

var appContainer = gontainer.New()

func RunService(conf *config.Config) {
	ctx, cancel := context.WithCancel(context.Background())

	appContainer.RegisterService("config", conf)
	appContainer.RegisterService("validator", validator.NewGoValidator())

	RegisterAdapters()
	RegisterServices()
	RegisterMiddleware()
	RegisterHandlers()

	// Startup the container
	if err := appContainer.Ready(); err != nil {
		golog.Panic("Failed to populate service", err)
	}

	// Start server
	fiberSvc := appContainer.GetServiceOrNil("rest")
	fiberApp, ok := fiberSvc.(*rest.Fiber)
	if !ok {
		golog.Panic("Failed to get rest service from container", errors.New("rest service not found"))
	}

	serverErrors := make(chan error, 1)
	go func() {
		golog.Info(fmt.Sprintf("Listening on %s:%d", fiberApp.Conf.Http.Host, fiberApp.Conf.Http.Port))
		serverErrors <- fiberApp.Listen(fmt.Sprintf("%s:%d", fiberApp.Conf.Http.Host, fiberApp.Conf.Http.Port))
	}()

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		select {
		case err := <-serverErrors:
			golog.Error("Server error: %v", err)
			cancel()
		case <-quit:
			golog.Info("Signal termination received")
			cancel()
		}
	}()

	<-ctx.Done()

	golog.Info("Cleaning up resources...")

	appContainer.Shutdown()

	golog.Info("Server exited")
}
