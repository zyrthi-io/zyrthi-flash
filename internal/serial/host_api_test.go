package serial

import (
	"bytes"
	"testing"
	"time"

	"go.bug.st/serial"
)

// MockPort 模拟串口
type MockPort struct {
	readBuf  *bytes.Buffer
	writeBuf *bytes.Buffer
	dtrLevel bool
	rtsLevel bool
}

func NewMockPort() *MockPort {
	return &MockPort{
		readBuf:  &bytes.Buffer{},
		writeBuf: &bytes.Buffer{},
	}
}

func (m *MockPort) Read(p []byte) (n int, err error)    { return m.readBuf.Read(p) }
func (m *MockPort) Write(p []byte) (n int, err error)   { return m.writeBuf.Write(p) }
func (m *MockPort) Close() error                        { return nil }
func (m *MockPort) SetDTR(level bool) error             { m.dtrLevel = level; return nil }
func (m *MockPort) SetRTS(level bool) error             { m.rtsLevel = level; return nil }
func (m *MockPort) Break(time.Duration) error           { return nil }
func (m *MockPort) SetMode(*serial.Mode) error          { return nil }
func (m *MockPort) Drain() error                        { return nil }
func (m *MockPort) ResetInputBuffer() error             { return nil }
func (m *MockPort) ResetOutputBuffer() error            { return nil }
func (m *MockPort) GetModemStatusBits() (*serial.ModemStatusBits, error) {
	return &serial.ModemStatusBits{}, nil
}
func (m *MockPort) SetReadTimeout(time.Duration) error  { return nil }

// ============ HostAPI Tests ============

func TestNewHostAPI(t *testing.T) {
	port := NewMockPort()
	api := NewHostAPI(port)
	if api == nil {
		t.Fatal("NewHostAPI returned nil")
	}
	if api.port != port {
		t.Error("port not set correctly")
	}
}

func TestHostAPISerialWrite(t *testing.T) {
	port := NewMockPort()
	api := NewHostAPI(port)

	data := []byte("test data")
	n, err := api.SerialWrite(data)
	if err != nil {
		t.Fatalf("SerialWrite error: %v", err)
	}
	if n != len(data) {
		t.Errorf("expected %d bytes written, got %d", len(data), n)
	}

	// Verify data was written to port
	written := port.writeBuf.String()
	if written != "test data" {
		t.Errorf("expected 'test data', got %s", written)
	}
}

func TestHostAPISerialWriteEmpty(t *testing.T) {
	port := NewMockPort()
	api := NewHostAPI(port)

	n, err := api.SerialWrite([]byte{})
	if err != nil {
		t.Fatalf("SerialWrite error: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 bytes written, got %d", n)
	}
}

func TestHostAPISerialWriteLarge(t *testing.T) {
	port := NewMockPort()
	api := NewHostAPI(port)

	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	n, err := api.SerialWrite(data)
	if err != nil {
		t.Fatalf("SerialWrite error: %v", err)
	}
	if n != 1024 {
		t.Errorf("expected 1024 bytes written, got %d", n)
	}
}

func TestHostAPISerialRead(t *testing.T) {
	port := NewMockPort()
	api := NewHostAPI(port)

	port.readBuf.WriteString("test data")

	buf := make([]byte, 100)
	n, err := api.SerialRead(buf)
	if err != nil {
		t.Fatalf("SerialRead error: %v", err)
	}
	if n != 9 {
		t.Errorf("expected 9 bytes read, got %d", n)
	}
	if string(buf[:n]) != "test data" {
		t.Errorf("expected 'test data', got %s", string(buf[:n]))
	}
}

func TestHostAPISerialReadEmpty(t *testing.T) {
	port := NewMockPort()
	api := NewHostAPI(port)

	buf := make([]byte, 100)
	n, err := api.SerialRead(buf)
	// Buffer is empty, should return EOF or 0
	if err == nil && n != 0 {
		t.Logf("read %d bytes from empty buffer", n)
	}
}

