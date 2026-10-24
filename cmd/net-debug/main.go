// net-debug is a deterministic fixture, not a real device adapter.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		os.Exit(2)
	}
	name := "eth0"
	device := "lab-a"
	for i := 2; i+1 < len(os.Args); i++ {
		if os.Args[i] == "--name" {
			name = os.Args[i+1]
		}
		if os.Args[i] == "--device" {
			device = os.Args[i+1]
		}
	}
	switch os.Args[1] {
	case "devices":
		fmt.Println(`{"version":1,"items":[{"id":"lab-a","label":"实验设备 A","description":"模拟路由器"},{"id":"lab-b","label":"实验设备 B","description":"模拟交换机"}]}`)
	case "device-interfaces":
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"version": 1, "items": []map[string]string{
			{"id": "eth0", "label": "以太网接口 0", "description": device + " · UP · 192.168.1.10"},
			{"id": "wlan0", "label": "无线接口", "description": device + " · DOWN"},
		}})
	case "list":
		fmt.Println(`{"version":1,"items":[{"id":"eth0","label":"以太网接口 0","description":"UP · 192.168.1.10"},{"id":"wlan0","label":"无线接口","description":"DOWN"}]}`)
	case "show":
		fmt.Printf("# 接口 %s\n\n**状态**: UP\n\n| 属性 | 值 |\n| --- | --- |\n| 名称 | %s |\n| MTU | 1500 |\n\n[适配器指南](../docs/adapter.md)\n", name, name)
	case "stats":
		fmt.Printf("# %s 计数器\n\n| RX bytes | TX bytes | errors |\n| --- | --- | --- |\n| 102400 | 51200 | 0 |\n", name)
	case "diagnose":
		for i := 1; ; i++ {
			fmt.Printf("[%04d] %s: 模拟诊断正常 ✓\n", i, name)
			time.Sleep(150 * time.Millisecond)
		}
	case "echo":
		b, _ := json.Marshal(os.Args[2:])
		fmt.Println(string(b))
	case "fail":
		fmt.Fprintln(os.Stderr, "模拟故障: unavailable")
		os.Exit(7)
	case "delay":
		n, _ := strconv.Atoi(os.Args[2])
		time.Sleep(time.Duration(n) * time.Millisecond)
		fmt.Println(strings.Join(os.Args[3:], " "))
	case "flood":
		for {
			fmt.Println(strings.Repeat("流式日志 ", 100))
		}
	default:
		fmt.Fprintln(os.Stderr, "unknown fixture command")
		os.Exit(2)
	}
}
