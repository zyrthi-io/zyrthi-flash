package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "zyrthi.yaml")

	content := `platform: esp32
chip: esp32c3
flash:
  plugin: https://example.com/plugin.wasm
  entry_addr: "0x0"
  flash_size: 4MB
  default_baud: 115200
project:
  name: test-project
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		t.Fatalf("loadConfig error: %v", err)
	}

	if cfg.Platform != "esp32" {
		t.Errorf("expected platform 'esp32', got %s", cfg.Platform)
	}
	if cfg.Chip != "esp32c3" {
		t.Errorf("expected chip 'esp32c3', got %s", cfg.Chip)
	}
	if cfg.Flash.Plugin != "https://example.com/plugin.wasm" {
		t.Errorf("expected plugin URL, got %s", cfg.Flash.Plugin)
	}
	if cfg.Project.Name != "test-project" {
		t.Errorf("expected project name 'test-project', got %s", cfg.Project.Name)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "zyrthi.yaml")

	// 最小配置
	content := `platform: esp32
chip: esp32c3
flash:
  plugin: https://example.com/plugin.wasm
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		t.Fatalf("loadConfig error: %v", err)
	}

	// 检查默认值
	if cfg.Project.Name != "firmware" {
		t.Errorf("expected default project name 'firmware', got %s", cfg.Project.Name)
	}
}

func TestLoadConfigNotExist(t *testing.T) {
	_, err := loadConfig("/nonexistent/zyrthi.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "zyrthi.yaml")

	content := `platform: esp32
chip: [invalid yaml
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := loadConfig(configPath)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestConfigStruct(t *testing.T) {
	cfg := Config{
		Platform: "esp32",
		Chip:     "esp32c3",
		Flash: FlashConfig{
			Plugin:      "https://example.com/plugin.wasm",
			EntryAddr:   "0x0",
			FlashSize:   "4MB",
			DefaultBaud: 115200,
		},
		Project: ProjectConfig{
			Name: "test-project",
		},
	}

	if cfg.Platform != "esp32" {
		t.Errorf("expected platform 'esp32', got %s", cfg.Platform)
	}
	if cfg.Flash.Plugin != "https://example.com/plugin.wasm" {
		t.Errorf("expected plugin URL, got %s", cfg.Flash.Plugin)
	}
	if cfg.Project.Name != "test-project" {
		t.Errorf("expected project name 'test-project', got %s", cfg.Project.Name)
	}
}

func TestFlashConfigStruct(t *testing.T) {
	fc := FlashConfig{
		Plugin:      "https://example.com/plugin.wasm",
		EntryAddr:   "0x0",
		FlashSize:   "4MB",
		DefaultBaud: 115200,
		MaxBaud:     921600,
	}

	if fc.Plugin != "https://example.com/plugin.wasm" {
		t.Errorf("expected plugin URL, got %s", fc.Plugin)
	}
	if fc.DefaultBaud != 115200 {
		t.Errorf("expected default baud 115200, got %d", fc.DefaultBaud)
	}
}

func TestProjectConfigStruct(t *testing.T) {
	pc := ProjectConfig{
		Name: "my-firmware",
	}

	if pc.Name != "my-firmware" {
		t.Errorf("expected name 'my-firmware', got %s", pc.Name)
	}
}

func TestDownloadPluginInvalidURL(t *testing.T) {
	tmpDir := t.TempDir()
	pluginPath := filepath.Join(tmpDir, "plugin.wasm")

	err := downloadPlugin("://invalid-url", pluginPath)
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}
