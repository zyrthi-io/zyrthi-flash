package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
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

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}

	if cfg.Platform != "esp32" {
		t.Errorf("expected platform 'esp32', got %s", cfg.Platform)
	}
	if cfg.Flash.Plugin != "https://example.com/plugin.wasm" {
		t.Errorf("expected plugin URL, got %s", cfg.Flash.Plugin)
	}
}

func TestLoadDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "zyrthi.yaml")

	content := `platform: esp32
chip: esp32c3
flash:
  plugin: https://example.com/plugin.wasm
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}

	if cfg.Project.Name != "firmware" {
		t.Errorf("expected default project name 'firmware', got %s", cfg.Project.Name)
	}
}

func TestLoadNotExist(t *testing.T) {
	_, err := Load("/nonexistent/zyrthi.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "zyrthi.yaml")

	content := `platform: esp32
chip: [invalid yaml
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}