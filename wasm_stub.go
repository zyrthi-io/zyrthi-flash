//go:build nowasm

package main

import (
	"context"
	"fmt"
)

// WasmPlugin WASM 插件存根（无 WASM 支持时使用）
type WasmPlugin struct{}

func newWasmPlugin(ctx context.Context, wasmPath string, cfg *Config, baud int) (*WasmPlugin, error) {
	return nil, fmt.Errorf("WASM 支持未编译，请使用 '-tags nowasm' 以外的构建标签")
}

func (p *WasmPlugin) Init(api *HostAPI) error {
	return fmt.Errorf("WASM 不支持")
}

func (p *WasmPlugin) Detect() (string, error) {
	return "", fmt.Errorf("WASM 不支持")
}

func (p *WasmPlugin) Flash(chip string, firmware []byte, offset uint32) error {
	return fmt.Errorf("WASM 不支持")
}

func (p *WasmPlugin) Erase(chip string, offset uint32, size uint32) error {
	return fmt.Errorf("WASM 不支持")
}

func (p *WasmPlugin) Reset(chip string) error {
	return fmt.Errorf("WASM 不支持")
}

func (p *WasmPlugin) Close() error {
	return nil
}
