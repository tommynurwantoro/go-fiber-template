package cmd

import (
	"app/config"
	"app/internal/bootstrap"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/tommynurwantoro/golog"
)

func RunService() *cobra.Command {
	command := &cobra.Command{
		Use:     "service",
		Aliases: []string{"svc"},
		Short:   "Run the service",
		Run: func(_ *cobra.Command, _ []string) {
			// Load env variables from .env file
			_ = godotenv.Load(".env")

			// Load configurations
			conf := config.Config{}
			conf.Load()

			// Initialize Logger
			loggerConfig := golog.Config{
				App:           conf.AppName,
				AppVer:        conf.AppVersion,
				Env:           conf.Environment,
				FileLocation:  conf.Log.FileLocation,
				FileMaxSize:   conf.Log.FileMaxSize,
				FileMaxBackup: conf.Log.FileMaxBackup,
				FileMaxAge:    conf.Log.FileMaxAge,
				Stdout:        conf.Log.Stdout,
			}
			golog.Load(loggerConfig)

			bootstrap.RunService(&conf)
		},
	}

	return command
}
