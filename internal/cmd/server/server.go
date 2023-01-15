package server

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/snapp-incubator/matrix-on-call-bot/internal/config"
	"github.com/snapp-incubator/matrix-on-call-bot/internal/database"
	"github.com/snapp-incubator/matrix-on-call-bot/internal/matrix"
	"github.com/snapp-incubator/matrix-on-call-bot/internal/model"
)

const sigChanSize = 2

func main(cfg config.Config) {
	oncallDB := database.WithRetry(database.Create, cfg.Database)

	roomRepo := &model.SQLRoomRepo{DB: oncallDB}
	shiftRepo := &model.SQLShiftRepo{DB: oncallDB}
	followUpRepo := &model.SQLFollowUpRepo{DB: oncallDB}

	bot, err := matrix.New(cfg.Matrix.URL, cfg.Matrix.UserID, cfg.Matrix.Token,
		cfg.Matrix.DisplayName, roomRepo, shiftRepo, followUpRepo)
	if err != nil {
		logrus.WithField("error", err.Error()).Error("cannot create bot instance")
	}

	if err := bot.RegisterListeners(); err != nil {
		logrus.WithField("error", err.Error()).Fatalf("couldn't register listeners")
	}

	sigChan := make(chan os.Signal, sigChanSize)
	// add any other syscalls that you want to be notified with
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logrus.Info("bot is started!")
	bot.Run()

	<-sigChan

	logrus.Info("stopping bot loop!")
	bot.Stop()

	logrus.Info("closing DB connections!")

	sqlDB, err := oncallDB.DB()
	if err != nil {
		logrus.WithError(err).Fatal("error in accessing sql DB instance")
	}

	if err := sqlDB.Close(); err != nil {
		logrus.WithError(err).Error("error in closing connection to database")
	}
}

// Register registers server command for matrix-matrix-on-call-bot binary.
func Register(root *cobra.Command, cfg config.Config) {
	root.AddCommand(
		&cobra.Command{
			Use:   "server",
			Short: "Run on call bot server component",
			Run: func(cmd *cobra.Command, args []string) {
				main(cfg)
			},
		},
	)
}
