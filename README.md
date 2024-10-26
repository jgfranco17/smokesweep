# SmokeSweep

![Build Status](https://github.com/jgfranco17/smokesweep/actions/workflows/cicd.yaml/badge.svg?style=for-the-badge)

## Overview

**SmokeSweep** is a command-line interface (CLI) tool for running smoke tests on API services.
It allows users to define a configuration in YAML format where they can specify the base URL
and individual endpoints along with their expected HTTP status codes. SmokeSweep then pings
these endpoints, records the response times, and provides user-friendly output, indicating the
success or failure of each endpoint check.

### Key Features

- Simple YAML configuration for defining API endpoints and expected statuses
- Color-coded output for quick visual feedback
- Supports customizable verbosity levels to enhance log visibility

## Installation

### Prerequisites

- Go (version 1.20 or later)

### Build from Source

1. Clone the repository:

   ```bash
   git clone https://github.com/jgfranco17/smokesweep.git
   cd smokesweep
   ```

2. Build the binary:

   ```bash
   go build -o smokesweep main.go
   ```

### Pre-built Binaries

You can also download the latest release from the [releases page](https://github.com/jgfranco17/smokesweep/releases).

## Usage

### Configuration

Create a YAML configuration file (e.g., `config.yaml`) with the following structure:

```yaml
Copy code
url: "https://api.example.com"
endpoints:
  - path: "/users"
    expected-status: 200
    max-response-time: 500
  - path: "/posts"
    expected-status: 200
```

### Running SmokeSweep

To run SmokeSweep, use the following command:

```bash
smokewweep run -vv ./config.yaml
```

`run`: The command to execute the smoke tests.
`-vvv`: Increases verbosity of the output for more detailed logs.
`./config.yaml`: The path to your configuration file.

The output will display the results of the smoke tests for each endpoint defined in the
configuration file.
