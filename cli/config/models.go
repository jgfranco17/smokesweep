package config

type TestConfig struct {
	URL       string     `yaml:"url"`
	Endpoints []Endpoint `yaml:"endpoints"`
}

type Endpoint struct {
	Path            string `yaml:"path"`
	ExpectedStatus  int    `yaml:"expected-status"`
	MaxResponseTime *int   `yaml:"max-response-time-ms,omitempty"`
}
