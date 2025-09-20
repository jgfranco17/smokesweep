// Package config provides functionality for loading and parsing the test suite
// configuration file.
package config

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// Load loads the test suite configuration from the provided reader.
func Load(reader io.Reader) (*TestSuite, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	var config TestSuite
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}
	return &config, nil
}
