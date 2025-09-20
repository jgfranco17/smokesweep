package core

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/jgfranco17/smokesweep/config"
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
			if len(args) != 1 {
				return fmt.Errorf("Not enough arguments, expected 1 but got %d", len(args))
			}

			configFilePath := args[0]
			file, err := os.Open(configFilePath)
			if err != nil {
				return fmt.Errorf("Error opening config file: %w", err)
			}
			defer file.Close()

			testConfigs, err := config.Load(file)
			if err != nil {
				return fmt.Errorf("Error loading config file: %w", err)
			}
			log.Debugf("Using config file: %s", configFilePath)
			report, err := RunTests(testConfigs, failFast)
			if err != nil {
				return fmt.Errorf("Error running tests: %w", err)
			}
			err = report.SummarizeResults()
			if err != nil {
				return fmt.Errorf("Error summarizing test results: %w", err)
			}
			return nil
		},
	}
	runCmd.Flags().BoolVarP(&failFast, "fail-fast", "f", false, "Stop executing tests on the first failure")
	return runCmd
}

func GetPingCommand() *cobra.Command {
	var timeout int
	cmd := &cobra.Command{
		Use:   "ping",
		Short: "Ping a target URL",
		Long:  "Check if a target URL is live and responds with a 2xx status code",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("Not enough arguments, expected 1 but got %d", len(args))
			}
			target := args[0]
			if err := PingUrl(target, timeout); err != nil {
				log.Errorf("Ping failed: %v", err)
				return err
			}
			return nil
		},
	}
	cmd.Flags().IntVarP(&timeout, "timeout", "t", 5, "Timeout duration (in seconds) for the ping request, default is 5s")
	return cmd
}
