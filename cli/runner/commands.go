package runner

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"cli/config"
	"cli/outputs"
)

func GetRunCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Run smoke test suite",
		Long:  "Run the smoke tests using the config file provided.",
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
			outputs.PrintStandardMessage("TEST SUITE", "Running smoke test suite...")
			log.Debugf("Target: %s", testConfigs.URL)
			return nil
		},
	}
}
