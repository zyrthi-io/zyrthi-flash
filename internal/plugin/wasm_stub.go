//go:build nowasm

package plugin

import (
	"context"
	"fmt"

	"github.com/zyrthi-io/zyrthi-flash/internal/config"
	"github.com/zyrthi-io/zyrthi-flash/internal/serial"
)

// WasmPlugin WASM 插件存根
type WasmPlugin struct{}

func newWasmPlugin(ctx context.Context, wasmPath string, cfg *config.Config, baud int) (*WasmPlugin, error) {
	return nil, fmt.Errorf("WASM 支持未编译")
}

func (p *WasmPlugin) Init(api *serial.HostAPI) error {
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