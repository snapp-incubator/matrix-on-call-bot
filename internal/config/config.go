package config

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/go-playground/validator/v10"

	"github.com/snapp-incubator/matrix-on-call-bot/internal/database"
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
		Driver  string           `mapstructure:"driver"`
		ConnStr string           `mapstructure:"conn-str"`
		Options database.Options `mapstructure:"options"`
	}
)

// Validate validates Config struct.
func (c Config) Validate() error {
	return errors.Wrap(validator.New().Struct(c), "config validation failed")
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
