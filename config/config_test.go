package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// Helper types and functions
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, assert.AnError
}

func intPtr(i int) *int {
	return &i
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name            string
		config          string
		expectedError   string
		expectedURL     string
		expectedCount   int
		expectedPaths   []string
		expectedStatus  []int
		expectedTimeout []*int
	}{
		{
			name: "valid basic config",
			config: `---
url: "https://example.com"
endpoints:
  - path: "/"
    expected-status: 200`,
			expectedURL:    "https://example.com",
			expectedCount:  1,
			expectedPaths:  []string{"/"},
			expectedStatus: []int{200},
		},
		{
			name: "valid config with multiple endpoints",
			config: `---
url: "https://api.example.com"
endpoints:
  - path: "/users"
    expected-status: 200
  - path: "/posts"
    expected-status: 201
  - path: "/comments"
    expected-status: 404`,
			expectedURL:    "https://api.example.com",
			expectedCount:  3,
			expectedPaths:  []string{"/users", "/posts", "/comments"},
			expectedStatus: []int{200, 201, 404},
		},
		{
			name: "valid config with timeouts",
			config: `---
url: "https://example.com"
endpoints:
  - path: "/fast"
    expected-status: 200
    timeout-ms: 100
  - path: "/slow"
    expected-status: 200
    timeout-ms: 5000`,
			expectedURL:     "https://example.com",
			expectedCount:   2,
			expectedPaths:   []string{"/fast", "/slow"},
			expectedStatus:  []int{200, 200},
			expectedTimeout: []*int{intPtr(100), intPtr(5000)},
		},
		{
			name: "valid config with mixed timeouts",
			config: `---
url: "https://example.com"
endpoints:
  - path: "/no-timeout"
    expected-status: 200
  - path: "/with-timeout"
    expected-status: 201
    timeout-ms: 2000`,
			expectedURL:     "https://example.com",
			expectedCount:   2,
			expectedPaths:   []string{"/no-timeout", "/with-timeout"},
			expectedStatus:  []int{200, 201},
			expectedTimeout: []*int{nil, intPtr(2000)},
		},
		{
			name: "empty config",
			config: `---
url: ""
endpoints: []`,
			expectedURL:   "",
			expectedCount: 0,
		},
		{
			name: "config with no endpoints",
			config: `---
url: "https://example.com"
endpoints: []`,
			expectedURL:   "https://example.com",
			expectedCount: 0,
		},
		{
			name: "config with empty endpoints",
			config: `---
url: "https://example.com"
endpoints:`,
			expectedURL:   "https://example.com",
			expectedCount: 0,
		},
		{
			name: "invalid YAML syntax",
			config: `---
url: "https://example.com"
endpoints:
  - path: "/"
    expected-status: 200
  invalid: yaml: [`,
			expectedError: "error parsing config file",
		},
		{
			name: "invalid endpoint structure",
			config: `---
url: "https://example.com"
endpoints:
  - path: {}
    expected-status: 200`,
			expectedError: "error parsing config file",
		},
		{
			name: "invalid status code type",
			config: `---
url: "https://example.com"
endpoints:
  - path: "/"
    expected-status: "200"`,
			expectedError: "error parsing config file",
		},
		{
			name: "invalid timeout type",
			config: `---
url: "https://example.com"
endpoints:
  - path: "/"
    expected-status: 200
    timeout-ms: "1000"`,
			expectedError: "error parsing config file",
		},
		{
			name: "missing required fields",
			config: `---
endpoints:
  - path: "/"
    expected-status: 200`,
			expectedURL:   "",
			expectedCount: 1,
			expectedPaths: []string{"/"},
		},
		{
			name: "extra fields ignored",
			config: `---
url: "https://example.com"
extra_field: "ignored"
endpoints:
  - path: "/"
    expected-status: 200
    extra_endpoint_field: "ignored"`,
			expectedURL:   "https://example.com",
			expectedCount: 1,
			expectedPaths: []string{"/"},
		},
		{
			name: "config with special characters in URL",
			config: `---
url: "https://api.example.com:8080/v1"
endpoints:
  - path: "/test-endpoint"
    expected-status: 200`,
			expectedURL:   "https://api.example.com:8080/v1",
			expectedCount: 1,
			expectedPaths: []string{"/test-endpoint"},
		},
		{
			name: "config with special characters in paths",
			config: `---
url: "https://example.com"
endpoints:
  - path: "/api/v1/users"
    expected-status: 200
  - path: "/api/v2/posts?limit=10"
    expected-status: 200`,
			expectedURL:   "https://example.com",
			expectedCount: 2,
			expectedPaths: []string{"/api/v1/users", "/api/v2/posts?limit=10"},
		},
		{
			name: "config with zero timeout",
			config: `---
url: "https://example.com"
endpoints:
  - path: "/"
    expected-status: 200
    timeout-ms: 0`,
			expectedURL:     "https://example.com",
			expectedCount:   1,
			expectedPaths:   []string{"/"},
			expectedStatus:  []int{200},
			expectedTimeout: []*int{intPtr(0)},
		},
		{
			name: "config with negative timeout",
			config: `---
url: "https://example.com"
endpoints:
  - path: "/"
    expected-status: 200
    timeout-ms: -100`,
			expectedURL:     "https://example.com",
			expectedCount:   1,
			expectedPaths:   []string{"/"},
			expectedStatus:  []int{200},
			expectedTimeout: []*int{intPtr(-100)},
		},
		{
			name: "config with very large timeout",
			config: `---
url: "https://example.com"
endpoints:
  - path: "/"
    expected-status: 200
    timeout-ms: 999999`,
			expectedURL:     "https://example.com",
			expectedCount:   1,
			expectedPaths:   []string{"/"},
			expectedStatus:  []int{200},
			expectedTimeout: []*int{intPtr(999999)},
		},
		{
			name: "config with various HTTP status codes",
			config: `---
url: "https://example.com"
endpoints:
  - path: "/ok"
    expected-status: 200
  - path: "/created"
    expected-status: 201
  - path: "/accepted"
    expected-status: 202
  - path: "/no-content"
    expected-status: 204
  - path: "/redirect"
    expected-status: 301
  - path: "/not-found"
    expected-status: 404
  - path: "/server-error"
    expected-status: 500`,
			expectedURL:    "https://example.com",
			expectedCount:  7,
			expectedPaths:  []string{"/ok", "/created", "/accepted", "/no-content", "/redirect", "/not-found", "/server-error"},
			expectedStatus: []int{200, 201, 202, 204, 301, 404, 500},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configReader := strings.NewReader(tt.config)
			config, err := Load(configReader)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, config)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, config)
			assert.Equal(t, tt.expectedURL, config.URL)
			assert.Equal(t, tt.expectedCount, len(config.Endpoints))

			for i, endpoint := range config.Endpoints {
				if i < len(tt.expectedPaths) {
					assert.Equal(t, tt.expectedPaths[i], endpoint.Path, "Path mismatch for endpoint %d", i)
				}
				if i < len(tt.expectedStatus) {
					assert.Equal(t, tt.expectedStatus[i], endpoint.ExpectedStatus, "Status mismatch for endpoint %d", i)
				}
				if i < len(tt.expectedTimeout) {
					if tt.expectedTimeout[i] == nil {
						assert.Nil(t, endpoint.Timeout, "Expected nil timeout for endpoint %d", i)
					} else {
						require.NotNil(t, endpoint.Timeout, "Expected non-nil timeout for endpoint %d", i)
						assert.Equal(t, *tt.expectedTimeout[i], *endpoint.Timeout, "Timeout mismatch for endpoint %d", i)
					}
				}
			}
		})
	}
}

