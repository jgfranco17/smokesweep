package core

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jgfranco17/smokesweep/config"
	"github.com/jgfranco17/smokesweep/logging"
	"github.com/sirupsen/logrus"

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

func newContextWithLogger(t *testing.T) (context.Context, bytes.Buffer) {
	t.Helper()
	ctx := context.Background()
	var buf bytes.Buffer

	logger := logging.New(os.Stdout, logrus.TraceLevel)
	ctx = logging.ApplyToContext(ctx, logger)

	return ctx, buf
}

func TestRunTestsSuccess(t *testing.T) {
	ctx, _ := newContextWithLogger(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	endpoints := []config.Endpoint{
		{Path: "/users", ExpectedStatus: 200},
	}
	mockConfig := newMockConfig(server.URL, endpoints)
	report, err := RunTests(ctx, mockConfig, false)
	assert.NoError(t, err)
	assert.Equal(t, len(endpoints), len(report.Results))
	for i, result := range report.Results {
		assert.Contains(t, result.Target, endpoints[i].Path)
		assert.True(t, result.Passed)
	}
}

func TestRunTestsFailedCall(t *testing.T) {
	ctx, _ := newContextWithLogger(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	endpoints := []config.Endpoint{
		{Path: "/users", ExpectedStatus: 200},
	}
	mockConfig := newMockConfig(server.URL, endpoints)
	report, err := RunTests(ctx, mockConfig, false)
	assert.NoError(t, err)
	assert.Equal(t, len(endpoints), len(report.Results))
	for i, result := range report.Results {
		assert.Contains(t, result.Target, endpoints[i].Path)
		assert.False(t, result.Passed)
	}
}

func TestRunTestsFailFast(t *testing.T) {
	ctx, _ := newContextWithLogger(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	endpoints := []config.Endpoint{
		{Path: "/users", ExpectedStatus: 200},
	}
	mockConfig := newMockConfig(server.URL, endpoints)
	report, err := RunTests(ctx, mockConfig, true)
	assert.ErrorContains(t, err, "expected HTTP 200 but got 500")
	assert.Nil(t, report.Results)
}

func TestRunTestsUnreachable(t *testing.T) {
	ctx, _ := newContextWithLogger(t)

	endpoints := []config.Endpoint{
		{Path: "/users", ExpectedStatus: http.StatusOK},
	}
	mockConfig := newMockConfig("my-server", endpoints)
	report, err := RunTests(ctx, mockConfig, false)
	assert.NoError(t, err)
	for i, result := range report.Results {
		assert.Contains(t, result.Target, endpoints[i].Path)
		assert.False(t, result.Passed)
	}
}

func TestRunTestsUnreachableFailFast(t *testing.T) {
	ctx, _ := newContextWithLogger(t)

	endpoints := []config.Endpoint{
		{Path: "/users", ExpectedStatus: http.StatusOK},
	}
	mockConfig := newMockConfig("my-server", endpoints)
	report, err := RunTests(ctx, mockConfig, true)
	assert.ErrorContains(t, err, "failed to reach target my-server/users")
	assert.Nil(t, report.Results)
}
