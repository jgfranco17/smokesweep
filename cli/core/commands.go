package core

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"cli/config"
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
			configFile := args[0]
			testConfigs, err := config.LoadTestSuiteConfig(configFile)
			if err != nil {
				return fmt.Errorf("Error loading config file: %w", err)
			}
			log.Debugf("Using config file: %s", configFile)
			results, err := RunTests(testConfigs, failFast)
			if err != nil {
				return fmt.Errorf("Error running tests: %w", err)
			}
			SummarizeResults(results)
			return nil
		},
	}
	runCmd.Flags().BoolVarP(&failFast, "fail-fast", "f", false, "Stop executing tests on the first failure")
	return runCmd
}