func TestLoad_ReaderErrors(t *testing.T) {
	tests := []struct {
		name        string
		reader      io.Reader
		expectedErr string
	}{
		{
			name:        "nil reader",
			reader:      nil,
			expectedErr: "runtime error",
		},
		{
			name:        "reader that returns error",
			reader:      &errorReader{},
			expectedErr: "general error for testing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use defer recover to handle panics
			defer func() {
				if r := recover(); r != nil {
					if tt.expectedErr != "" {
						assert.Contains(t, fmt.Sprintf("%v", r), tt.expectedErr)
					}
				}
			}()

			config, err := Load(tt.reader)

			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				assert.Nil(t, config)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestTestSuite_Write(t *testing.T) {
	tests := []struct {
		name        string
		config      *TestSuite
		expectedErr string
		validate    func(t *testing.T, filePath string)
	}{
		{
			name: "valid config write",
			config: &TestSuite{
				URL: "https://example.com",
				Endpoints: []Endpoint{
					{Path: "/users", ExpectedStatus: 200},
					{Path: "/posts", ExpectedStatus: 201, Timeout: intPtr(1000)},
				},
			},
			validate: func(t *testing.T, filePath string) {
				// Read the file back and verify contents
				data, err := os.ReadFile(filePath)
				require.NoError(t, err)
				assert.Contains(t, string(data), "https://example.com")
				assert.Contains(t, string(data), "/users")
				assert.Contains(t, string(data), "/posts")
				assert.Contains(t, string(data), "expected-status: 200")
				assert.Contains(t, string(data), "expected-status: 201")
				assert.Contains(t, string(data), "timeout-ms: 1000")
			},
		},
		{
			name: "empty config write",
			config: &TestSuite{
				URL:       "",
				Endpoints: []Endpoint{},
			},
			validate: func(t *testing.T, filePath string) {
				data, err := os.ReadFile(filePath)
				require.NoError(t, err)
				assert.Contains(t, string(data), "url: \"\"")
				assert.Contains(t, string(data), "endpoints: []")
			},
		},
		{
			name: "config with nil endpoints",
			config: &TestSuite{
				URL:       "https://example.com",
				Endpoints: nil,
			},
			validate: func(t *testing.T, filePath string) {
				data, err := os.ReadFile(filePath)
				require.NoError(t, err)
				assert.Contains(t, string(data), "https://example.com")
			},
		},
		{
			name: "config with special characters",
			config: &TestSuite{
				URL: "https://api.example.com:8080/v1",
				Endpoints: []Endpoint{
					{Path: "/api/v1/users?limit=10&offset=0", ExpectedStatus: 200},
					{Path: "/api/v1/posts/search?q=test&sort=date", ExpectedStatus: 200, Timeout: intPtr(5000)},
				},
			},
			validate: func(t *testing.T, filePath string) {
				data, err := os.ReadFile(filePath)
				require.NoError(t, err)
				assert.Contains(t, string(data), "https://api.example.com:8080/v1")
				assert.Contains(t, string(data), "/api/v1/users?limit=10&offset=0")
				assert.Contains(t, string(data), "/api/v1/posts/search?q=test&sort=date")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file
			tempDir := t.TempDir()
			filePath := filepath.Join(tempDir, "test-config.yaml")

			err := tt.config.Write(filePath)

			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				return
			}

			require.NoError(t, err)

			_, err = os.Stat(filePath)
			require.NoError(t, err)

			if tt.validate != nil {
				tt.validate(t, filePath)
			}

			data, err := os.ReadFile(filePath)
			require.NoError(t, err)

			var parsedConfig TestSuite
			err = yaml.Unmarshal(data, &parsedConfig)
			require.NoError(t, err)
			assert.Equal(t, tt.config.URL, parsedConfig.URL)
			assert.Equal(t, len(tt.config.Endpoints), len(parsedConfig.Endpoints))
		})
	}
}

func TestTestSuite_Write_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		expectedErr string
	}{
		{
			name:        "invalid file path",
			filePath:    "/invalid/path/that/does/not/exist/config.yaml",
			expectedErr: "no such file or directory",
		},
		{
			name:        "directory instead of file",
			filePath:    "/tmp",
			expectedErr: "is a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &TestSuite{
				URL: "https://example.com",
				Endpoints: []Endpoint{
					{Path: "/test", ExpectedStatus: 200},
				},
			}

			err := config.Write(tt.filePath)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestEndpoint_Fields(t *testing.T) {
	timeout := 1000
	endpoint := Endpoint{
		Path:           "/test",
		ExpectedStatus: 200,
		Timeout:        &timeout,
	}

	assert.Equal(t, "/test", endpoint.Path)
	assert.Equal(t, 200, endpoint.ExpectedStatus)
	assert.Equal(t, &timeout, endpoint.Timeout)
	assert.Equal(t, 1000, *endpoint.Timeout)
}

func TestTestSuite_Fields(t *testing.T) {
	endpoints := []Endpoint{
		{Path: "/users", ExpectedStatus: 200},
		{Path: "/posts", ExpectedStatus: 201},
	}
	config := TestSuite{
		URL:       "https://example.com",
		Endpoints: endpoints,
	}

	assert.Equal(t, "https://example.com", config.URL)
	assert.Equal(t, endpoints, config.Endpoints)
	assert.Len(t, config.Endpoints, 2)
}

func TestLoad_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		config      string
		expectedErr string
		validate    func(t *testing.T, config *TestSuite)
	}{
		{
			name: "YAML with comments",
			config: `---
# This is a comment
url: "https://example.com"  # Inline comment
endpoints:
  - path: "/users"  # Another comment
    expected-status: 200
  - path: "/posts"
    expected-status: 201
    timeout-ms: 1000  # Timeout comment`,
			validate: func(t *testing.T, config *TestSuite) {
				assert.Equal(t, "https://example.com", config.URL)
				assert.Len(t, config.Endpoints, 2)
				assert.Equal(t, "/users", config.Endpoints[0].Path)
				assert.Equal(t, 200, config.Endpoints[0].ExpectedStatus)
				assert.Equal(t, "/posts", config.Endpoints[1].Path)
				assert.Equal(t, 201, config.Endpoints[1].ExpectedStatus)
				require.NotNil(t, config.Endpoints[1].Timeout)
				assert.Equal(t, 1000, *config.Endpoints[1].Timeout)
			},
		},
		{
			name: "YAML with anchors and aliases",
			config: `---
url: "https://example.com"
endpoints:
  - &default_endpoint
    path: "/default"
    expected-status: 200
    timeout-ms: 1000
  - <<: *default_endpoint
    path: "/override"
  - <<: *default_endpoint
    path: "/another"
    expected-status: 201`,
			validate: func(t *testing.T, config *TestSuite) {
				// Note: YAML anchors/aliases might not work as expected in this simple case
				// but the test ensures the YAML is valid
				assert.Equal(t, "https://example.com", config.URL)
				assert.Len(t, config.Endpoints, 3)
			},
		},
		{
			name: "YAML with multi-line strings",
			config: `---
url: "https://example.com"
endpoints:
  - path: |
      /very/long/path/that/spans
      multiple/lines
    expected-status: 200`,
			validate: func(t *testing.T, config *TestSuite) {
				assert.Equal(t, "https://example.com", config.URL)
				assert.Len(t, config.Endpoints, 1)
				assert.Contains(t, config.Endpoints[0].Path, "/very/long/path")
			},
		},
		{
			name: "YAML with null values",
			config: `---
url: "https://example.com"
endpoints:
  - path: "/test"
    expected-status: 200
    timeout-ms: null`,
			validate: func(t *testing.T, config *TestSuite) {
				assert.Equal(t, "https://example.com", config.URL)
				assert.Len(t, config.Endpoints, 1)
				assert.Equal(t, "/test", config.Endpoints[0].Path)
				assert.Equal(t, 200, config.Endpoints[0].ExpectedStatus)
				assert.Nil(t, config.Endpoints[0].Timeout)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configReader := strings.NewReader(tt.config)
			config, err := Load(configReader)

			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, config)

			if tt.validate != nil {
				tt.validate(t, config)
			}
		})
	}
}

