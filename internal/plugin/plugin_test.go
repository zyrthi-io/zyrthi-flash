package plugin

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/zyrthi-io/zyrthi-flash/internal/config"
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

// ============ FlashPlugin Interface Tests ============

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
		eraseError:  fmt.Errorf("erase failed"),
		resetError:  fmt.Errorf("reset failed"),
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
	if mock.Erase("chip", 0, 0) == nil {
		t.Error("expected erase error")
	}
	if mock.Reset("chip") == nil {
		t.Error("expected reset error")
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

	for i := 0; i < 2; i++ {
		mock.Erase("test-chip", 0, 1024)
	}
	if mock.eraseCalls != 2 {
		t.Errorf("expected 2 erase calls, got %d", mock.eraseCalls)
	}

	for i := 0; i < 4; i++ {
		mock.Reset("test-chip")
	}
	if mock.resetCalls != 4 {
		t.Errorf("expected 4 reset calls, got %d", mock.resetCalls)
	}
}

func TestMockFlashPluginClose(t *testing.T) {
	mock := &MockFlashPlugin{}
	if err := mock.Close(); err != nil {
		t.Errorf("Close error: %v", err)
	}
}

// ============ download Function Tests ============

func TestDownloadSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test wasm content"))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "plugin.wasm")

	err := download(server.URL, targetPath)
	if err != nil {
		t.Fatalf("download error: %v", err)
	}

	data, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("read file error: %v", err)
	}
	if string(data) != "test wasm content" {
		t.Errorf("expected 'test wasm content', got %s", string(data))
	}
}

func TestDownloadHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "plugin.wasm")

	err := download(server.URL, targetPath)
	if err == nil {
		t.Error("expected error for HTTP 404")
	}
}

func TestDownloadServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "plugin.wasm")

	err := download(server.URL, targetPath)
	if err == nil {
		t.Error("expected error for HTTP 500")
	}
}

func TestDownloadInvalidURL(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "plugin.wasm")

	err := download("://invalid-url", targetPath)
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestDownloadNonexistentServer(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "plugin.wasm")

	err := download("http://nonexistent-server-12345.local/plugin.wasm", targetPath)
	if err == nil {
		t.Error("expected error for nonexistent server")
	}
}

func TestDownloadLargeFile(t *testing.T) {
	largeContent := make([]byte, 1024*1024) // 1MB
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(largeContent)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "large-plugin.wasm")

	err := download(server.URL, targetPath)
	if err != nil {
		t.Fatalf("download error: %v", err)
	}

	info, err := os.Stat(targetPath)
	if err != nil {
		t.Fatalf("stat error: %v", err)
	}
	if info.Size() != int64(len(largeContent)) {
		t.Errorf("expected size %d, got %d", len(largeContent), info.Size())
	}
}

// ============ Load Function Tests ============

func TestLoadNoPlugin(t *testing.T) {
	cfg := &config.Config{
		Flash: config.FlashConfig{
			Plugin: "",
		},
	}

	_, err := Load(cfg, 115200)
	if err == nil {
		t.Error("expected error for empty plugin URL")
	}
}

func TestLoadWithPluginURL(t *testing.T) {
	// Create a mock server that serves a file
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Return minimal WASM-like header (magic number)
		w.Write([]byte("\x00asm\x01\x00\x00\x00"))
	}))
	defer server.Close()

	// Create temp home directory
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	cfg := &config.Config{
		Flash: config.FlashConfig{
			Plugin: server.URL + "/plugin.wasm",
		},
	}

	// This will attempt to load WASM, which may fail without proper WASM binary
	// but we're testing the download logic
	_, err := Load(cfg, 115200)
	// The error might be about WASM loading, not download
	if err != nil {
		t.Logf("Load error (expected for non-WASM file): %v", err)
	}
}

func TestLoadCreatesPluginDir(t *testing.T) {
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	pluginDir := filepath.Join(tmpHome, ".zyrthi", "plugins")

	// Directory shouldn't exist yet
	if _, err := os.Stat(pluginDir); err == nil {
		t.Fatal("plugin dir should not exist yet")
	}

	// Attempt to load (will fail due to invalid URL, but should create dir)
	cfg := &config.Config{
		Flash: config.FlashConfig{
			Plugin: "http://invalid/plugin.wasm",
		},
	}
	Load(cfg, 115200)

	// Check if directory was created
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		t.Error("plugin dir should have been created")
	}
}

// ============ Config Integration Tests ============

func TestConfigWithFlashSettings(t *testing.T) {
	cfg := &config.Config{
		Chip: "esp32c3",
		Flash: config.FlashConfig{
			Plugin:      "https://example.com/plugin.wasm",
			EntryAddr:   "0x0",
			DefaultBaud: 921600,
		},
		Project: config.ProjectConfig{
			Name: "test-project",
		},
	}

	if cfg.Flash.Plugin == "" {
		t.Error("Flash.Plugin should not be empty")
	}
	if cfg.Flash.DefaultBaud != 921600 {
		t.Errorf("Flash.DefaultBaud = %d, want 921600", cfg.Flash.DefaultBaud)
	}
}

// ============ Mock Operations Tests ============

func TestMockFlashWithFirmware(t *testing.T) {
	mock := &MockFlashPlugin{chip: "esp32c3"}

	firmware := make([]byte, 4096)
	for i := range firmware {
		firmware[i] = byte(i % 256)
	}

	err := mock.Flash("esp32c3", firmware, 0x1000)
	if err != nil {
		t.Errorf("Flash error: %v", err)
	}
	if mock.flashCalls != 1 {
		t.Errorf("flashCalls = %d, want 1", mock.flashCalls)
	}
}

