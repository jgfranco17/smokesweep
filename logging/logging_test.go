package logging

import (
	"bytes"
	"context"
	"testing"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSetLoggingLevelPerColor(t *testing.T) {
	expectedColorPerLevel := map[string]color.Attribute{
		// Valid cases
		"DEBUG": color.FgCyan,
		"INFO":  color.FgGreen,
		"WARN":  color.FgYellow,
		"ERROR": color.FgRed,
		// Sample invalid case hits default
		"INVALID": color.FgWhite,
	}
	for level, color := range expectedColorPerLevel {
		assert.Equal(t, color, setOutputColorPerLevel(level))
	}
}

func TestApplyToContext(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, logrus.TraceLevel)
	ctx := WithContext(context.Background(), logger)
	assert.Equal(t, logger, FromContext(ctx))
}

func TestFromContext(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, logrus.TraceLevel)
	ctx := WithContext(context.Background(), logger)
	assert.Equal(t, logger, FromContext(ctx))
}
