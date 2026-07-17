package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/exilesprx/zig-installer/internal/tui"
)

// ILogger defines the interface for logging
type ILogger interface {
	Close() error
	Log(message string)
	LogError(format string, args ...interface{})
	LogInfo(format string, args ...interface{})
}

// FileLogger implements logging to a file
type FileLogger struct {
	file      *os.File
	enableLog bool
	logFile   string
}

// NewFileLogger creates a new logger instance
func NewFileLogger(logFile string, enableLog bool) (*FileLogger, error) {
	if !enableLog {
		return &FileLogger{
			file:      nil,
			enableLog: enableLog,
			logFile:   logFile,
		}, nil
	}

	// Get executable directory
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}

	execDir := filepath.Dir(execPath)
	logPath := filepath.Join(execDir, logFile)

	// Open log file for appending
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &FileLogger{
		file:      file,
		enableLog: enableLog,
		logFile:   logFile,
	}, nil
}

// Close closes the log file
func (l *FileLogger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// Log logs a message with timestamp
func (l *FileLogger) Log(message string) {
	if l.file == nil || !l.enableLog {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMessage := fmt.Sprintf("[%s] %s\n", timestamp, message)

	_, err := l.file.WriteString(logMessage)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write to log file: %v\n", err)
	}
}

// LogError logs an error with timestamp
func (l *FileLogger) LogError(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.Log(fmt.Sprintf("ERROR: %s", message))
}

// LogInfo logs an informational message with timestamp
func (l *FileLogger) LogInfo(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.Log(fmt.Sprintf("INFO: %s", message))
}

// ConsoleLogger logs to the console using tui styles
type ConsoleLogger struct {
	styles  *tui.Styles
	noColor bool
}

// NewConsoleLogger creates a new console logger with tui styling
func NewConsoleLogger(styles *tui.Styles, noColor bool) *ConsoleLogger {
	return &ConsoleLogger{
		styles:  styles,
		noColor: noColor,
	}
}

// Close is a no-op for console logger
func (l *ConsoleLogger) Close() error {
	return nil
}

// Log logs a message to the console
func (l *ConsoleLogger) Log(message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	stamped := fmt.Sprintf("[%s] %s", timestamp, message)
	if l.styles != nil && !l.noColor {
		fmt.Println(tui.PrintWithStyles(stamped, l.styles.Detail, l.noColor))
	} else {
		fmt.Println(stamped)
	}
}

// LogError logs an error message to the console
func (l *ConsoleLogger) LogError(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	stamped := fmt.Sprintf("ERROR: %s", message)
	if l.styles != nil && !l.noColor {
		fmt.Println(tui.PrintWithStyles(stamped, l.styles.Error, l.noColor))
	} else {
		fmt.Println(stamped)
	}
}

// LogInfo logs an info message to the console
func (l *ConsoleLogger) LogInfo(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	stamped := fmt.Sprintf("INFO: %s", message)
	if l.styles != nil && !l.noColor {
		fmt.Println(tui.PrintWithStyles(stamped, l.styles.Info, l.noColor))
	} else {
		fmt.Println(stamped)
	}
}

// MultiLogger delegates logging to multiple ILogger instances
type MultiLogger struct {
	loggers []ILogger
}

// NewMultiLogger creates a MultiLogger that wraps multiple loggers
func NewMultiLogger(loggers ...ILogger) *MultiLogger {
	return &MultiLogger{loggers: loggers}
}

// Close closes all wrapped loggers
func (m *MultiLogger) Close() error {
	for _, l := range m.loggers {
		if err := l.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Log delegates to all wrapped loggers
func (m *MultiLogger) Log(message string) {
	for _, l := range m.loggers {
		l.Log(message)
	}
}

// LogError delegates to all wrapped loggers
func (m *MultiLogger) LogError(format string, args ...interface{}) {
	for _, l := range m.loggers {
		l.LogError(format, args...)
	}
}

// LogInfo delegates to all wrapped loggers
func (m *MultiLogger) LogInfo(format string, args ...interface{}) {
	for _, l := range m.loggers {
		l.LogInfo(format, args...)
	}
}
