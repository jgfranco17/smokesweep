package core

import (
	"fmt"
	"net/http"
	"time"

	"cli/config"
	"cli/outputs"

	log "github.com/sirupsen/logrus"
)

type TestResult struct {
	Target         string
	Duration       time.Duration
	Timeout        *time.Duration
	HttpStatus     int
	ExpectedStatus int
	Passed         bool
}

type TestReport struct {
	Timestamp time.Time
	Results   []TestResult
}

func RunTests(conf *config.TestConfig, failFast bool) ([]TestResult, error) {
	var results []TestResult
	fmt.Printf("Running %d tests [URL: %s]\n", len(conf.Endpoints), conf.URL)
	fmt.Println("------------------------------")
	for _, endpoint := range conf.Endpoints {
		target := fmt.Sprintf("%s%s", conf.URL, endpoint.Path)
		log.Debugf("Pinging %s", target)
		start := time.Now()
		resp, err := http.Get(target)
		duration := time.Since(start)
		if err != nil {
			outputs.PrintColoredMessage("red", "UNREACHABLE", "Failed to reach target %s", target)
			if failFast {
				return nil, fmt.Errorf("Failed to reach target %s: %w", target, err)
			}
			continue
		}
		defer resp.Body.Close()
		if failFast && resp.StatusCode != endpoint.ExpectedStatus {
			return nil, fmt.Errorf("Target %s expected HTTP %d but got %d", target, endpoint.ExpectedStatus, resp.StatusCode)
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
	return results, nil
}

func SummarizeResults(results []TestResult) {
	for _, result := range results {
		if result.Passed {
			if result.Timeout != nil {
				if !(result.Duration < *result.Timeout) {
					outputs.PrintColoredMessage("yellow", "SLOW", "%s (%vms) exceeded threshold", result.Target, result.Duration.Milliseconds())
				}
			}
			outputs.PrintColoredMessage("green", "SUCCESS", "%s (%vms) OK", result.Target, result.Duration.Milliseconds())
		} else {
			outputs.PrintColoredMessage("red", "FAILED", "Target '%s' expected HTTP status %d but got %d", result.Target, result.ExpectedStatus, result.HttpStatus)
		}
	}
}