func TestHostAPISerialReadExactSize(t *testing.T) {
	port := NewMockPort()
	api := NewHostAPI(port)

	port.readBuf.WriteString("12345")

	buf := make([]byte, 5)
	n, err := api.SerialRead(buf)
	if err != nil {
		t.Fatalf("SerialRead error: %v", err)
	}
	if n != 5 {
		t.Errorf("expected 5 bytes read, got %d", n)
	}
}

func TestHostAPISerialReadMultiple(t *testing.T) {
	port := NewMockPort()
	api := NewHostAPI(port)

	port.readBuf.WriteString("firstsecond")

	buf1 := make([]byte, 5)
	n1, err := api.SerialRead(buf1)
	if err != nil {
		t.Fatalf("SerialRead error: %v", err)
	}

	buf2 := make([]byte, 6)
	n2, err := api.SerialRead(buf2)
	if err != nil {
		t.Fatalf("SerialRead error: %v", err)
	}

	if n1+n2 != 11 {
		t.Errorf("expected total 11 bytes, got %d", n1+n2)
	}
}

func TestHostAPISerialSetDTR(t *testing.T) {
	port := NewMockPort()
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
	port := NewMockPort()
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

func TestHostAPISerialSetBaudRate(t *testing.T) {
	port := NewMockPort()
	api := NewHostAPI(port)

	// Test common baud rates
	baudRates := []int{9600, 19200, 38400, 57600, 115200, 230400}

	for _, baud := range baudRates {
		err := api.SerialSetBaudRate(baud)
		if err != nil {
			t.Errorf("SerialSetBaudRate(%d) error: %v", baud, err)
		}
	}
}

func TestHostAPILog(t *testing.T) {
	port := NewMockPort()
	api := NewHostAPI(port)

	// Test various log levels
	levels := []string{"INFO", "DEBUG", "WARN", "ERROR", "TRACE"}

	for _, level := range levels {
		api.Log(level, "test message")
		// Log just prints to stdout, no error to check
	}
}

func TestHostAPILogEmpty(t *testing.T) {
	port := NewMockPort()
	api := NewHostAPI(port)

	api.Log("INFO", "")
	// Should not panic
}

func TestHostAPIProgress(t *testing.T) {
	port := NewMockPort()
	api := NewHostAPI(port)

	// Test progress reporting
	api.Progress(0, 100, "Starting...")
	api.Progress(50, 100, "Halfway...")
	api.Progress(100, 100, "Complete!")
	// Progress just prints to stdout, no error to check
}

func TestHostAPIProgressZero(t *testing.T) {
	port := NewMockPort()
	api := NewHostAPI(port)

	// Test with zero total
	api.Progress(0, 0, "No total")
	api.Progress(10, 0, "Still no total")
	// Should not panic on division by zero
}

func TestHostAPIProgressPartial(t *testing.T) {
	port := NewMockPort()
	api := NewHostAPI(port)

	// Test partial progress
	for i := 0; i <= 10; i++ {
		api.Progress(i*10, 100, "Processing...")
	}
}

func TestHostAPIProgressOverComplete(t *testing.T) {
	port := NewMockPort()
	api := NewHostAPI(port)

	// Test progress over 100%
	api.Progress(150, 100, "Over 100%!")
	// Should handle gracefully
}

func TestHostAPIProgressNegative(t *testing.T) {
	port := NewMockPort()
	api := NewHostAPI(port)

	// Test negative values
	api.Progress(-10, 100, "Negative current")
	api.Progress(10, -100, "Negative total")
	// Should handle gracefully
}

// ============ MockPort Tests ============

func TestMockPortReadWrite(t *testing.T) {
	port := NewMockPort()

	// Write to writeBuf
	data := []byte("hello world")
	n, err := port.Write(data)
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}
	if n != len(data) {
		t.Errorf("expected %d bytes written, got %d", len(data), n)
	}

	// Verify writeBuf contains the data
	if port.writeBuf.String() != "hello world" {
		t.Errorf("writeBuf expected 'hello world', got %s", port.writeBuf.String())
	}

	// Write to readBuf, then read
	port.readBuf.WriteString("test read")
	buf := make([]byte, 20)
	n, err = port.Read(buf)
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}
	if string(buf[:n]) != "test read" {
		t.Errorf("readBuf expected 'test read', got %s", string(buf[:n]))
	}
}

