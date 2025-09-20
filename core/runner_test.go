package core

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jgfranco17/smokesweep/config"

	"github.com/stretchr/testify/assert"
)

type TestEntry struct {
	name         string
	endpoints    []config.Endpoint
	mockResponse func(w http.ResponseWriter, r *http.Request)
	shouldError  bool
}

func newMockConfig(url string, endpoints []config.Endpoint) *config.TestSuite {
	return &config.TestSuite{
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
	report, err := RunTests(mockConfig, false)
	assert.NoError(t, err)
	assert.Equal(t, len(endpoints), len(report.Results))
	for i, result := range report.Results {
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
	report, err := RunTests(mockConfig, false)
	assert.NoError(t, err)
	assert.Equal(t, len(endpoints), len(report.Results))
	for i, result := range report.Results {
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
	report, err := RunTests(mockConfig, true)
	assert.ErrorContains(t, err, "expected HTTP 200 but got 500")
	assert.Nil(t, report.Results)
}

func TestRunTestsUnreachable(t *testing.T) {
	endpoints := []config.Endpoint{
		{Path: "/users", ExpectedStatus: http.StatusOK},
	}
	mockConfig := newMockConfig("my-server", endpoints)
	report, err := RunTests(mockConfig, false)
	assert.NoError(t, err)
	for i, result := range report.Results {
		assert.Contains(t, result.Target, endpoints[i].Path)
		assert.False(t, result.Passed)
	}
}

func TestRunTestsUnreachableFailFast(t *testing.T) {
	endpoints := []config.Endpoint{
		{Path: "/users", ExpectedStatus: http.StatusOK},
	}
	mockConfig := newMockConfig("my-server", endpoints)
	report, err := RunTests(mockConfig, true)
	assert.ErrorContains(t, err, "failed to reach target my-server/users")
	assert.Nil(t, report.Results)
}
