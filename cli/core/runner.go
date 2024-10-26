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
	Target           string
	Duration         time.Duration
	ExpectedDuration *time.Duration
	HttpStatus       int
	ExpectedStatus   int
	Passed           bool
}

type TestReport struct {
	Timestamp time.Time
	Results   []TestResult
}

func RunTests(conf *config.TestConfig) []TestResult {
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
			continue
		}
		defer resp.Body.Close()
		result := TestResult{
			Target:         target,
			Duration:       duration,
			ExpectedStatus: endpoint.ExpectedStatus,
			HttpStatus:     resp.StatusCode,
			Passed:         resp.StatusCode == endpoint.ExpectedStatus,
		}
		if endpoint.MaxResponseTime != nil {
			d := time.Duration(*endpoint.MaxResponseTime) * time.Millisecond
			result.ExpectedDuration = &d
		}
		results = append(results, result)
	}
	return results
}

func SummarizeResults(results []TestResult) {
	for _, result := range results {
		if result.Passed {
			if result.ExpectedDuration != nil {
				if !(result.Duration < *result.ExpectedDuration) {
					outputs.PrintColoredMessage("yellow", "SLOW", "%s (%vms) exceeded threshold", result.Target, result.Duration.Milliseconds())
				}
			}
			outputs.PrintColoredMessage("green", "SUCCESS", "%s (%vms) OK", result.Target, result.Duration.Milliseconds())
		} else {
			outputs.PrintColoredMessage("red", "FAILED", "Target '%s' expected HTTP status %d but got %d", result.Target, result.ExpectedStatus, result.HttpStatus)
		}
	}
}
