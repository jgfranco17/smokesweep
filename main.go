package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/jgfranco17/dev-tooling-go/logging"
	"github.com/jgfranco17/smokesweep/core"
)

const (
	projectName        = "smokesweep"
	projectDescription = "SmokeSweep: A CLI tool for executing smoke tests on REST API services."
)

var (
	version string = "0.0.0-dev.1"
)

func init() {
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})
}

func main() {
	commandsList := []*cobra.Command{
		core.GetRunCommand(),
		core.GetPingCommand(),
	}
	command := core.NewCommandRegistry(projectName, projectDescription, version)
	command.RegisterCommands(commandsList)

	logger := logging.New(os.Stderr, logrus.ErrorLevel)
	err := command.Execute()
	if err != nil {
		logger.Error(err.Error())
	}
}
