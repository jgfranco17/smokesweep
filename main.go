package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"cli/logging"
	"cli/runner"
)

const (
	projectName        string = "smokesweep"
	projectDescription string = "Smoke test CLI utility for testing REST API services.\nDeveloped by Joaquin Franco."
)

func init() {
	log.SetReportCaller(true)
	log.SetFormatter(&logging.CustomFormatter{})
}

func main() {
	commandsList := []*cobra.Command{
		runner.GetRunCommand(),
	}
	command := runner.NewCommandRegistry(projectName, projectDescription)
	command.RegisterCommands(commandsList)

	err := command.Execute()
	if err != nil {
		log.Error(err.Error())
	}
}
