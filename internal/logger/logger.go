package logger

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/exilesprx/zig-installer/internal/tui"
)

type Logger interface {
	Close() error
	Log(message string)
	LogError(format string, args ...any)
	LogInfo(format string, args ...any)
}

type FileLogger struct {
	file      *os.File
	enableLog bool
	logFile   string
}

func NewFileLogger(logFile string, enableLog bool) (*FileLogger, error) {
	if !enableLog {
		return &FileLogger{
			file:      nil,
			enableLog: enableLog,
			logFile:   logFile,
		}, nil
	}

	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}

	execDir := filepath.Dir(execPath)
	logPath := filepath.Join(execDir, logFile)

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

func (l *FileLogger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

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

func (l *FileLogger) LogError(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	l.Log(fmt.Sprintf("ERROR: %s", message))
}

func (l *FileLogger) LogInfo(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	l.Log(fmt.Sprintf("INFO: %s", message))
}

type ConsoleLogger struct {
	styles  *tui.Styles
	noColor bool
}

func NewConsoleLogger(styles *tui.Styles, noColor bool) *ConsoleLogger {
	return &ConsoleLogger{
		styles:  styles,
		noColor: noColor,
	}
}

func (l *ConsoleLogger) Close() error {
	return nil
}

func (l *ConsoleLogger) Log(message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	stamped := fmt.Sprintf("[%s] %s", timestamp, message)
	if l.styles != nil && !l.noColor {
		fmt.Println(tui.PrintWithStyles(stamped, l.styles.Detail, l.noColor))
	} else {
		fmt.Println(stamped)
	}
}

func (l *ConsoleLogger) LogError(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	stamped := fmt.Sprintf("ERROR: %s", message)
	if l.styles != nil && !l.noColor {
		fmt.Println(tui.PrintWithStyles(stamped, l.styles.Error, l.noColor))
	} else {
		fmt.Println(stamped)
	}
}

func (l *ConsoleLogger) LogInfo(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	stamped := fmt.Sprintf("INFO: %s", message)
	if l.styles != nil && !l.noColor {
		fmt.Println(tui.PrintWithStyles(stamped, l.styles.Info, l.noColor))
	} else {
		fmt.Println(stamped)
	}
}

type MultiLogger struct {
	loggers []Logger
}

func NewMultiLogger(loggers ...Logger) *MultiLogger {
	return &MultiLogger{loggers: loggers}
}

func (m *MultiLogger) Close() error {
	var errs error
	for _, l := range m.loggers {
		if err := l.Close(); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return errs
}

func (m *MultiLogger) Log(message string) {
	for _, l := range m.loggers {
		l.Log(message)
	}
}

func (m *MultiLogger) LogError(format string, args ...any) {
	for _, l := range m.loggers {
		l.LogError(format, args...)
	}
}

func (m *MultiLogger) LogInfo(format string, args ...any) {
	for _, l := range m.loggers {
		l.LogInfo(format, args...)
	}
}
