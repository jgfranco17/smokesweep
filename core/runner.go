package core

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jgfranco17/smokesweep/config"
	"github.com/jgfranco17/smokesweep/logging"
	"github.com/jgfranco17/smokesweep/outputs"
)

// RunTests executes the provided test suite and returns the test report.
func RunTests(ctx context.Context, conf *config.TestSuite, failFast bool) (TestReport, error) {
	logger := logging.FromContext(ctx)
	var results []TestResult
	logger.Infof("Running %d tests [URL: %s]\n", len(conf.Endpoints), conf.URL)

	testRunStartTime := time.Now()
	for _, endpoint := range conf.Endpoints {
		target := fmt.Sprintf("%s%s", conf.URL, endpoint.Path)
		logger.Debugf("Pinging %s", target)
		start := time.Now()
		resp, err := http.Get(target)
		duration := time.Since(start)
		if err != nil {
			outputs.PrintColoredMessage("red", "UNREACHABLE", "Failed to reach target %s", target)
			if failFast {
				return TestReport{}, fmt.Errorf("failed to reach target %s: %w", target, err)
			}
			continue
		}
		defer resp.Body.Close()
		if failFast && resp.StatusCode != endpoint.ExpectedStatus {
			return TestReport{}, fmt.Errorf("target %s expected HTTP %d but got %d", target, endpoint.ExpectedStatus, resp.StatusCode)
		}
		result := TestResult{
			Target:         target,
			Duration:       duration,
			ExpectedStatus: endpoint.ExpectedStatus,
			HttpStatus:     resp.StatusCode,
			Passed:         resp.StatusCode == endpoint.ExpectedStatus,
		}
		if endpoint.Timeout != nil {
			d := time.Duration(*endpoint.Timeout) * time.Millisecond
			result.Timeout = &d
		}
		results = append(results, result)
	}
	return TestReport{
		Timestamp: testRunStartTime,
		Results:   results,
	}, nil
}

/*
Description: Ping a provided URL for liveness.

[IN] url (string): Target URL to ping

[IN] timeoutSeconds (int): Timeout duration for HTTP client

[OUT] error: Any error occurred during the test run
*/
func PingUrl(ctx context.Context, url string, timeoutSeconds int) error {
	logger := logging.FromContext(ctx)

	client := &http.Client{
		Timeout: time.Duration(timeoutSeconds) * time.Second,
	}

	start := time.Now()
	logger.Debugf("Checking URL %s for liveness", url)
	resp, err := client.Head(url)
	duration := time.Since(start)
	if err != nil {
		return fmt.Errorf("failed to reach target %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		outputs.PrintColoredMessage("green", "LIVE", "Target %s responded in %vms", url, duration.Milliseconds())
	} else {
		outputs.PrintColoredMessage("red", "DOWN", "Target %s returned HTTP status %d", url, resp.StatusCode)
	}
	return nil
}
