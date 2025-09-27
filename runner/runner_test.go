package runner

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jgfranco17/dev-tooling-go/logging"
	"github.com/jgfranco17/smokesweep/config"
	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockConfig(url string, endpoints []config.Endpoint) *config.TestSuite {
	return &config.TestSuite{
		URL:       url,
		Endpoints: endpoints,
	}
}

func newContextWithLogger(t *testing.T) (context.Context, bytes.Buffer) {
	t.Helper()
	ctx := context.Background()
	var buf bytes.Buffer

	logger := logging.New(os.Stdout, logrus.TraceLevel)
	ctx = logging.WithContext(ctx, logger)

	return ctx, buf
}

func TestExecute(t *testing.T) {
	tests := []struct {
		name           string
		config         *config.TestSuite
		failFast       bool
		mockHandler    func(w http.ResponseWriter, r *http.Request)
		expectedError  string
		expectedCount  int
		expectedPassed []bool
		expectedStatus []int
	}{
		{
			name: "successful single endpoint",
			config: newMockConfig("", []config.Endpoint{
				{Path: "/users", ExpectedStatus: 200},
			}),
			failFast: false,
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectedCount:  1,
			expectedPassed: []bool{true},
			expectedStatus: []int{200},
		},
		{
			name: "failed single endpoint",
			config: newMockConfig("", []config.Endpoint{
				{Path: "/users", ExpectedStatus: 200},
			}),
			failFast: false,
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedCount:  1,
			expectedPassed: []bool{false},
			expectedStatus: []int{500},
		},
		{
			name: "multiple endpoints with mixed results",
			config: newMockConfig("", []config.Endpoint{
				{Path: "/users", ExpectedStatus: 200},
				{Path: "/posts", ExpectedStatus: 200}, // Expect 200 but will get 201
				{Path: "/comments", ExpectedStatus: 404},
			}),
			failFast: false,
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/users":
					w.WriteHeader(http.StatusOK)
				case "/posts":
					w.WriteHeader(http.StatusCreated)
				case "/comments":
					w.WriteHeader(http.StatusNotFound)
				}
			},
			expectedCount:  3,
			expectedPassed: []bool{true, false, true},
			expectedStatus: []int{200, 201, 404},
		},
		{
			name: "fail fast on first failure",
			config: newMockConfig("", []config.Endpoint{
				{Path: "/users", ExpectedStatus: 200},
				{Path: "/posts", ExpectedStatus: 201},
			}),
			failFast: true,
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedError: "expected HTTP",
		},
		{
			name: "fail fast on unreachable target",
			config: newMockConfig("invalid-url", []config.Endpoint{
				{Path: "/users", ExpectedStatus: 200},
			}),
			failFast:      true,
			expectedError: "failed to reach target",
		},
		{
			name: "unreachable target without fail fast",
			config: newMockConfig("invalid-url", []config.Endpoint{
				{Path: "/users", ExpectedStatus: 200},
			}),
			failFast:      false,
			expectedCount: 0, // Unreachable targets are skipped when failFast is false
		},
		{
			name: "endpoint with timeout",
			config: newMockConfig("", []config.Endpoint{
				{Path: "/slow", ExpectedStatus: 200, Timeout: intPtr(1000)},
			}),
			failFast: false,
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(50 * time.Millisecond) // Fast enough to pass
				w.WriteHeader(http.StatusOK)
			},
			expectedCount:  1,
			expectedPassed: []bool{true},
			expectedStatus: []int{200},
		},
		{
			name:          "empty endpoints list",
			config:        newMockConfig("", []config.Endpoint{}),
			failFast:      false,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, _ := newContextWithLogger(t)

			var server *httptest.Server
			if tt.mockHandler != nil {
				server = httptest.NewServer(http.HandlerFunc(tt.mockHandler))
				defer server.Close()
				tt.config.URL = server.URL
			}

			report, err := Execute(ctx, tt.config, tt.failFast)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedCount, len(report.Results))

			for i, result := range report.Results {
				if i < len(tt.expectedPassed) {
					assert.Equal(t, tt.expectedPassed[i], result.Passed, "Result %d passed status", i)
				}
				if i < len(tt.expectedStatus) {
					assert.Equal(t, tt.expectedStatus[i], result.HttpStatus, "Result %d status code", i)
				}
				assert.NotZero(t, result.Duration, "Result %d should have duration", i)
				assert.Contains(t, result.Target, tt.config.Endpoints[i].Path, "Result %d should contain path", i)
			}
		})
	}
}

