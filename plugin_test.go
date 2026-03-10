package main

import (
	"fmt"
	"testing"
)

// TestFlashPluginInterface 测试 FlashPlugin 接口定义
func TestFlashPluginInterface(t *testing.T) {
	// 创建一个简单的 mock 插件实现
	mock := &MockFlashPlugin{
		initError:   nil,
		detectError: nil,
		flashError:  nil,
		eraseError:  nil,
		resetError:  nil,
		chip:        "test-chip",
	}

	// 测试 Init
	err := mock.Init(nil)
	if err != nil {
		t.Errorf("Init error: %v", err)
	}

	// 测试 Detect
	chip, err := mock.Detect()
	if err != nil {
		t.Errorf("Detect error: %v", err)
	}
	if chip != "test-chip" {
		t.Errorf("expected 'test-chip', got %s", chip)
	}

	// 测试 Flash
	err = mock.Flash("test-chip", []byte{0x00, 0x01}, 0)
	if err != nil {
		t.Errorf("Flash error: %v", err)
	}

	// 测试 Erase
	err = mock.Erase("test-chip", 0, 4096)
	if err != nil {
		t.Errorf("Erase error: %v", err)
	}

	// 测试 Reset
	err = mock.Reset("test-chip")
	if err != nil {
		t.Errorf("Reset error: %v", err)
	}
}

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

func (m *MockFlashPlugin) Init(api *HostAPI) error {
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

// TestMockFlashPluginErrors 测试错误处理
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

// TestMockFlashPluginCallCount 测试调用计数
func TestMockFlashPluginCallCount(t *testing.T) {
	mock := &MockFlashPlugin{
		chip: "test-chip",
	}

	// 多次调用
	for i := 0; i < 3; i++ {
		mock.Flash("test-chip", []byte{0x00}, 0)
	}
	if mock.flashCalls != 3 {
		t.Errorf("expected 3 flash calls, got %d", mock.flashCalls)
	}

	for i := 0; i < 2; i++ {
		mock.Erase("test-chip", 0, 4096)
	}
	if mock.eraseCalls != 2 {
		t.Errorf("expected 2 erase calls, got %d", mock.eraseCalls)
	}

	mock.Reset("test-chip")
	if mock.resetCalls != 1 {
		t.Errorf("expected 1 reset call, got %d", mock.resetCalls)
	}
}

// TestHostAPISerialWriteLargeData 测试大数据写入
func TestHostAPISerialWriteLargeData(t *testing.T) {
	port := NewMockSerialPort()
	hostAPI := NewHostAPI(port)

	// 写入 1KB 数据
	largeData := make([]byte, 1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	n, err := hostAPI.SerialWrite(largeData)
	if err != nil {
		t.Fatalf("SerialWrite error: %v", err)
	}
	if n != len(largeData) {
		t.Errorf("expected %d bytes written, got %d", len(largeData), n)
	}
}

// TestHostAPISerialReadLargeData 测试大数据读取
func TestHostAPISerialReadLargeData(t *testing.T) {
	port := NewMockSerialPort()
	hostAPI := NewHostAPI(port)

	// 准备 1KB 数据
	largeData := make([]byte, 1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}
	port.readBuf.Write(largeData)

	readBuf := make([]byte, 2048)
	n, err := hostAPI.SerialRead(readBuf)
	if err != nil {
		t.Fatalf("SerialRead error: %v", err)
	}
	if n != 1024 {
		t.Errorf("expected 1024 bytes read, got %d", n)
	}
}

// TestHostAPISerialSetBaudRate 测试波特率设置
func TestHostAPISerialSetBaudRate(t *testing.T) {
	port := NewMockSerialPort()
	hostAPI := NewHostAPI(port)

	err := hostAPI.SerialSetBaudRate(921600)
	if err != nil {
		t.Errorf("SerialSetBaudRate error: %v", err)
	}
}

// TestFlashPluginWithFirmware 测试烧录固件数据
func TestFlashPluginWithFirmware(t *testing.T) {
	mock := &MockFlashPlugin{
		chip: "esp32c3",
	}

	// 模拟固件数据
	firmware := make([]byte, 4096)
	for i := range firmware {
		firmware[i] = byte(i)
	}

	// 烧录
	err := mock.Flash("esp32c3", firmware, 0x1000)
	if err != nil {
		t.Errorf("Flash error: %v", err)
	}
	if mock.flashCalls != 1 {
		t.Errorf("expected 1 flash call, got %d", mock.flashCalls)
	}
}

// TestEraseChip 全片擦除测试
func TestEraseChip(t *testing.T) {
	mock := &MockFlashPlugin{
		chip: "esp32",
	}

	// 全片擦除 (offset=0, size=0 表示全片)
	err := mock.Erase("esp32", 0, 0)
	if err != nil {
		t.Errorf("Erase error: %v", err)
	}

	// 部分擦除
	err = mock.Erase("esp32", 0x10000, 0x10000)
	if err != nil {
		t.Errorf("Erase error: %v", err)
	}

	if mock.eraseCalls != 2 {
		t.Errorf("expected 2 erase calls, got %d", mock.eraseCalls)
	}
}

// TestResetChip 复位测试
func TestResetChip(t *testing.T) {
	mock := &MockFlashPlugin{
		chip: "esp32c3",
	}

	err := mock.Reset("esp32c3")
	if err != nil {
		t.Errorf("Reset error: %v", err)
	}

	// 使用 HostAPI 的复位逻辑（DTR/RTS）
	port := NewMockSerialPort()
	api := NewHostAPI(port)

	// 模拟复位序列
	api.SerialSetDTR(false)
	api.SerialSetRTS(true)
	api.SerialSetRTS(false)

	if port.rtsLevel {
		t.Error("expected RTS to be false after reset sequence")
	}
}