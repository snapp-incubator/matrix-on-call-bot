package database

import (
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/snapp-incubator/matrix-on-call-bot/internal/config"
)

const (
	healthCheckInterval = 1
	maxAttempts         = 60
)

// Create creates a database connection.
func Create(cfg config.Database) (*gorm.DB, error) {
	var dialect gorm.Dialector

	switch strings.ToLower(cfg.Driver) {
	case "mysql":
		dialect = mysql.Open(cfg.MySQLConnectionURI())
	//nolint:godox
	// TODO: As our migrations are not Postgresql compatible, we don't support postgres for now.
	// case "postgres", "postgresql":
	//	dialect = postgres.Open(connStr)
	default:
		return nil, errors.New("unknown database driver")
	}

	database, err := gorm.Open(dialect, &gorm.Config{})
	if err != nil {
		return nil, errors.Wrap(err, "error opening connection to db")
	}

	sqlDB, err := database.DB()
	if err != nil {
		return nil, errors.Wrap(err, "error in accessing sql DB instance")
	}

	sqlDB.SetConnMaxLifetime(cfg.ConnectionLifetime)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConnections)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConnections)

	return database, nil
}

// WithRetry provides functionality for having retry for connecting to database.
func WithRetry(
	fn func(cfg config.Database) (*gorm.DB, error),
	cfg config.Database,
) *gorm.DB {
	for i := 0; i < maxAttempts; i++ {
		db, err := fn(cfg)
		if err == nil {
			return db
		}

		logrus.Errorf(
			"cannot connect to database. Waiting %d second. Error is: %s",
			healthCheckInterval, err.Error(),
		)

		time.Sleep(healthCheckInterval * time.Second)
	}

	logrus.Fatalf("could not connect to database after %d attempts", maxAttempts)

	return nil
}
