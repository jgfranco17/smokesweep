package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type TestConfig struct {
	URL       string     `yaml:"url"`
	Endpoints []Endpoint `yaml:"endpoints"`
}

func (tc *TestConfig) Write(filePath string) error {
	data, err := yaml.Marshal(tc)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

type Endpoint struct {
	Path            string `yaml:"path"`
	ExpectedStatus  int    `yaml:"expected-status"`
	MaxResponseTime *int   `yaml:"max-response-time-ms,omitempty"`
}
