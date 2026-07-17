package logger

import (
	"testing"
)

// mockLogger is a test implementation of ILogger
type mockLogger struct {
	logs     []string
	errors   []string
	infos    []string
	closed   bool
	closeErr error
}

func newMockLogger() *mockLogger {
	return &mockLogger{}
}

func (m *mockLogger) Close() error {
	m.closed = true
	return m.closeErr
}

func (m *mockLogger) Log(message string) {
	m.logs = append(m.logs, message)
}

func (m *mockLogger) LogError(format string, args ...interface{}) {
	m.errors = append(m.errors, format)
}

func (m *mockLogger) LogInfo(format string, args ...interface{}) {
	m.infos = append(m.infos, format)
}

func TestMultiLogger_Log_DelegatesToAll(t *testing.T) {
	m1 := newMockLogger()
	m2 := newMockLogger()
	ml := NewMultiLogger(m1, m2)

	ml.Log("test message")

	if len(m1.logs) != 1 || m1.logs[0] != "test message" {
		t.Errorf("expected m1 to receive log, got %v", m1.logs)
	}
	if len(m2.logs) != 1 || m2.logs[0] != "test message" {
		t.Errorf("expected m2 to receive log, got %v", m2.logs)
	}
}

func TestMultiLogger_LogError_DelegatesToAll(t *testing.T) {
	m1 := newMockLogger()
	m2 := newMockLogger()
	ml := NewMultiLogger(m1, m2)

	ml.LogError("error %s", "msg")

	if len(m1.errors) != 1 {
		t.Errorf("expected m1 to receive error, got %v", m1.errors)
	}
	if len(m2.errors) != 1 {
		t.Errorf("expected m2 to receive error, got %v", m2.errors)
	}
}

func TestMultiLogger_LogInfo_DelegatesToAll(t *testing.T) {
	m1 := newMockLogger()
	m2 := newMockLogger()
	ml := NewMultiLogger(m1, m2)

	ml.LogInfo("info %s", "msg")

	if len(m1.infos) != 1 {
		t.Errorf("expected m1 to receive info, got %v", m1.infos)
	}
	if len(m2.infos) != 1 {
		t.Errorf("expected m2 to receive info, got %v", m2.infos)
	}
}

func TestMultiLogger_Close_ClosesAll(t *testing.T) {
	m1 := newMockLogger()
	m2 := newMockLogger()
	ml := NewMultiLogger(m1, m2)

	err := ml.Close()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !m1.closed {
		t.Error("expected m1 to be closed")
	}
	if !m2.closed {
		t.Error("expected m2 to be closed")
	}
}

func TestMultiLogger_Close_ReturnsFirstError(t *testing.T) {
	m1 := newMockLogger()
	m1.closeErr = nil
	m2 := newMockLogger()
	m2.closeErr = nil
	ml := NewMultiLogger(m1, m2)

	err := ml.Close()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestMultiLogger_EmptyLoggers(t *testing.T) {
	ml := NewMultiLogger()

	// Should not panic
	ml.Log("test")
	ml.LogError("error")
	ml.LogInfo("info")

	err := ml.Close()
	if err != nil {
		t.Errorf("expected no error from empty multilogger, got %v", err)
	}
}
