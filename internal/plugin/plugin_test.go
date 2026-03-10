package plugin

import (
	"fmt"
	"testing"

	"github.com/zyrthi-io/zyrthi-flash/internal/serial"
)

// MockFlashPlugin 模拟 Flash 插件
type MockFlashPlugin struct {
	initError   error
	detectError error
	flashError  error
	eraseError  error
	resetError  error
	chip        string
	flashCalls  int
	eraseCalls  int
	resetCalls  int
}

func (m *MockFlashPlugin) Init(api *serial.HostAPI) error {
	return m.initError
}

func (m *MockFlashPlugin) Detect() (string, error) {
	return m.chip, m.detectError
}

func (m *MockFlashPlugin) Flash(chip string, firmware []byte, offset uint32) error {
	m.flashCalls++
	return m.flashError
}

func (m *MockFlashPlugin) Erase(chip string, offset uint32, size uint32) error {
	m.eraseCalls++
	return m.eraseError
}

func (m *MockFlashPlugin) Reset(chip string) error {
	m.resetCalls++
	return m.resetError
}

func (m *MockFlashPlugin) Close() error {
	return nil
}

func TestFlashPluginInterface(t *testing.T) {
	mock := &MockFlashPlugin{chip: "test-chip"}

	if err := mock.Init(nil); err != nil {
		t.Errorf("Init error: %v", err)
	}

	chip, err := mock.Detect()
	if err != nil {
		t.Errorf("Detect error: %v", err)
	}
	if chip != "test-chip" {
		t.Errorf("expected 'test-chip', got %s", chip)
	}

	if err := mock.Flash("test-chip", []byte{0x00, 0x01}, 0); err != nil {
		t.Errorf("Flash error: %v", err)
	}
}

func TestMockFlashPluginErrors(t *testing.T) {
	mock := &MockFlashPlugin{
		initError:   fmt.Errorf("init failed"),
		detectError: fmt.Errorf("detect failed"),
		flashError:  fmt.Errorf("flash failed"),
	}

	if mock.Init(nil) == nil {
		t.Error("expected init error")
	}
	if _, err := mock.Detect(); err == nil {
		t.Error("expected detect error")
	}
	if mock.Flash("chip", nil, 0) == nil {
		t.Error("expected flash error")
	}
}

func TestMockFlashPluginCallCount(t *testing.T) {
	mock := &MockFlashPlugin{chip: "test-chip"}

	for i := 0; i < 3; i++ {
		mock.Flash("test-chip", []byte{0x00}, 0)
	}
	if mock.flashCalls != 3 {
		t.Errorf("expected 3 flash calls, got %d", mock.flashCalls)
	}
}