// Test the round-trip functionality (write then read)
func TestConfig_RoundTrip(t *testing.T) {
	originalConfig := &TestSuite{
		URL: "https://api.example.com:8080/v1",
		Endpoints: []Endpoint{
			{Path: "/users", ExpectedStatus: 200},
			{Path: "/posts", ExpectedStatus: 201, Timeout: intPtr(1000)},
			{Path: "/comments", ExpectedStatus: 404, Timeout: intPtr(500)},
		},
	}

	// Write to temporary file
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "roundtrip-test.yaml")

	err := originalConfig.Write(filePath)
	require.NoError(t, err)

	// Read it back
	file, err := os.Open(filePath)
	require.NoError(t, err)
	defer file.Close()

	loadedConfig, err := Load(file)
	require.NoError(t, err)

	// Compare the configurations
	assert.Equal(t, originalConfig.URL, loadedConfig.URL)
	assert.Equal(t, len(originalConfig.Endpoints), len(loadedConfig.Endpoints))

	for i, originalEndpoint := range originalConfig.Endpoints {
		loadedEndpoint := loadedConfig.Endpoints[i]
		assert.Equal(t, originalEndpoint.Path, loadedEndpoint.Path)
		assert.Equal(t, originalEndpoint.ExpectedStatus, loadedEndpoint.ExpectedStatus)

		if originalEndpoint.Timeout == nil {
			assert.Nil(t, loadedEndpoint.Timeout)
		} else {
			require.NotNil(t, loadedEndpoint.Timeout)
			assert.Equal(t, *originalEndpoint.Timeout, *loadedEndpoint.Timeout)
		}
	}
}
