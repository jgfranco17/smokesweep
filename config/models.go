package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// TestSuite represents the top-level structure of the smoke test
// suite configuration file.
type TestSuite struct {
	// URL is the base URL of the target application.
	URL string `yaml:"url"`

	// Endpoints is the list of endpoints to test.
	Endpoints []Endpoint `yaml:"endpoints"`
}

// Write writes the test suite configuration to a file.
func (tc *TestSuite) Write(filePath string) error {
	data, err := yaml.Marshal(tc)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

// Endpoint represents a single endpoint to test.
type Endpoint struct {
	// Path is the path of the endpoint to test.
	Path string `yaml:"path"`

	// ExpectedStatus is the expected HTTP status code of the response.
	ExpectedStatus int `yaml:"expected-status"`

	// Timeout is the timeout for the test.
	Timeout *int `yaml:"timeout-ms,omitempty"`
}
