package runner

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/jgfranco17/dev-tooling-go/logging"
	"github.com/jgfranco17/smokesweep/config"
	"github.com/jgfranco17/smokesweep/outputs"
	"github.com/sirupsen/logrus"
)

const (
	DefaultConfigFile string = ".smokesweep.yaml"
)

// Execute runs the provided test suite and returns the test report.
func Execute(ctx context.Context, conf *config.TestSuite, failFast bool) (TestReport, error) {
	logger := logging.FromContext(ctx)
	var results []TestResult
	logger.WithFields(logrus.Fields{
		"count": len(conf.Endpoints),
		"url":   conf.URL,
	}).Info("Starting test execution")

	testRunStartTime := time.Now()
	for _, endpoint := range conf.Endpoints {
		target := joinURL(conf.URL, endpoint.Path)
		logger.WithFields(logrus.Fields{
			"target": target,
		}).Info("Pinging target")
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

// PingURL make a simple GET request to a provided URL for liveness.
func PingURL(ctx context.Context, url string, timeout time.Duration) error {
	logger := logging.FromContext(ctx).WithFields(
		logrus.Fields{
			"url":     url,
			"timeout": timeout,
		},
	)
	client := &http.Client{
		Timeout: timeout,
	}

	start := time.Now()
	logger.WithFields(logrus.Fields{
		"url":     url,
		"timeout": timeout,
	}).Debug("Checking URL for liveness")
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

func joinURL(base string, paths ...string) string {
	p := path.Join(paths...)
	return fmt.Sprintf("%s/%s", strings.TrimRight(base, "/"), strings.TrimLeft(p, "/"))
}
