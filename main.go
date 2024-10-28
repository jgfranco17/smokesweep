package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"cli/core"
	"cli/logging"
)

const (
	projectName        = "smokesweep"
	projectDescription = "SmokeSweep: A CLI tool for executing smoke tests on REST API services."
)

var (
	version string = "0.0.0-dev.1"
)

func init() {
	log.SetReportCaller(true)
	log.SetFormatter(&logging.CustomFormatter{})
}

func main() {
	commandsList := []*cobra.Command{
		core.GetVersionCommand(version),
		core.GetRunCommand(),
	}
	command := core.NewCommandRegistry(projectName, projectDescription, version)
	command.RegisterCommands(commandsList)

	err := command.Execute()
	if err != nil {
		log.Error(err.Error())
	}
}
