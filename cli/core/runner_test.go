package core

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"cli/config"

	"github.com/stretchr/testify/assert"
)

type TestEntry struct {
	name         string
	endpoints    []config.Endpoint
	mockResponse func(w http.ResponseWriter, r *http.Request)
	shouldError  bool
}

func newMockConfig(url string, endpoints []config.Endpoint) *config.TestConfig {
	return &config.TestConfig{
		URL:       url,
		Endpoints: endpoints,
	}
}

func TestRunTestsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	endpoints := []config.Endpoint{
		{Path: "/users", ExpectedStatus: 200},
	}
	mockConfig := newMockConfig(server.URL, endpoints)
	results, err := RunTests(mockConfig, false)
	assert.NoError(t, err)
	assert.Equal(t, len(endpoints), len(results))
	for i, result := range results {
		assert.Contains(t, result.Target, endpoints[i].Path)
		assert.True(t, result.Passed)
	}
}

func TestRunTestsFailedCall(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	endpoints := []config.Endpoint{
		{Path: "/users", ExpectedStatus: 200},
	}
	mockConfig := newMockConfig(server.URL, endpoints)
	results, err := RunTests(mockConfig, false)
	assert.NoError(t, err)
	assert.Equal(t, len(endpoints), len(results))
	for i, result := range results {
		assert.Contains(t, result.Target, endpoints[i].Path)
		assert.False(t, result.Passed)
	}
}

func TestRunTestsFailFast(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	endpoints := []config.Endpoint{
		{Path: "/users", ExpectedStatus: 200},
	}
	mockConfig := newMockConfig(server.URL, endpoints)
	results, err := RunTests(mockConfig, true)
	assert.Error(t, err)
	assert.Nil(t, results)
}

func TestRunTestsUnreachable(t *testing.T) {
	endpoints := []config.Endpoint{
		{Path: "/users", ExpectedStatus: http.StatusOK},
	}
	mockConfig := newMockConfig("my-server", endpoints)
	results, err := RunTests(mockConfig, false)
	assert.NoError(t, err)
	for i, result := range results {
		assert.Contains(t, result.Target, endpoints[i].Path)
		assert.False(t, result.Passed)
	}
}

func TestRunTestsUnreachableFailFast(t *testing.T) {
	endpoints := []config.Endpoint{
		{Path: "/users", ExpectedStatus: http.StatusOK},
	}
	mockConfig := newMockConfig("my-server", endpoints)
	results, err := RunTests(mockConfig, true)
	assert.ErrorContains(t, err, "Failed to reach target my-server/users")
	assert.Nil(t, results)
}
