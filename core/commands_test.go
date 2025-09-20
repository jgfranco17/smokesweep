package core

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/jgfranco17/smokesweep/config"

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
	mockConfig := config.TestConfig{
		URL:       server.URL,
		Endpoints: endpoints,
	}
	temp := createTempDir(t)
	defer os.RemoveAll(temp)
	writePath := filepath.Join(temp, "config.yaml")
	err := mockConfig.Write(writePath)
	assert.NoError(t, err)

	output := ExecuteTestCommand(GetRunCommand, writePath)
	assert.NoError(t, output.Error, "Unexpected error while executing run command")
}

func TestRunCommandInvalidConfig(t *testing.T) {
	output := ExecuteTestCommand(GetRunCommand, "non-existent.yaml")
	assert.ErrorContains(t, output.Error, "Error loading config file: open non-existent.yaml: no such file or directory")
}

func TestRunCommandFailFast(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	endpoints := []config.Endpoint{
		{Path: "/some-endpoint", ExpectedStatus: 200},
	}
	mockConfig := config.TestConfig{
		URL:       server.URL,
		Endpoints: endpoints,
	}
	temp := createTempDir(t)
	defer os.RemoveAll(temp)
	writePath := filepath.Join(temp, "config.yaml")
	err := mockConfig.Write(writePath)
	assert.NoError(t, err)

	output := ExecuteTestCommand(GetRunCommand, writePath, "--fail-fast")
	assert.ErrorContains(t, output.Error, "expected HTTP 200 but got 500")
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

	output := ExecuteTestCommand(GetPingCommand, server.URL, "--timeout", "1")
	assert.NoError(t, output.Error, "Unexpected error while executing ping command")
}

func TestPingCommandUnreachable(t *testing.T) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	output := ExecuteTestCommand(GetPingCommand, server.URL, "--timeout", "1")
	assert.ErrorContains(t, output.Error, "Failed to reach target")
}
