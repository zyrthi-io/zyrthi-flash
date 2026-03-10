package plugin

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/zyrthi-io/zyrthi-flash/internal/config"
	"github.com/zyrthi-io/zyrthi-flash/internal/serial"
)

// FlashPlugin 烧录插件接口
type FlashPlugin interface {
	Init(api *serial.HostAPI) error
	Detect() (string, error)
	Flash(chip string, firmware []byte, offset uint32) error
	Erase(chip string, offset uint32, size uint32) error
	Reset(chip string) error
	Close() error
}

// Load 加载 WASM 插件
func Load(cfg *config.Config, baud int) (FlashPlugin, error) {
	pluginURL := cfg.Flash.Plugin
	if pluginURL == "" {
		return nil, fmt.Errorf("未配置烧录插件")
	}

	pluginDir := filepath.Join(os.Getenv("HOME"), ".zyrthi", "plugins")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return nil, fmt.Errorf("创建插件目录失败: %w", err)
	}

	pluginName := filepath.Base(pluginURL)
	pluginPath := filepath.Join(pluginDir, pluginName)

	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		fmt.Printf("下载插件: %s\n", pluginURL)
		if err := download(pluginURL, pluginPath); err != nil {
			return nil, fmt.Errorf("下载插件失败: %w", err)
		}
	}

	ctx := context.Background()
	return newWasmPlugin(ctx, pluginPath, cfg, baud)
}

// download 下载插件
func download(url string, path string) error {
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

	tmpPath := path + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		os.Remove(tmpPath)
		return err
	}

	return os.Rename(tmpPath, path)
}
