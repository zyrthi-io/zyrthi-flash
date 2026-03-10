//go:build !nowasm

package plugin

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/zyrthi-io/zyrthi-flash/internal/config"
	"github.com/zyrthi-io/zyrthi-flash/internal/serial"
)

// WasmPlugin WASM 插件实现
type WasmPlugin struct {
	ctx     context.Context
	runtime wazero.Runtime
	module  api.Module
	cfg     *config.Config
	baud    int
	api     *serial.HostAPI
}

func newWasmPlugin(ctx context.Context, wasmPath string, cfg *config.Config, baud int) (*WasmPlugin, error) {
	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		return nil, fmt.Errorf("读取 WASM 文件失败: %w", err)
	}

	r := wazero.NewRuntime(ctx)

	_, err = r.NewHostModuleBuilder("env").
		NewFunctionBuilder().WithFunc(hostSerialWrite).Export("serial_write").
		NewFunctionBuilder().WithFunc(hostSerialRead).Export("serial_read").
		NewFunctionBuilder().WithFunc(hostSerialSetDTR).Export("serial_set_dtr").
		NewFunctionBuilder().WithFunc(hostSerialSetRTS).Export("serial_set_rts").
		NewFunctionBuilder().WithFunc(hostDelay).Export("delay").
		NewFunctionBuilder().WithFunc(hostLogInfo).Export("log_info").
		NewFunctionBuilder().WithFunc(hostProgress).Export("progress").
		Instantiate(ctx)
	if err != nil {
		r.Close(ctx)
		return nil, fmt.Errorf("实例化 Host API 失败: %w", err)
	}

	module, err := r.Instantiate(ctx, wasmBytes)
	if err != nil {
		r.Close(ctx)
		return nil, fmt.Errorf("加载 WASM 模块失败: %w", err)
	}

	return &WasmPlugin{
		ctx:     ctx,
		runtime: r,
		module:  module,
		cfg:     cfg,
		baud:    baud,
	}, nil
}

var currentPlugin *WasmPlugin

func (p *WasmPlugin) Init(api *serial.HostAPI) error {
	p.api = api
	currentPlugin = p

	initFn := p.module.ExportedFunction("init")
	if initFn == nil {
		return nil
	}
	_, err := initFn.Call(p.ctx)
	return err
}

func (p *WasmPlugin) Detect() (string, error) {
	fn := p.module.ExportedFunction("detect")
	if fn == nil {
		return "", fmt.Errorf("WASM 模块缺少 detect 函数")
	}

	result, err := fn.Call(p.ctx)
	if err != nil {
		return "", err
	}
	if result[0] != 0 {
		return "", fmt.Errorf("芯片检测失败")
	}

	return p.cfg.Chip, nil
}

func (p *WasmPlugin) Flash(chip string, firmware []byte, offset uint32) error {
	fn := p.module.ExportedFunction("flash")
	if fn == nil {
		return fmt.Errorf("WASM 模块缺少 flash 函数")
	}

	mem := p.module.Memory()
	firmwarePtr := uint32(0x1000)
	mem.Write(firmwarePtr, firmware)

	result, err := fn.Call(p.ctx, uint64(firmwarePtr), uint64(len(firmware)), uint64(offset))
	if err != nil {
		return err
	}
	if result[0] != 0 {
		return fmt.Errorf("烧录失败")
	}
	return nil
}

func (p *WasmPlugin) Erase(chip string, offset uint32, size uint32) error {
	fn := p.module.ExportedFunction("erase")
	if fn == nil {
		return fmt.Errorf("WASM 模块缺少 erase 函数")
	}

	result, err := fn.Call(p.ctx, uint64(offset), uint64(size))
	if err != nil {
		return err
	}
	if result[0] != 0 {
		return fmt.Errorf("擦除失败")
	}
	return nil
}

func (p *WasmPlugin) Reset(chip string) error {
	fn := p.module.ExportedFunction("reset")
	if fn == nil {
		p.api.SerialSetDTR(true)
		p.api.SerialSetRTS(true)
		time.Sleep(100 * time.Millisecond)
		p.api.SerialSetRTS(false)
		return nil
	}

	result, err := fn.Call(p.ctx)
	if err != nil {
		return err
	}
	if result[0] != 0 {
		return fmt.Errorf("复位失败")
	}
	return nil
}

func (p *WasmPlugin) Close() error {
	return p.runtime.Close(p.ctx)
}

// === Host API 实现 ===

func hostSerialWrite(ctx context.Context, m api.Module, ptr uint32, length uint32) uint32 {
	if currentPlugin == nil || currentPlugin.api == nil {
		return 0
	}
	buf, ok := m.Memory().Read(ptr, length)
	if !ok {
		return 0
	}
	n, _ := currentPlugin.api.SerialWrite(buf)
	return uint32(n)
}

func hostSerialRead(ctx context.Context, m api.Module, ptr uint32, length uint32) uint32 {
	if currentPlugin == nil || currentPlugin.api == nil {
		return 0
	}
	buf := make([]byte, length)
	n, _ := currentPlugin.api.SerialRead(buf)
	m.Memory().Write(ptr, buf[:n])
	return uint32(n)
}

func hostSerialSetDTR(ctx context.Context, level uint32) {
	if currentPlugin != nil && currentPlugin.api != nil {
		currentPlugin.api.SerialSetDTR(level != 0)
	}
}

func hostSerialSetRTS(ctx context.Context, level uint32) {
	if currentPlugin != nil && currentPlugin.api != nil {
		currentPlugin.api.SerialSetRTS(level != 0)
	}
}

func hostDelay(ctx context.Context, ms uint32) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

func hostLogInfo(ctx context.Context, m api.Module, msgPtr uint32, msgLen uint32) {
	if currentPlugin == nil {
		return
	}
	msg, ok := m.Memory().Read(msgPtr, msgLen)
	if !ok {
		return
	}
	currentPlugin.api.Log("info", string(msg))
}

func hostProgress(ctx context.Context, current uint32, total uint32) {
	if currentPlugin != nil && currentPlugin.api != nil {
		currentPlugin.api.Progress(int(current), int(total), "")
	}
}