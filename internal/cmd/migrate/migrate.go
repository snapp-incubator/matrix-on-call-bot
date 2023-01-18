package migrate

import (
	"github.com/pkg/errors"

	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file" // Imported for its side effects

	"github.com/golang-migrate/migrate/v4"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/snapp-incubator/matrix-on-call-bot/internal/config"
	"github.com/snapp-incubator/matrix-on-call-bot/internal/database"
)

const (
	flagPath = "path"
)

var ErrFlags = errors.New("error parsing flags")

func main(path string, cfg config.Database) error {
	oncallDB := database.WithRetry(database.Create, cfg)

	sqlDB, err := oncallDB.DB()
	if err != nil {
		logrus.WithError(err).Fatal("error in accessing sql DB instance")
	}

	defer func() {
		if err := sqlDB.Close(); err != nil {
			logrus.Errorf("db connection close error: %s", err.Error())
		}
	}()

	driver, err := mysql.WithInstance(sqlDB, &mysql.Config{})
	if err != nil {
		return errors.Wrap(err, "error creating driver")
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+path, "mysql", driver)
	if err != nil {
		return errors.Wrap(err, "error creating migrations")
	}

	if err := m.Up(); errors.Is(err, migrate.ErrNoChange) {
		logrus.Info("no change detected. All migrations have already been applied!")

		return nil
	} else if err != nil {
		return errors.Wrap(err, "error running migrations")
	}

	return nil
}

// Register migrate command.
func Register(root *cobra.Command, cfg config.Config) {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Provides DB migration functionality",

		// PreRunE does some validation, parsing, etc. Then populates migrationPath
		PreRunE: func(cmd *cobra.Command, args []string) error {
			path, err := cmd.Flags().GetString(flagPath)
			if err != nil {
				return ErrFlags
			}
			if path == "" {
				return ErrFlags
			}

			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := cmd.Flags().GetString(flagPath)
			if err != nil {
				return errors.Wrap(err, "error getting path")
			}

			if err := main(path, cfg.Database); err != nil {
				return errors.Wrap(err, "error running main)")
			}

			cmd.Println("migrations ran successfully")

			return nil
		},
	}

	cmd.Flags().StringP(flagPath, "p", "", "migration folder path")

	root.AddCommand(cmd)
}