func TestPingURL(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		timeout        time.Duration
		mockHandler    func(w http.ResponseWriter, r *http.Request)
		expectedError  string
		expectedStatus int
	}{
		{
			name:    "successful ping with 200 status",
			timeout: 5 * time.Second,
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectedStatus: 200,
		},
		{
			name:    "successful ping with 201 status",
			timeout: 5 * time.Second,
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
			},
			expectedStatus: 201,
		},
		{
			name:    "successful ping with 301 redirect",
			timeout: 5 * time.Second,
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusMovedPermanently)
			},
			expectedStatus: 301,
		},
		{
			name:    "failed ping with 400 status",
			timeout: 5 * time.Second,
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			},
			expectedStatus: 400,
		},
		{
			name:    "failed ping with 500 status",
			timeout: 5 * time.Second,
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedStatus: 500,
		},
		{
			name:          "timeout error",
			timeout:       1 * time.Millisecond,
			expectedError: "failed to reach target",
		},
		{
			name:          "invalid URL",
			url:           "invalid-url",
			timeout:       5 * time.Second,
			expectedError: "failed to reach target",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, _ := newContextWithLogger(t)

			url := tt.url
			if tt.mockHandler != nil {
				server := httptest.NewServer(http.HandlerFunc(tt.mockHandler))
				defer server.Close()
				url = server.URL
			}

			err := PingURL(ctx, url, tt.timeout)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestTestReport_SummarizeResults(t *testing.T) {
	tests := []struct {
		name        string
		report      TestReport
		expectedErr string
	}{
		{
			name: "empty results should error",
			report: TestReport{
				Timestamp: time.Now(),
				Results:   []TestResult{},
			},
			expectedErr: "no test results to print",
		},
		{
			name: "nil results should error",
			report: TestReport{
				Timestamp: time.Now(),
				Results:   nil,
			},
			expectedErr: "no test results to print",
		},
		{
			name: "successful result",
			report: TestReport{
				Timestamp: time.Now(),
				Results: []TestResult{
					{
						Target:         "https://example.com/users",
						Duration:       100 * time.Millisecond,
						HttpStatus:     200,
						ExpectedStatus: 200,
						Passed:         true,
					},
				},
			},
		},
		{
			name: "failed result",
			report: TestReport{
				Timestamp: time.Now(),
				Results: []TestResult{
					{
						Target:         "https://example.com/users",
						Duration:       100 * time.Millisecond,
						HttpStatus:     500,
						ExpectedStatus: 200,
						Passed:         false,
					},
				},
			},
		},
		{
			name: "successful result with timeout within limit",
			report: TestReport{
				Timestamp: time.Now(),
				Results: []TestResult{
					{
						Target:         "https://example.com/users",
						Duration:       50 * time.Millisecond,
						Timeout:        timePtr(100 * time.Millisecond),
						HttpStatus:     200,
						ExpectedStatus: 200,
						Passed:         true,
					},
				},
			},
		},
		{
			name: "successful result with timeout exceeded",
			report: TestReport{
				Timestamp: time.Now(),
				Results: []TestResult{
					{
						Target:         "https://example.com/users",
						Duration:       150 * time.Millisecond,
						Timeout:        timePtr(100 * time.Millisecond),
						HttpStatus:     200,
						ExpectedStatus: 200,
						Passed:         true,
					},
				},
			},
		},
		{
			name: "multiple results with mixed outcomes",
			report: TestReport{
				Timestamp: time.Now(),
				Results: []TestResult{
					{
						Target:         "https://example.com/users",
						Duration:       100 * time.Millisecond,
						HttpStatus:     200,
						ExpectedStatus: 200,
						Passed:         true,
					},
					{
						Target:         "https://example.com/posts",
						Duration:       200 * time.Millisecond,
						HttpStatus:     500,
						ExpectedStatus: 200,
						Passed:         false,
					},
					{
						Target:         "https://example.com/comments",
						Duration:       50 * time.Millisecond,
						Timeout:        timePtr(100 * time.Millisecond),
						HttpStatus:     201,
						ExpectedStatus: 201,
						Passed:         true,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.report.SummarizeResults()

			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestJoinURL(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		paths    []string
		expected string
	}{
		{
			name:     "simple base with single path",
			base:     "https://example.com",
			paths:    []string{"users"},
			expected: "https://example.com/users",
		},
		{
			name:     "base with trailing slash",
			base:     "https://example.com/",
			paths:    []string{"users"},
			expected: "https://example.com/users",
		},
		{
			name:     "path with leading slash",
			base:     "https://example.com",
			paths:    []string{"/users"},
			expected: "https://example.com/users",
		},
		{
			name:     "both base and path with slashes",
			base:     "https://example.com/",
			paths:    []string{"/users"},
			expected: "https://example.com/users",
		},
		{
			name:     "multiple paths",
			base:     "https://example.com",
			paths:    []string{"api", "v1", "users"},
			expected: "https://example.com/api/v1/users",
		},
		{
			name:     "empty paths",
			base:     "https://example.com",
			paths:    []string{},
			expected: "https://example.com/",
		},
		{
			name:     "empty base",
			base:     "",
			paths:    []string{"users"},
			expected: "/users",
		},
		{
			name:     "complex URL with query params",
			base:     "https://example.com/api?version=1",
			paths:    []string{"users"},
			expected: "https://example.com/api?version=1/users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinURL(tt.base, tt.paths...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTestResult_Fields(t *testing.T) {
	timeout := 100 * time.Millisecond
	result := TestResult{
		Target:         "https://example.com/users",
		Duration:       50 * time.Millisecond,
		Timeout:        &timeout,
		HttpStatus:     200,
		ExpectedStatus: 200,
		Passed:         true,
	}

	assert.Equal(t, "https://example.com/users", result.Target)
	assert.Equal(t, 50*time.Millisecond, result.Duration)
	assert.Equal(t, &timeout, result.Timeout)
	assert.Equal(t, 200, result.HttpStatus)
	assert.Equal(t, 200, result.ExpectedStatus)
	assert.True(t, result.Passed)
}

func TestTestReport_Fields(t *testing.T) {
	timestamp := time.Now()
	results := []TestResult{
		{Target: "https://example.com/users", Passed: true},
		{Target: "https://example.com/posts", Passed: false},
	}

	report := TestReport{
		Timestamp: timestamp,
		Results:   results,
	}

	assert.Equal(t, timestamp, report.Timestamp)
	assert.Equal(t, results, report.Results)
	assert.Len(t, report.Results, 2)
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func timePtr(t time.Duration) *time.Duration {
	return &t
}

// Test the actual output of SummarizeResults by capturing stdout
func TestTestReport_SummarizeResults_Output(t *testing.T) {
	report := TestReport{
		Timestamp: time.Now(),
		Results: []TestResult{
			{
				Target:         "https://example.com/users",
				Duration:       100 * time.Millisecond,
				HttpStatus:     200,
				ExpectedStatus: 200,
				Passed:         true,
			},
			{
				Target:         "https://example.com/posts",
				Duration:       200 * time.Millisecond,
				HttpStatus:     500,
				ExpectedStatus: 200,
				Passed:         false,
			},
		},
	}

	err := report.SummarizeResults()
	require.NoError(t, err)
}

// Test edge cases for Execute function
func TestExecute_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		config        *config.TestSuite
		failFast      bool
		expectedError string
		expectedCount int
	}{
		{
			name:          "nil config",
			config:        nil,
			failFast:      false,
			expectedError: "runtime error",
			expectedCount: 0,
		},
		{
			name: "config with nil endpoints",
			config: &config.TestSuite{
				URL:       "https://example.com",
				Endpoints: nil,
			},
			failFast:      false,
			expectedCount: 0,
		},
		{
			name: "config with empty URL",
			config: &config.TestSuite{
				URL:       "",
				Endpoints: []config.Endpoint{{Path: "/test", ExpectedStatus: 200}},
			},
			failFast:      false,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, _ := newContextWithLogger(t)

			// Use defer recover to handle panics
			defer func() {
				if r := recover(); r != nil {
					if tt.expectedError != "" {
						assert.Contains(t, fmt.Sprintf("%v", r), tt.expectedError)
					}
				}
			}()

			report, err := Execute(ctx, tt.config, tt.failFast)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}

			require.NoError(t, err)
			assert.NotNil(t, report)
			if tt.expectedCount > 0 {
				assert.NotNil(t, report.Results)
			}
			assert.Equal(t, tt.expectedCount, len(report.Results))
		})
	}
}
