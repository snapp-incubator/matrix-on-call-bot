package server

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/snapp-incubator/matrix-on-call-bot/internal/config"
	"github.com/snapp-incubator/matrix-on-call-bot/internal/database"
	internalhttp "github.com/snapp-incubator/matrix-on-call-bot/internal/http"
	"github.com/snapp-incubator/matrix-on-call-bot/internal/matrix"
	"github.com/snapp-incubator/matrix-on-call-bot/internal/model"
)

func main(cfg config.Config) {
	db := database.WithRetry(
		database.Create,
		cfg.Database.Driver,
		cfg.Database.ConnStr,
		cfg.Database.Options,
	)

	roomRepo := &model.SQLRoomRepo{DB: db}
	shiftRepo := &model.SQLShiftRepo{DB: db}
	followUpRepo := &model.SQLFollowUpRepo{DB: db}

	bot, err := matrix.New(cfg.Matrix.URL, cfg.Matrix.UserID, cfg.Matrix.Token,
		cfg.Matrix.DisplayName, roomRepo, shiftRepo, followUpRepo)
	if err != nil {
		logrus.WithField("error", err.Error()).Error("cannot create bot instance")
	}

	if err := bot.RegisterListeners(); err != nil {
		logrus.WithField("error", err.Error()).Fatalf("couldn't register listeners")
	}

	logrus.Info("bot is started!")

	s := internalhttp.NewServer()
	go s.Run(cfg.Server.Listen)
	logrus.Info("bot is started!")

	bot.Run()
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
