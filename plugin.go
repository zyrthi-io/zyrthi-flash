package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// FlashPlugin 烧录插件接口
type FlashPlugin interface {
	Init(api *HostAPI) error
	Detect() (string, error)
	Flash(chip string, firmware []byte, offset uint32) error
	Erase(chip string, offset uint32, size uint32) error
	Reset(chip string) error
	Close() error
}

// loadPlugin 加载 WASM 插件
func loadPlugin(cfg *Config, baud int) (FlashPlugin, error) {
	pluginURL := cfg.Flash.Plugin
	if pluginURL == "" {
		return nil, fmt.Errorf("未配置烧录插件")
	}

	// 确保 ~/.zyrthi/plugins 目录存在
	pluginDir := filepath.Join(os.Getenv("HOME"), ".zyrthi", "plugins")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return nil, fmt.Errorf("创建插件目录失败: %w", err)
	}

	// 从 URL 提取文件名
	pluginName := filepath.Base(pluginURL)
	pluginPath := filepath.Join(pluginDir, pluginName)

	// 检查本地缓存
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		// 下载插件
		fmt.Printf("下载插件: %s\n", pluginURL)
		if err := downloadPlugin(pluginURL, pluginPath); err != nil {
			return nil, fmt.Errorf("下载插件失败: %w", err)
		}
	}

	// 加载 WASM 模块
	ctx := context.Background()
	return newWasmPlugin(ctx, pluginPath, cfg, baud)
}

// downloadPlugin 下载插件
func downloadPlugin(url string, path string) error {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// 创建临时文件
	tmpPath := path + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// 写入文件
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		os.Remove(tmpPath)
		return err
	}

	// 重命名
	return os.Rename(tmpPath, path)
}