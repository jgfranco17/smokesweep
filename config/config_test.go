package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	configValid string = `---
url: "https://example.com"
endpoints:
  - path: "/"
    expected-status: 200
  - path: "/info"
    expected-status: 200
    timeout-ms: 500`
	configInvalidDomain string = `---
	url: "nfs:///example/path"
endpoints:
  - path: "/"
    expected-status: 200`
	configInvalidEndpoint string = `---
	url: "https://example.com"
endpoints:
  - path: {}
    expected-status: 200`
	configInvalidCode string = `---
	url: "https://example.com"
endpoints:
  - path: "/"
    expected-status: "200"`
)

func TestLoad_Valid(t *testing.T) {
	configReader := strings.NewReader(configValid)
	definition, err := Load(configReader)
	assert.NoError(t, err)
	assert.Contains(t, definition.URL, "example.com", "Did not find expected URL in config")
}

func TestLoad_Invalid(t *testing.T) {
	testcases := []struct {
		name   string
		config string
	}{
		{
			name:   "Invalid URL",
			config: configInvalidDomain,
		},
		{
			name:   "Invalid endpoint",
			config: configInvalidEndpoint,
		},
		{
			name:   "Invalid status code",
			config: configInvalidCode,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			configReader := strings.NewReader(tc.config)
			definition, err := Load(configReader)
			assert.Nil(t, definition)
			require.Error(t, err)
			assert.ErrorContains(t, err, "error parsing config")
		})
	}
}
