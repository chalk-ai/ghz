package runner

import (
	"fmt"
	"log/slog"
)

// SlogAdapter adapts slog.Logger to the runner.Logger interface
type SlogAdapter struct {
	logger *slog.Logger
}

// NewSlogAdapter creates a new slog adapter
func NewSlogAdapter(logger *slog.Logger) Logger {
	return &SlogAdapter{logger: logger}
}

// Debug logs a debug message
func (s *SlogAdapter) Debug(args ...interface{}) {
	s.logger.Debug(sprint(args...))
}

// Debugf logs a formatted debug message
func (s *SlogAdapter) Debugf(template string, args ...interface{}) {
	s.logger.Debug(sprintf(template, args...))
}

// Debugw logs a debug message with structured key-value pairs
func (s *SlogAdapter) Debugw(msg string, keysAndValues ...interface{}) {
	s.logger.Debug(msg, convertToSlogArgs(keysAndValues...)...)
}

// Error logs an error message
func (s *SlogAdapter) Error(args ...interface{}) {
	s.logger.Error(sprint(args...))
}

// Errorf logs a formatted error message
func (s *SlogAdapter) Errorf(template string, args ...interface{}) {
	s.logger.Error(sprintf(template, args...))
}

// Errorw logs an error message with structured key-value pairs
func (s *SlogAdapter) Errorw(msg string, keysAndValues ...interface{}) {
	s.logger.Error(msg, convertToSlogArgs(keysAndValues...)...)
}

// convertToSlogArgs converts key-value pairs to slog.Attr
func convertToSlogArgs(keysAndValues ...interface{}) []interface{} {
	if len(keysAndValues)%2 != 0 {
		// If odd number, just return as-is (slog will handle it)
		return keysAndValues
	}

	attrs := make([]interface{}, 0, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			// If key is not a string, skip this pair
			continue
		}
		attrs = append(attrs, slog.Any(key, keysAndValues[i+1]))
	}
	return attrs
}

// Helper functions for formatting
func sprint(args ...interface{}) string {
	return fmt.Sprint(args...)
}

func sprintf(template string, args ...interface{}) string {
	return fmt.Sprintf(template, args...)
}
