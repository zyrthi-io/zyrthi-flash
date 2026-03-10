package main

import (
	"bytes"
	"testing"
	"time"

	"go.bug.st/serial"
)

// MockSerialPort 模拟串口
type MockSerialPort struct {
	readBuf  *bytes.Buffer
	writeBuf *bytes.Buffer
	dtrLevel bool
	rtsLevel bool
}

func NewMockSerialPort() *MockSerialPort {
	return &MockSerialPort{
		readBuf:  &bytes.Buffer{},
		writeBuf: &bytes.Buffer{},
	}
}

func (m *MockSerialPort) Read(p []byte) (n int, err error) {
	return m.readBuf.Read(p)
}

func (m *MockSerialPort) Write(p []byte) (n int, err error) {
	return m.writeBuf.Write(p)
}

func (m *MockSerialPort) Close() error {
	return nil
}

func (m *MockSerialPort) SetDTR(level bool) error {
	m.dtrLevel = level
	return nil
}

func (m *MockSerialPort) SetRTS(level bool) error {
	m.rtsLevel = level
	return nil
}

func (m *MockSerialPort) Break(breakDuration time.Duration) error {
	return nil
}

func (m *MockSerialPort) SetMode(mode *serial.Mode) error {
	return nil
}

func (m *MockSerialPort) Drain() error {
	return nil
}

func (m *MockSerialPort) ResetInputBuffer() error {
	return nil
}

func (m *MockSerialPort) ResetOutputBuffer() error {
	return nil
}

func (m *MockSerialPort) GetModemStatusBits() (*serial.ModemStatusBits, error) {
	return &serial.ModemStatusBits{}, nil
}

func (m *MockSerialPort) SetReadTimeout(t time.Duration) error {
	return nil
}

func TestNewHostAPI(t *testing.T) {
	port := NewMockSerialPort()
	api := NewHostAPI(port)
	if api == nil {
		t.Fatal("NewHostAPI returned nil")
	}
}

func TestHostAPISerialWrite(t *testing.T) {
	port := NewMockSerialPort()
	api := NewHostAPI(port)

	data := []byte("test data")
	n, err := api.SerialWrite(data)
	if err != nil {
		t.Fatalf("SerialWrite error: %v", err)
	}
	if n != len(data) {
		t.Errorf("expected %d bytes written, got %d", len(data), n)
	}

	// 验证数据被写入
	if port.writeBuf.String() != string(data) {
		t.Errorf("expected '%s', got '%s'", string(data), port.writeBuf.String())
	}
}

func TestHostAPISerialRead(t *testing.T) {
	port := NewMockSerialPort()
	api := NewHostAPI(port)

	// 先写入数据到 readBuf
	port.readBuf.WriteString("test data")

	buf := make([]byte, 100)
	n, err := api.SerialRead(buf)
	if err != nil {
		t.Fatalf("SerialRead error: %v", err)
	}
	if n != 9 {
		t.Errorf("expected 9 bytes read, got %d", n)
	}
}

func TestHostAPISerialSetDTR(t *testing.T) {
	port := NewMockSerialPort()
	api := NewHostAPI(port)

	err := api.SerialSetDTR(true)
	if err != nil {
		t.Fatalf("SerialSetDTR error: %v", err)
	}
	if !port.dtrLevel {
		t.Error("expected DTR to be true")
	}

	err = api.SerialSetDTR(false)
	if err != nil {
		t.Fatalf("SerialSetDTR error: %v", err)
	}
	if port.dtrLevel {
		t.Error("expected DTR to be false")
	}
}

func TestHostAPISerialSetRTS(t *testing.T) {
	port := NewMockSerialPort()
	api := NewHostAPI(port)

	err := api.SerialSetRTS(true)
	if err != nil {
		t.Fatalf("SerialSetRTS error: %v", err)
	}
	if !port.rtsLevel {
		t.Error("expected RTS to be true")
	}

	err = api.SerialSetRTS(false)
	if err != nil {
		t.Fatalf("SerialSetRTS error: %v", err)
	}
	if port.rtsLevel {
		t.Error("expected RTS to be false")
	}
}

func TestHostAPILog(t *testing.T) {
	port := NewMockSerialPort()
	api := NewHostAPI(port)

	// Log 方法只是打印，没有返回值
	api.Log("info", "test message")
	// 无错误即为成功
}

func TestHostAPIProgress(t *testing.T) {
	port := NewMockSerialPort()
	api := NewHostAPI(port)

	// Progress 方法只是打印进度
	api.Progress(50, 100, "testing")
	api.Progress(100, 100, "done")
	// 无错误即为成功
}

func TestHostAPIProgressZeroTotal(t *testing.T) {
	port := NewMockSerialPort()
	api := NewHostAPI(port)

	// 总数为 0 时不应该 panic
	api.Progress(0, 0, "testing")
	// 无错误即为成功
}
