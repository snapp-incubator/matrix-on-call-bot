package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	// "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	healthCheckInterval = 1
	maxAttempts         = 60
)

// Options represents a struct for creating database connection configurations.
type Options struct {
	ConnectionLifetime time.Duration `mapstructure:"connection-lifetime"`
	MaxOpenConnections int           `mapstructure:"max-open-connections"`
	MaxIdleConnections int           `mapstructure:"max-idle-connections"`
}

// Create creates a database connection.
func Create(driver string, connStr string, options Options) (*gorm.DB, error) {
	var dialect gorm.Dialector

	switch strings.ToLower(driver) {
	case "mysql":
		dialect = mysql.Open(connStr)
	//case "postgres", "postgresql":
	//	dialect = postgres.Open(connStr)
	default:
		return nil, fmt.Errorf("uknown database driver `%s`", driver)
	}

	database, err := gorm.Open(dialect, &gorm.Config{})
	if err != nil {
		return nil, errors.Wrap(err, "error opening connection to db")
	}

	sqlDB, err := database.DB()
	if err != nil {
		return nil, errors.Wrap(err, "error in accessing sql DB instance")
	}

	sqlDB.SetConnMaxLifetime(options.ConnectionLifetime)
	sqlDB.SetMaxOpenConns(options.MaxOpenConnections)
	sqlDB.SetMaxIdleConns(options.MaxIdleConnections)

	return database, nil
}

// WithRetry provides functionality for having retry for connecting to database.
func WithRetry(
	fn func(driver string, connStr string, options Options,
	) (*gorm.DB, error), driver string, connStr string, options Options,
) *gorm.DB {
	for i := 0; i < maxAttempts; i++ {
		db, err := fn(driver, connStr, options)
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
