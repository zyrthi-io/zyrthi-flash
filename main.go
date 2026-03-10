package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"go.bug.st/serial"
)

var (
	flagConfig   = flag.String("config", "zyrthi.yaml", "配置文件路径")
	flagPort     = flag.String("port", "", "串口设备")
	flagBaud     = flag.Int("baud", 0, "波特率")
	flagFirmware = flag.String("firmware", "", "固件文件路径")
	flagErase    = flag.Bool("erase", false, "烧录前全片擦除")
	flagVerify   = flag.Bool("verify", false, "烧录后校验")
)

func main() {
	flag.Parse()

	// 读取配置
	cfg, err := loadConfig(*flagConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法读取配置文件: %v\n", err)
		os.Exit(1)
	}

	// 确定波特率
	baud := *flagBaud
	if baud == 0 {
		baud = cfg.Flash.DefaultBaud
		if baud == 0 {
			baud = 115200
		}
	}

	// 确定串口
	port := *flagPort
	if port == "" {
		ports, err := serial.GetPortsList()
		if err != nil || len(ports) == 0 {
			fmt.Fprintln(os.Stderr, "错误: 未找到串口设备")
			os.Exit(1)
		}
		port = ports[0]
		fmt.Printf("自动选择串口: %s\n", port)
	}

	// 确定固件文件
	firmware := *flagFirmware
	if firmware == "" {
		// 默认使用 build/firmware.bin
		firmware = filepath.Join("build", cfg.Project.Name+".bin")
	}

	if _, err := os.Stat(firmware); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 固件文件不存在: %s\n", firmware)
		os.Exit(1)
	}

	// 加载插件
	plugin, err := loadPlugin(cfg, baud)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 加载插件失败: %v\n", err)
		os.Exit(1)
	}
	defer plugin.Close()

	// 打开串口
	serialPort, err := serial.Open(port, &serial.Mode{
		BaudRate: baud,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法打开串口 %s: %v\n", port, err)
		os.Exit(1)
	}
	defer serialPort.Close()

	fmt.Printf("已连接 %s @ %d baud\n", port, baud)

	// 注入 HostAPI
	hostAPI := NewHostAPI(serialPort)
	if err := plugin.Init(hostAPI); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 初始化插件失败: %v\n", err)
		os.Exit(1)
	}

	// 检测芯片
	chip, err := plugin.Detect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 检测芯片失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("检测到芯片: %s\n", chip)

	// 擦除
	if *flagErase {
		fmt.Println("擦除中...")
		if err := plugin.Erase(chip, 0, 0); err != nil {
			fmt.Fprintf(os.Stderr, "错误: 擦除失败: %v\n", err)
			os.Exit(1)
		}
	}

	// 读取固件
	firmwareData, err := os.ReadFile(firmware)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 读取固件失败: %v\n", err)
		os.Exit(1)
	}

	// 烧录
	fmt.Printf("烧录 %s (%d 字节)...\n", firmware, len(firmwareData))
	if err := plugin.Flash(chip, firmwareData, 0); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 烧录失败: %v\n", err)
		os.Exit(1)
	}

	// 校验
	if *flagVerify {
		fmt.Println("校验中...")
		// TODO: 实现校验
	}

	// 复位
	fmt.Println("复位设备...")
	if err := plugin.Reset(chip); err != nil {
		fmt.Fprintf(os.Stderr, "警告: 复位失败: %v\n", err)
	}

	fmt.Println("烧录完成")
}
