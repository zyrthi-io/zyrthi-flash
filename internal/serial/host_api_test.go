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

func TestNewHostAPI(t *testing.T) {
	port := NewMockPort()
	api := NewHostAPI(port)
	if api == nil {
		t.Fatal("NewHostAPI returned nil")
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
}