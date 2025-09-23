package runner

import (
	"fmt"
	"time"

	"github.com/jgfranco17/smokesweep/outputs"
)

type TestResult struct {
	// Target is the URL of the endpoint that was tested.
	Target string

	// Duration is the time it took to test the endpoint.
	Duration time.Duration

	// Timeout is the timeout for the test.
	Timeout *time.Duration

	// HttpStatus is the HTTP status code of the response.
	HttpStatus int

	// ExpectedStatus is the expected HTTP status code of the response.
	ExpectedStatus int

	// Passed is true if the test passed, false otherwise.
	Passed bool
}

type TestReport struct {
	// Timestamp is the timestamp of the test.
	Timestamp time.Time

	// Results is the list of test results.
	Results []TestResult
}

/*
Description: Prints a summary of the test results with colorized output.

[OUT] error: Any error occurred during the summary printing
*/
func (tr *TestReport) SummarizeResults() error {
	if len(tr.Results) < 1 {
		return fmt.Errorf("no test results to print.")
	}
	fmt.Println("------------------------------")
	for _, result := range tr.Results {
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
	return nil
}
