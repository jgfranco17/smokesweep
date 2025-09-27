package runner

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/jgfranco17/dev-tooling-go/logging"
	"github.com/jgfranco17/smokesweep/config"
	"github.com/jgfranco17/smokesweep/outputs"
	"github.com/sirupsen/logrus"
)

const (
	DefaultConfigFile string = ".smokesweep.yaml"
)

// TestJob represents a single test job to be executed
type TestJob struct {
	Endpoint config.Endpoint
	Target   string
	Index    int
}

// TestResultWithIndex wraps TestResult with an index for ordering
type TestResultWithIndex struct {
	Result TestResult
	Index  int
}

// Execute runs the provided test suite asynchronously and returns the test report.
func Execute(ctx context.Context, conf *config.TestSuite, failFast bool) (TestReport, error) {
	logger := logging.FromContext(ctx)
	logger.WithFields(logrus.Fields{
		"count": len(conf.Endpoints),
		"url":   conf.URL,
	}).Info("Starting async test execution")

	testRunStartTime := time.Now()

	// Handle empty endpoints list
	if len(conf.Endpoints) == 0 {
		return TestReport{
			Timestamp: testRunStartTime,
			Results:   []TestResult{},
		}, nil
	}

	// Create channels for job distribution and result collection
	jobChan := make(chan TestJob, len(conf.Endpoints))
	resultChan := make(chan TestResultWithIndex, len(conf.Endpoints))
	errorChan := make(chan error, len(conf.Endpoints))

	// Create context with cancellation for fail-fast behavior
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start worker goroutines
	numWorkers := len(conf.Endpoints)
	if numWorkers > 10 { // Limit max workers to prevent resource exhaustion
		numWorkers = 10
	}

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(ctx, &wg, jobChan, resultChan, errorChan, logger, failFast)
	}

	// Send jobs to workers
	go func() {
		defer close(jobChan)
		for i, endpoint := range conf.Endpoints {
			target := joinURL(conf.URL, endpoint.Path)
			select {
			case jobChan <- TestJob{Endpoint: endpoint, Target: target, Index: i}:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	// Collect results
	results := make([]TestResult, len(conf.Endpoints))
	completed := 0
	hasError := false

	for completed < len(conf.Endpoints) {
		select {
		case result, ok := <-resultChan:
			if !ok {
				// Channel closed, all workers done
				break
			}
			results[result.Index] = result.Result
			completed++

		case err, ok := <-errorChan:
			if !ok {
				// Channel closed, all workers done
				break
			}
			hasError = true
			if failFast {
				cancel() // Cancel all remaining workers
				return TestReport{}, err
			}
			// For non-fail-fast, we'll continue collecting results
			completed++

		case <-ctx.Done():
			return TestReport{}, ctx.Err()
		}
	}

	// For non-fail-fast mode with errors, filter out empty results
	if hasError && !failFast {
		filteredResults := make([]TestResult, 0, len(results))
		for _, result := range results {
			if result.Target != "" { // Only include results that were actually processed
				filteredResults = append(filteredResults, result)
			}
		}
		results = filteredResults
	}

	return TestReport{
		Timestamp: testRunStartTime,
		Results:   results,
	}, nil
}

// worker processes test jobs from the job channel
func worker(ctx context.Context, wg *sync.WaitGroup, jobChan <-chan TestJob, resultChan chan<- TestResultWithIndex, errorChan chan<- error, logger *logrus.Logger, failFast bool) {
	defer wg.Done()

	for {
		select {
		case job, ok := <-jobChan:
			if !ok {
				return
			}

			logger.WithFields(logrus.Fields{
				"target": job.Target,
			}).Info("Pinging target")

			result, err := executeSingleTest(ctx, job)
			if err != nil {
				outputs.PrintColoredMessage("red", "UNREACHABLE", "Failed to reach target %s", job.Target)
				errorChan <- fmt.Errorf("failed to reach target %s: %w", job.Target, err)
				return
			}

			// Check for status code mismatch
			if result.HttpStatus != result.ExpectedStatus {
				if failFast {
					errorChan <- fmt.Errorf("target %s expected HTTP %d but got %d", job.Target, result.ExpectedStatus, result.HttpStatus)
					return
				}
				// For non-fail-fast, still send the result but mark it as failed
			}

			select {
			case resultChan <- TestResultWithIndex{Result: result, Index: job.Index}:
			case <-ctx.Done():
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

// executeSingleTest executes a single test and returns the result
func executeSingleTest(ctx context.Context, job TestJob) (TestResult, error) {
	start := time.Now()

	// Create HTTP client with timeout if specified
	client := &http.Client{}
	if job.Endpoint.Timeout != nil {
		timeout := time.Duration(*job.Endpoint.Timeout) * time.Millisecond
		client.Timeout = timeout
	}

	req, err := http.NewRequestWithContext(ctx, "GET", job.Target, nil)
	if err != nil {
		return TestResult{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return TestResult{}, err
	}
	defer resp.Body.Close()

	duration := time.Since(start)

	result := TestResult{
		Target:         job.Target,
		Duration:       duration,
		ExpectedStatus: job.Endpoint.ExpectedStatus,
		HttpStatus:     resp.StatusCode,
		Passed:         resp.StatusCode == job.Endpoint.ExpectedStatus,
	}

	if job.Endpoint.Timeout != nil {
		d := time.Duration(*job.Endpoint.Timeout) * time.Millisecond
		result.Timeout = &d
	}

	return result, nil
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
