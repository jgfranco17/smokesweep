package runner

import (
	"fmt"
	"net/http"
	"time"

	"cli/config"

	"github.com/fatih/color"
)

func RunTests(conf *config.TestConfig) error {
	for _, endpoint := range conf.Endpoints {
		fullURL := fmt.Sprintf("%s%s", conf.URL, endpoint.Path)

		start := time.Now()
		resp, err := http.Get(fullURL)
		duration := time.Since(start)

		if err != nil {
			printResult("FAIL", fullURL, duration, err.Error())
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == endpoint.ExpectedStatus {
			if endpoint.MaxResponseTime != nil {
				if duration < time.Duration(*endpoint.MaxResponseTime)*time.Millisecond {
					printResult("SUCCESS", fullURL, duration, "")
				} else {
					printResult("SLOW", fullURL, duration, "Response time exceeded threshold")
				}
			}
		} else {
			printResult("FAIL", fullURL, duration, fmt.Sprintf("Expected %d, got %d", endpoint.ExpectedStatus, resp.StatusCode))
		}
	}
	return nil
}

func printResult(status, url string, duration time.Duration, message string) {
	switch status {
	case "SUCCESS":
		color.New(color.FgGreen).Printf("[SUCCESS] %s - %s (%v)\n", url, message, duration)
	case "SLOW":
		color.New(color.FgYellow).Printf("[SLOW] %s - %s (%v)\n", url, message, duration)
	case "FAIL":
		color.New(color.FgRed).Printf("[FAIL] %s - %s (%v)\n", url, message, duration)
	}
}
