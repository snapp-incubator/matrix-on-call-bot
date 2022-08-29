package cmd

import (
	"github.com/spf13/cobra"

	"github.com/snapp-incubator/matrix-on-call-bot/internal/cmd/migrate"
	"github.com/snapp-incubator/matrix-on-call-bot/internal/cmd/server"
	"github.com/snapp-incubator/matrix-on-call-bot/internal/config"
)

// NewRootCommand creates a new matrix-matrix-on-call-bot root command.
func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use: "matrix-on-call-bot",
	}

	cfg := config.Init()

	server.Register(root, cfg)
	migrate.Register(root, cfg)

	return root
}
