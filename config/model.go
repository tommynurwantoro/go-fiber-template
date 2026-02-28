package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	AppName     string         `mapstructure:"app_name"`
	AppVersion  string         `mapstructure:"app_version"`
	Environment string         `mapstructure:"environment"`
	Http        HttpConfig     `mapstructure:"http"`
	Log         LogConfig      `mapstructure:"log"`
	Database    DatabaseConfig `mapstructure:"database"`
	JWT         JWTConfig      `mapstructure:"jwt"`
	SMTP        SMTPConfig     `mapstructure:"smtp"`
	OAuth2      OAuth2Config   `mapstructure:"oauth2"`
}

type HttpConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
}

type LogConfig struct {
	FileLocation  string `mapstructure:"file_location"`
	FileMaxSize   int    `mapstructure:"file_max_size"`
	FileMaxBackup int    `mapstructure:"file_max_backup"`
	FileMaxAge    int    `mapstructure:"file_max_age"`
	Stdout        bool   `mapstructure:"stdout"`
}

type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Name            string        `mapstructure:"name"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxIdleConn     int           `mapstructure:"max_idle_conn"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	MaxOpenConn     int           `mapstructure:"max_open_conn"`
}

type JWTConfig struct {
	Secret              string        `mapstructure:"secret"`
	Expire              time.Duration `mapstructure:"expire"`
	RefreshExpire       int           `mapstructure:"refresh_expire"`
	ResetPasswordExpire time.Duration `mapstructure:"reset_password_expire"`
	VerifyEmailExpire   time.Duration `mapstructure:"verify_email_expire"`
}

type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
}

type OAuth2Config struct {
	GoogleClientID     string `mapstructure:"google_client_id"`
	GoogleClientSecret string `mapstructure:"google_client_secret"`
	RedirectURL        string `mapstructure:"redirect_url"`
}

func (c *Config) Load() {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("failed to read config file: %w", err))
	}

	// Setup environment variable mapping usinf struct tags
	viper.AutomaticEnv()

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err = viper.Unmarshal(c)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal config: %w", err))
	}
}
