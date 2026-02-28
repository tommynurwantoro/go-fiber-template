package database

import (
	"app/config"
	"fmt"

	"github.com/tommynurwantoro/golog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseAdapter interface {
	Ping() error
	GetDB() *gorm.DB
}

type Gorm struct {
	*gorm.DB
	Conf *config.Config `inject:"config"`
}

func (g *Gorm) Startup() error {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		g.Conf.Database.Host,
		g.Conf.Database.User,
		g.Conf.Database.Password,
		g.Conf.Database.Name,
		g.Conf.Database.Port,
		g.Conf.Database.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Info),
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		TranslateError:         true,
	})
	if err != nil {
		golog.Error("Failed to connect to database: %+v", err)
		return err
	}

	sqlDB, errDB := db.DB()
	if errDB != nil {
		golog.Error("Failed to get database instance: %+v", errDB)
		return errDB
	}

	// Config connection pooling
	sqlDB.SetMaxIdleConns(g.Conf.Database.MaxIdleConn)
	sqlDB.SetMaxOpenConns(g.Conf.Database.MaxOpenConn)
	sqlDB.SetConnMaxLifetime(g.Conf.Database.ConnMaxLifetime)

	g.DB = db

	return nil
}

func (g *Gorm) Shutdown() error {
	sqlDB, err := g.DB.DB()
	if err != nil {
		golog.Error("Failed to get database instance: %+v", err)
		return err
	}

	if err := sqlDB.Close(); err != nil {
		golog.Error("Failed to close database connection: %+v", err)
		return err
	}

	return nil
}

func (g *Gorm) Ping() error {
	sqlDB, err := g.DB.DB()
	if err != nil {
		golog.Error("Failed to get database instance: %+v", err)
		return err
	}

	return sqlDB.Ping()
}

func (g *Gorm) GetDB() *gorm.DB {
	return g.DB
}
