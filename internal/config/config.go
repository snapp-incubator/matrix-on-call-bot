package config

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/go-playground/validator/v10"
)

const (
	app                = "matrix-on-call-bot"
	cfgDefaultFileName = "config"
	cfgFileExtension   = "yaml"
	cfgEnvPrefix       = "matrixoncallbot"
)

type (
	Config struct {
		Matrix   Matrix   `mapstructure:"matrix"`
		Database Database `mapstructure:"database"`
	}

	Matrix struct {
		URL         string `mapstructure:"url"`
		UserID      string `mapstructure:"userID"`
		Token       string `mapstructure:"token"`
		DisplayName string `mapstructure:"display-name"`
	}

	Database struct {
		Driver             string        `mapstructure:"driver"`
		Host               string        `mapstructure:"host"`
		Port               int           `mapstructure:"port"`
		DBName             string        `mapstructure:"db_name"`
		Username           string        `mapstructure:"username"`
		Password           string        `mapstructure:"password"`
		Timeout            time.Duration `mapstructure:"timeout"`
		ReadTimeout        time.Duration `mapstructure:"read_timeout"`
		WriteTimeout       time.Duration `mapstructure:"write_timeout"`
		ConnectionLifetime time.Duration `mapstructure:"connection_lifetime"`
		MaxOpenConnections int           `mapstructure:"max_open_connections"`
		MaxIdleConnections int           `mapstructure:"max_idle_connections"`
	}
)

// Validate validates Config struct.
func (c Config) Validate() error {
	return errors.Wrap(validator.New().Struct(c), "config validation failed")
}

// MySQLConnectionURI returns URI for connecting to a MySQL liked database.
func (d Database) MySQLConnectionURI() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?timeout=%s&readTimeout=%s&writeTimeout=%s&parseTime=True",
		d.Username,
		d.Password,
		d.Host,
		d.Port,
		d.DBName,
		d.Timeout.String(),
		d.ReadTimeout.String(),
		d.WriteTimeout.String())
}

// Init reads and validates application configs.
func Init() Config {
	var cfg Config

	read(app, cfgDefaultFileName, cfgFileExtension, &cfg, defaultConfig, cfgEnvPrefix)

	if err := cfg.Validate(); err != nil {
		logrus.Fatalf("failed to validate configurations: %s", err.Error())
	}

	return cfg
}

// read initializes a config struct using default, file, and environment variables.
func read(app, defaultFilename, fileExt string, cfg interface{}, defaultConfig string, envPrefix string) interface{} {
	//nolint:varnamelen
	v := viper.New()
	v.SetConfigType("yaml")

	if err := v.ReadConfig(bytes.NewReader([]byte(defaultConfig))); err != nil {
		logrus.Fatalf("error loading default configs: %s", err.Error())
	}

	v.SetConfigName(defaultFilename) // name of config defaultFilename (without extension)
	v.SetConfigType(fileExt)         // REQUIRED because of this bug: https://github.com/spf13/viper/issues/390
	v.SetEnvPrefix(envPrefix)
	v.AddConfigPath(fmt.Sprintf("/etc/%s/", app))
	v.AddConfigPath(fmt.Sprintf("$HOME/.%s", app))
	v.AddConfigPath(".")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	//nolint:errorlint
	switch err := v.MergeInConfig(); err.(type) {
	case nil:
	case *os.PathError:
		logrus.Warn("no config defaultFilename found. Using defaults and environment variables")
	default:
		logrus.Warnf("failed to load config defaultFilename: %s", err.Error())
	}

	if err := v.UnmarshalExact(&cfg); err != nil {
		logrus.Fatalf("failed to unmarshal config into struct: %s", err.Error())
	}

	return cfg
}