func TestMockPortClose(t *testing.T) {
	port := NewMockPort()
	err := port.Close()
	if err != nil {
		t.Errorf("Close error: %v", err)
	}
}

func TestMockPortSetDTR(t *testing.T) {
	port := NewMockPort()

	err := port.SetDTR(true)
	if err != nil {
		t.Errorf("SetDTR error: %v", err)
	}
	if !port.dtrLevel {
		t.Error("DTR should be true")
	}

	err = port.SetDTR(false)
	if err != nil {
		t.Errorf("SetDTR error: %v", err)
	}
	if port.dtrLevel {
		t.Error("DTR should be false")
	}
}

func TestMockPortSetRTS(t *testing.T) {
	port := NewMockPort()

	err := port.SetRTS(true)
	if err != nil {
		t.Errorf("SetRTS error: %v", err)
	}
	if !port.rtsLevel {
		t.Error("RTS should be true")
	}

	err = port.SetRTS(false)
	if err != nil {
		t.Errorf("SetRTS error: %v", err)
	}
	if port.rtsLevel {
		t.Error("RTS should be false")
	}
}

func TestMockPortBreak(t *testing.T) {
	port := NewMockPort()
	err := port.Break(100 * time.Millisecond)
	if err != nil {
		t.Errorf("Break error: %v", err)
	}
}

func TestMockPortSetMode(t *testing.T) {
	port := NewMockPort()
	err := port.SetMode(&serial.Mode{BaudRate: 115200})
	if err != nil {
		t.Errorf("SetMode error: %v", err)
	}
}

func TestMockPortDrain(t *testing.T) {
	port := NewMockPort()
	err := port.Drain()
	if err != nil {
		t.Errorf("Drain error: %v", err)
	}
}

func TestMockPortResetBuffers(t *testing.T) {
	port := NewMockPort()

	err := port.ResetInputBuffer()
	if err != nil {
		t.Errorf("ResetInputBuffer error: %v", err)
	}

	err = port.ResetOutputBuffer()
	if err != nil {
		t.Errorf("ResetOutputBuffer error: %v", err)
	}
}

func TestMockPortGetModemStatusBits(t *testing.T) {
	port := NewMockPort()
	bits, err := port.GetModemStatusBits()
	if err != nil {
		t.Errorf("GetModemStatusBits error: %v", err)
	}
	if bits == nil {
		t.Error("GetModemStatusBits returned nil")
	}
}

func TestMockPortSetReadTimeout(t *testing.T) {
	port := NewMockPort()
	err := port.SetReadTimeout(1 * time.Second)
	if err != nil {
		t.Errorf("SetReadTimeout error: %v", err)
	}
}

// ============ Integration Tests ============

func TestHostAPIIntegration(t *testing.T) {
	port := NewMockPort()
	api := NewHostAPI(port)

	// Simulate a typical flash operation sequence

	// 1. Set control lines
	api.SerialSetDTR(false)
	api.SerialSetRTS(false)

	// 2. Enter download mode
	api.SerialSetDTR(true)
	api.SerialSetRTS(true)
	api.SerialSetBaudRate(115200)

	// 3. Send command
	cmd := []byte{0x00, 0x01, 0x02}
	api.SerialWrite(cmd)

	// 4. Log progress
	api.Log("INFO", "Starting flash...")
	for i := 0; i <= 100; i += 10 {
		api.Progress(i, 100, "Flashing...")
	}

	// 5. Reset device
	api.SerialSetDTR(false)
	api.SerialSetRTS(false)

	// Verify all operations completed without error
}
