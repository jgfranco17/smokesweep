package core

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/jgfranco17/dev-tooling-go/logging"
	"github.com/jgfranco17/smokesweep/config"
	"github.com/jgfranco17/smokesweep/runner"
)

var (
	failFast bool
)

func GetRunCommand() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run smoke test suite",
		Long:  "Run the smoke tests using the config file provided.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := logging.FromContext(cmd.Context())

			configFilePath := args[0]
			file, err := os.Open(configFilePath)
			if err != nil {
				return fmt.Errorf("error opening config file: %w", err)
			}
			defer file.Close()

			testConfigs, err := config.Load(file)
			if err != nil {
				return fmt.Errorf("error loading config file: %w", err)
			}
			logger.WithFields(
				logrus.Fields{
					"config": configFilePath,
				},
			).Debug("Config file loaded successfully")
			report, err := runner.RunTests(cmd.Context(), testConfigs, failFast)
			if err != nil {
				return fmt.Errorf("error running tests: %w", err)
			}
			if err := report.SummarizeResults(); err != nil {
				return fmt.Errorf("error summarizing test results: %w", err)
			}
			return nil
		},
	}
	runCmd.Flags().BoolVarP(&failFast, "fail-fast", "f", false, "Stop executing tests on the first failure")
	return runCmd
}

func GetPingCommand() *cobra.Command {
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "ping",
		Short: "Ping a target URL",
		Long:  "Check if a target URL is live and responds with a 2xx status code",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := logging.FromContext(cmd.Context())
			target := args[0]
			if err := runner.PingUrl(cmd.Context(), target, timeout); err != nil {
				logger.WithFields(
					logrus.Fields{
						"target": target,
						"error":  err.Error(),
					},
				).Error("Ping failed", err)
				return err
			}
			return nil
		},
	}
	defaultTimeout := 5 * time.Second
	cmd.Flags().DurationVarP(&timeout, "timeout", "t", defaultTimeout, "Timeout duration (in seconds) for the ping request, default is 5s")
	return cmd
}
