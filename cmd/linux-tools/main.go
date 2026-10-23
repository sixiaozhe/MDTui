package main

import (
	"context"
	"fmt"
	"os"

	"mdtui/internal/linuxops"
)

func main() {
	args := os.Args[1:]
	if len(args) == 1 && args[0] == "topics" {
		fmt.Println(linuxops.Topics())
		return
	}
	if len(args) == 1 && args[0] == "follow" {
		if err := linuxops.Follow(context.Background(), os.Stdout, os.Stderr); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if len(args) != 2 || (args[0] != "overview" && args[0] != "topic" && args[0] != "diagnose" && args[0] != "service" && args[0] != "explain") {
		fmt.Fprintln(os.Stderr, "用法: linux-tools topics | overview - | topic 类别 | explain 类别 | diagnose 主机 | service 服务 | follow")
		os.Exit(2)
	}
	var result string
	var err error
	if args[0] == "explain" {
		result, err = linuxops.Explain(args[1])
	} else {
		result, err = linuxops.Render(context.Background(), linuxops.OS{}, args[0], args[1])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(result)
}
