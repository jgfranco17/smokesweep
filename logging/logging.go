package logging

import (
	"context"
	"io"
	"time"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

type contextLogKey string

const contextKey contextLogKey = "logger"

func setOutputColorPerLevel(level string) color.Attribute {
	var selectedColor color.Attribute
	switch level {
	case "DEBUG":
		selectedColor = color.FgCyan
	case "INFO":
		selectedColor = color.FgGreen
	case "WARN", "WARNING":
		selectedColor = color.FgYellow
	case "ERROR", "PANIC", "FATAL":
		selectedColor = color.FgRed
	default:
		selectedColor = color.FgWhite
	}
	return selectedColor
}

func New(stream io.Writer, level logrus.Level) *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(stream)
	logger.SetLevel(level)

	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors:          false,
		PadLevelText:           true,
		QuoteEmptyFields:       true,
		FullTimestamp:          true,
		DisableSorting:         true,
		DisableLevelTruncation: true,
		TimestampFormat:        time.DateTime,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "@timestamp",
			logrus.FieldKeyLevel: "@level",
			logrus.FieldKeyMsg:   "@message",
		},
	})
	return logger
}

func ApplyToContext(ctx context.Context, logger *logrus.Logger) context.Context {
	return context.WithValue(ctx, contextKey, logger)
}

func FromContext(ctx context.Context) *logrus.Logger {
	if logger, ok := ctx.Value(contextKey).(*logrus.Logger); ok {
		return logger
	}
	panic("no logger set in context")
}
