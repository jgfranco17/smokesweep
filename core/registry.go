package core

import (
	"context"

	"github.com/jgfranco17/smokesweep/logging"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type CommandRegistry struct {
	rootCmd   *cobra.Command
	verbosity int
}

// NewCommandRegistry creates a new instance of CommandRegistry
func NewCommandRegistry(name string, description string, version string) *CommandRegistry {
	root := &cobra.Command{
		Use:     name,
		Version: version,
		Short:   description,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			verbosity, _ := cmd.Flags().GetCount("verbose")
			var level logrus.Level
			switch verbosity {
			case 1:
				level = logrus.InfoLevel
			case 2:
				level = logrus.DebugLevel
			case 3:
				level = logrus.TraceLevel
			default:
				level = logrus.WarnLevel
			}
			logger := logging.New(cmd.ErrOrStderr(), level)
			ctx := logging.ApplyToContext(context.TODO(), logger)
			cmd.SetContext(ctx)
		},
	}
	newRegistry := &CommandRegistry{
		rootCmd: root,
	}
	root.PersistentFlags().CountVarP(&newRegistry.verbosity, "verbose", "v", "Increase verbosity (-v or -vv)")
	root.Flags().BoolP("version", "V", false, "Print the version number of SmokeSweep")
	return newRegistry
}

// RegisterCommand registers a new command with the CommandRegistry
func (cr *CommandRegistry) RegisterCommands(commands []*cobra.Command) {
	for _, cmd := range commands {
		cr.rootCmd.AddCommand(cmd)
	}
}

// Execute executes the root command
func (cr *CommandRegistry) Execute() error {
	ctx := context.TODO()
	return cr.rootCmd.ExecuteContext(ctx)
}
