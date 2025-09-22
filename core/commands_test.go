package core

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/jgfranco17/smokesweep/config"
	"github.com/jgfranco17/smokesweep/logging"
	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type CliCommandFunction func() *cobra.Command

type CommandRunner func(cmd *cobra.Command, args []string)

type CliRunResult struct {
	ShellOutput string
	Error       error
}

func createTempDir(t *testing.T) string {
	t.Helper()
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "test-dir-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Return the directory path for use in tests
	return tempDir
}

// Helper function to simulate CLI execution
func ExecuteTestCommand(cmdGetter CliCommandFunction, args ...string) CliRunResult {
	buf := new(bytes.Buffer)
	cmd := cmdGetter()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)

	loggerForTests := logging.New(os.Stderr, logrus.WarnLevel)
	ctx := logging.ApplyToContext(context.Background(), loggerForTests)
	cmd.SetContext(ctx)

	_, err := cmd.ExecuteC()
	return CliRunResult{
		ShellOutput: buf.String(),
		Error:       err,
	}
}

func TestRunCommandHelp(t *testing.T) {
	output := ExecuteTestCommand(GetRunCommand, "--help")
	assert.NoError(t, output.Error)
	assert.Contains(t, output.ShellOutput, "run [flags]", "Did not find usage guide for run command")
}

func TestRunCommandSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	endpoints := []config.Endpoint{
		{Path: "/users", ExpectedStatus: 200},
	}
	mockConfig := config.TestSuite{
		URL:       server.URL,
		Endpoints: endpoints,
	}
	temp := createTempDir(t)
	defer os.RemoveAll(temp)
	mockConfigFilePath := filepath.Join(temp, "config.yaml")
	err := mockConfig.Write(mockConfigFilePath)
	assert.NoError(t, err)

	output := ExecuteTestCommand(GetRunCommand, "-f", mockConfigFilePath)
	assert.NoError(t, output.Error, "Unexpected error while executing run command")
}

func TestRunCommandInvalidConfig(t *testing.T) {
	output := ExecuteTestCommand(GetRunCommand, "-f", "non-existent.yaml")
	assert.ErrorContains(t, output.Error, "no such file or directory")
}

func TestRunCommandFailFast(t *testing.T) {
	flags := []string{"--fail-fast", "-x"}
	for _, flag := range flags {
		t.Run(fmt.Sprintf("Fail fast with %s", flag), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}))
			defer server.Close()

			endpoints := []config.Endpoint{
				{Path: "/some-endpoint", ExpectedStatus: 200},
			}
			mockConfig := config.TestSuite{
				URL:       server.URL,
				Endpoints: endpoints,
			}
			temp := createTempDir(t)
			defer os.RemoveAll(temp)
			writePath := filepath.Join(temp, "config.yaml")
			err := mockConfig.Write(writePath)
			assert.NoError(t, err)

			output := ExecuteTestCommand(GetRunCommand, "-f", writePath, flag)
			assert.ErrorContains(t, output.Error, "expected HTTP 200 but got 500")
		})
	}
}

func TestPingCommandSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	output := ExecuteTestCommand(GetPingCommand, server.URL)
	assert.NoError(t, output.Error, "Unexpected error while executing ping command")
}

func TestPingCommandServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	output := ExecuteTestCommand(GetPingCommand, server.URL, "--timeout", "1s")
	assert.NoError(t, output.Error, "unexpected error while executing ping command")
}

func TestPingCommandUnreachable(t *testing.T) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	output := ExecuteTestCommand(GetPingCommand, server.URL, "--timeout", "1s")
	assert.ErrorContains(t, output.Error, "failed to reach target")
}