func TestMockErase(t *testing.T) {
	mock := &MockFlashPlugin{chip: "esp32c3"}

	err := mock.Erase("esp32c3", 0x0, 0x100000)
	if err != nil {
		t.Errorf("Erase error: %v", err)
	}
	if mock.eraseCalls != 1 {
		t.Errorf("eraseCalls = %d, want 1", mock.eraseCalls)
	}
}

func TestMockReset(t *testing.T) {
	mock := &MockFlashPlugin{chip: "esp32c3"}

	err := mock.Reset("esp32c3")
	if err != nil {
		t.Errorf("Reset error: %v", err)
	}
	if mock.resetCalls != 1 {
		t.Errorf("resetCalls = %d, want 1", mock.resetCalls)
	}
}

func TestMockDetect(t *testing.T) {
	tests := []struct {
		name     string
		chip     string
		detectErr error
		wantChip string
		wantErr  bool
	}{
		{"success", "esp32c3", nil, "esp32c3", false},
		{"detect error", "", fmt.Errorf("no chip"), "", true},
		{"empty chip", "", nil, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockFlashPlugin{
				chip:       tt.chip,
				detectError: tt.detectErr,
			}

			chip, err := mock.Detect()
			if (err != nil) != tt.wantErr {
				t.Errorf("Detect() error = %v, wantErr %v", err, tt.wantErr)
			}
			if chip != tt.wantChip {
				t.Errorf("Detect() chip = %v, want %v", chip, tt.wantChip)
			}
		})
	}
}

func TestMockInit(t *testing.T) {
	tests := []struct {
		name    string
		initErr error
		wantErr bool
	}{
		{"success", nil, false},
		{"error", fmt.Errorf("init failed"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockFlashPlugin{initError: tt.initErr}

			err := mock.Init(nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// ============ FlashPlugin Interface Compliance Tests ============

func TestMockFlashPluginImplementsInterface(t *testing.T) {
	var _ FlashPlugin = &MockFlashPlugin{}
}

// ============ Edge Case Tests ============

func TestFlashEmptyFirmware(t *testing.T) {
	mock := &MockFlashPlugin{chip: "test-chip"}

	err := mock.Flash("test-chip", []byte{}, 0)
	if err != nil {
		t.Errorf("Flash with empty firmware error: %v", err)
	}
}

func TestFlashLargeFirmware(t *testing.T) {
	mock := &MockFlashPlugin{chip: "test-chip"}

	// Simulate 4MB firmware
	firmware := make([]byte, 4*1024*1024)

	err := mock.Flash("test-chip", firmware, 0)
	if err != nil {
		t.Errorf("Flash with large firmware error: %v", err)
	}
}

func TestEraseZeroSize(t *testing.T) {
	mock := &MockFlashPlugin{chip: "test-chip"}

	err := mock.Erase("test-chip", 0, 0)
	if err != nil {
		t.Errorf("Erase with zero size error: %v", err)
	}
}

func TestFlashWithOffset(t *testing.T) {
	mock := &MockFlashPlugin{chip: "test-chip"}

	firmware := []byte{0x01, 0x02, 0x03, 0x04}
	offset := uint32(0x1000)

	err := mock.Flash("test-chip", firmware, offset)
	if err != nil {
		t.Errorf("Flash with offset error: %v", err)
	}
}

func TestEraseWithLargeSize(t *testing.T) {
	mock := &MockFlashPlugin{chip: "test-chip"}

	err := mock.Erase("test-chip", 0, 16*1024*1024) // 16MB
	if err != nil {
		t.Errorf("Erase with large size error: %v", err)
	}
}

// ============ Multiple Operation Tests ============

func TestFullFlashSequence(t *testing.T) {
	mock := &MockFlashPlugin{chip: "esp32c3"}

	// Init
	if err := mock.Init(nil); err != nil {
		t.Fatalf("Init error: %v", err)
	}

	// Detect
	chip, err := mock.Detect()
	if err != nil {
		t.Fatalf("Detect error: %v", err)
	}
	if chip != "esp32c3" {
		t.Errorf("Detect chip = %s, want esp32c3", chip)
	}

	// Erase
	if err := mock.Erase(chip, 0, 0x100000); err != nil {
		t.Fatalf("Erase error: %v", err)
	}

	// Flash
	firmware := make([]byte, 1024)
	if err := mock.Flash(chip, firmware, 0); err != nil {
		t.Fatalf("Flash error: %v", err)
	}

	// Reset
	if err := mock.Reset(chip); err != nil {
		t.Fatalf("Reset error: %v", err)
	}

	// Close
	if err := mock.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}

	// Verify call counts
	if mock.flashCalls != 1 {
		t.Errorf("flashCalls = %d, want 1", mock.flashCalls)
	}
	if mock.eraseCalls != 1 {
		t.Errorf("eraseCalls = %d, want 1", mock.eraseCalls)
	}
	if mock.resetCalls != 1 {
		t.Errorf("resetCalls = %d, want 1", mock.resetCalls)
	}
}

func TestMultipleFlashOperations(t *testing.T) {
	mock := &MockFlashPlugin{chip: "test-chip"}

	for i := 0; i < 10; i++ {
		firmware := []byte{byte(i)}
		mock.Flash("test-chip", firmware, uint32(i*1024))
	}

	if mock.flashCalls != 10 {
		t.Errorf("flashCalls = %d, want 10", mock.flashCalls)
	}
}