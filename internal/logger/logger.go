// Package logger provides a unified logging interface for the V Panel application.
// It supports JSON and text formats, configurable log levels, and structured fields.
package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// Level represents a log level.
type Level int

const (
	// DebugLevel is the most verbose level.
	DebugLevel Level = iota
	// InfoLevel is the default level.
	InfoLevel
	// WarnLevel is for warnings.
	WarnLevel
	// ErrorLevel is for errors.
	ErrorLevel
	// FatalLevel is for fatal errors that cause the application to exit.
	FatalLevel
)

// String returns the string representation of the log level.
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	default:
		return "unknown"
	}
}

// ParseLevel parses a string into a Level.
func ParseLevel(s string) Level {
	switch strings.ToLower(s) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "fatal":
		return FatalLevel
	default:
		return InfoLevel
	}
}

// Field represents a structured log field.
type Field struct {
	Key   string
	Value any
}

// F creates a new Field.
func F(key string, value any) Field {
	return Field{Key: key, Value: value}
}

// Logger is the interface for logging.
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	With(fields ...Field) Logger
	SetLevel(level Level)
	GetLevel() Level
}

// Config holds logger configuration.
type Config struct {
	Level  string
	Format string
	Output string
}

// JSONLogEntry represents a JSON log entry.
type JSONLogEntry struct {
	Timestamp string         `json:"timestamp"`
	Level     string         `json:"level"`
	Message   string         `json:"message"`
	Fields    map[string]any `json:"fields,omitempty"`
}

// defaultLogger is the default logger implementation.
type defaultLogger struct {
	mu       sync.Mutex
	level    Level
	format   string
	output   io.Writer
	fields   []Field
	exitFunc func(int) // For testing
}

// New creates a new logger with the given configuration.
func New(cfg Config) Logger {
	var output io.Writer
	switch strings.ToLower(cfg.Output) {
	case "stdout", "":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		// Try to open file
		f, err := os.OpenFile(cfg.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			output = os.Stdout
		} else {
			output = f
		}
	}

	format := strings.ToLower(cfg.Format)
	if format != "json" && format != "text" {
		format = "json"
	}

	return &defaultLogger{
		level:    ParseLevel(cfg.Level),
		format:   format,
		output:   output,
		fields:   nil,
		exitFunc: os.Exit,
	}
}

// NewDefault creates a new logger with default configuration.
func NewDefault() Logger {
	return New(Config{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	})
}

// SetLevel sets the log level.
func (l *defaultLogger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel returns the current log level.
func (l *defaultLogger) GetLevel() Level {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

// With returns a new logger with the given fields added.
func (l *defaultLogger) With(fields ...Field) Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newFields := make([]Field, len(l.fields)+len(fields))
	copy(newFields, l.fields)
	copy(newFields[len(l.fields):], fields)

	return &defaultLogger{
		level:    l.level,
		format:   l.format,
		output:   l.output,
		fields:   newFields,
		exitFunc: l.exitFunc,
	}
}

// Debug logs a debug message.
func (l *defaultLogger) Debug(msg string, fields ...Field) {
	l.log(DebugLevel, msg, fields...)
}

// Info logs an info message.
func (l *defaultLogger) Info(msg string, fields ...Field) {
	l.log(InfoLevel, msg, fields...)
}

// Warn logs a warning message.
func (l *defaultLogger) Warn(msg string, fields ...Field) {
	l.log(WarnLevel, msg, fields...)
}

// Error logs an error message.
func (l *defaultLogger) Error(msg string, fields ...Field) {
	l.log(ErrorLevel, msg, fields...)
}

// Fatal logs a fatal message and exits.
func (l *defaultLogger) Fatal(msg string, fields ...Field) {
	l.log(FatalLevel, msg, fields...)
	l.exitFunc(1)
}

// log writes a log entry.
func (l *defaultLogger) log(level Level, msg string, fields ...Field) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.level {
		return
	}

	// Merge fields
	allFields := make([]Field, len(l.fields)+len(fields))
	copy(allFields, l.fields)
	copy(allFields[len(l.fields):], fields)

	timestamp := time.Now().UTC().Format(time.RFC3339Nano)

	if l.format == "json" {
		l.writeJSON(timestamp, level, msg, allFields)
	} else {
		l.writeText(timestamp, level, msg, allFields)
	}
}

// writeJSON writes a JSON log entry.
func (l *defaultLogger) writeJSON(timestamp string, level Level, msg string, fields []Field) {
	entry := JSONLogEntry{
		Timestamp: timestamp,
		Level:     level.String(),
		Message:   msg,
	}

	if len(fields) > 0 {
		entry.Fields = make(map[string]any, len(fields))
		for _, f := range fields {
			entry.Fields[f.Key] = f.Value
		}
	}

	data, err := json.Marshal(entry)
	if err != nil {
		// Fallback to simple format
		fmt.Fprintf(l.output, `{"timestamp":"%s","level":"%s","message":"%s","error":"marshal_failed"}`+"\n",
			timestamp, level.String(), msg)
		return
	}

	l.output.Write(data)
	l.output.Write([]byte("\n"))
}

// writeText writes a text log entry.
func (l *defaultLogger) writeText(timestamp string, level Level, msg string, fields []Field) {
	var sb strings.Builder
	sb.WriteString(timestamp)
	sb.WriteString(" [")
	sb.WriteString(strings.ToUpper(level.String()))
	sb.WriteString("] ")
	sb.WriteString(msg)

	for _, f := range fields {
		sb.WriteString(" ")
		sb.WriteString(f.Key)
		sb.WriteString("=")
		sb.WriteString(fmt.Sprintf("%v", f.Value))
	}

	sb.WriteString("\n")
	l.output.Write([]byte(sb.String()))
}

// GetOutput returns the output writer (for testing).
func (l *defaultLogger) GetOutput() io.Writer {
	return l.output
}

// SetOutput sets the output writer (for testing).
func (l *defaultLogger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = w
}

// SetExitFunc sets the exit function (for testing).
func (l *defaultLogger) SetExitFunc(f func(int)) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.exitFunc = f
}

// Global logger instance
var globalLogger Logger = NewDefault()

// SetGlobal sets the global logger.
func SetGlobal(l Logger) {
	globalLogger = l
}

// Global returns the global logger.
func Global() Logger {
	return globalLogger
}

// Debug logs a debug message using the global logger.
func Debug(msg string, fields ...Field) {
	globalLogger.Debug(msg, fields...)
}

// Info logs an info message using the global logger.
func Info(msg string, fields ...Field) {
	globalLogger.Info(msg, fields...)
}

// Warn logs a warning message using the global logger.
func Warn(msg string, fields ...Field) {
	globalLogger.Warn(msg, fields...)
}

// Error logs an error message using the global logger.
func Error(msg string, fields ...Field) {
	globalLogger.Error(msg, fields...)
}

// Fatal logs a fatal message using the global logger and exits.
func Fatal(msg string, fields ...Field) {
	globalLogger.Fatal(msg, fields...)
}
