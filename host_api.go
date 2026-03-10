package main

import (
	"fmt"

	"go.bug.st/serial"
)

// HostAPI 宿主提供的 API
type HostAPI struct {
	port serial.Port
}

// NewHostAPI 创建 HostAPI
func NewHostAPI(port serial.Port) *HostAPI {
	return &HostAPI{port: port}
}

// SerialWrite 写入数据
func (h *HostAPI) SerialWrite(data []byte) (int, error) {
	return h.port.Write(data)
}

// SerialRead 读取数据
func (h *HostAPI) SerialRead(buf []byte) (int, error) {
	return h.port.Read(buf)
}

// SerialSetDTR 设置 DTR
func (h *HostAPI) SerialSetDTR(level bool) error {
	return h.port.SetDTR(level)
}

// SerialSetRTS 设置 RTS
func (h *HostAPI) SerialSetRTS(level bool) error {
	return h.port.SetRTS(level)
}

// SerialSetBaudRate 设置波特率
func (h *HostAPI) SerialSetBaudRate(baud int) error {
	return h.port.SetMode(&serial.Mode{BaudRate: baud})
}

// Log 日志输出
func (h *HostAPI) Log(level string, msg string) {
	fmt.Printf("[%s] %s\n", level, msg)
}

// Progress 进度报告
func (h *HostAPI) Progress(current int, total int, msg string) {
	percent := 0
	if total > 0 {
		percent = current * 100 / total
	}
	fmt.Printf("\r[%d%%] %s", percent, msg)
	if current >= total {
		fmt.Println()
	}
}
