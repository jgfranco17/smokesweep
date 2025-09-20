package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// TestSuite represents the top-level structure of the smoke test suite
// configuration file.
type TestSuite struct {
	URL       string     `yaml:"url"`
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

type Endpoint struct {
	Path           string `yaml:"path"`
	ExpectedStatus int    `yaml:"expected-status"`
	Timeout        *int   `yaml:"timeout-ms,omitempty"`
}
