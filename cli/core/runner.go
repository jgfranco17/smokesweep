package core

import (
	"fmt"
	"net/http"
	"time"

	"cli/config"
	"cli/outputs"

	log "github.com/sirupsen/logrus"
)

/*
Description: Runs the provided test suite and returns the test report.

[IN] conf (*config.TestConfig): A pointer to the test suite configuration

[OUT] TestReport: A test report containing the test results

[OUT] error: Any error occurred during the test run
*/
func RunTests(conf *config.TestConfig, failFast bool) (TestReport, error) {
	var results []TestResult
	log.Infof("Running %d tests [URL: %s]\n", len(conf.Endpoints), conf.URL)

	testRunStartTime := time.Now()
	for _, endpoint := range conf.Endpoints {
		target := fmt.Sprintf("%s%s", conf.URL, endpoint.Path)
		log.Debugf("Pinging %s", target)
		start := time.Now()
		resp, err := http.Get(target)
		duration := time.Since(start)
		if err != nil {
			outputs.PrintColoredMessage("red", "UNREACHABLE", "Failed to reach target %s", target)
			if failFast {
				return TestReport{}, fmt.Errorf("Failed to reach target %s: %w", target, err)
			}
			continue
		}
		defer resp.Body.Close()
		if failFast && resp.StatusCode != endpoint.ExpectedStatus {
			return TestReport{}, fmt.Errorf("Target %s expected HTTP %d but got %d", target, endpoint.ExpectedStatus, resp.StatusCode)
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
func PingUrl(url string, timeoutSeconds int) error {
	client := &http.Client{
		Timeout: time.Duration(timeoutSeconds) * time.Second,
	}
	start := time.Now()
	log.Debugf("Checking URL %s for liveness", url)
	resp, err := client.Head(url)
	duration := time.Since(start)
	if err != nil {
		return fmt.Errorf("Failed to reach target %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		outputs.PrintColoredMessage("green", "LIVE", "Target %s responded in %vms", url, duration.Milliseconds())
	} else {
		outputs.PrintColoredMessage("red", "DOWN", "Target %s returned HTTP status %d", url, resp.StatusCode)
	}
	return nil
}
