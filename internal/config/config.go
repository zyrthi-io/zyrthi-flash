package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config zyrthi.yaml 配置结构
type Config struct {
	Platform string        `yaml:"platform"`
	Chip     string        `yaml:"chip"`
	Flash    FlashConfig   `yaml:"flash"`
	Project  ProjectConfig `yaml:"project"`
}

// FlashConfig 烧录配置
type FlashConfig struct {
	Plugin      string `yaml:"plugin"`
	EntryAddr   string `yaml:"entry_addr"`
	FlashSize   string `yaml:"flash_size"`
	DefaultBaud int    `yaml:"default_baud"`
	MaxBaud     int    `yaml:"max_baud"`
}

// ProjectConfig 项目配置
type ProjectConfig struct {
	Name string `yaml:"name"`
}

// Load 从 zyrthi.yaml 加载配置
func Load(path string) (*Config, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// 设置默认值
	if cfg.Project.Name == "" {
		cfg.Project.Name = "firmware"
	}

	return &cfg, nil
}
