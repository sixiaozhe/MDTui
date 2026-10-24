package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gdamore/tcell/v2"
	"golang.org/x/term"
	"mdtui/internal/document"
	"mdtui/internal/runner"
	"mdtui/internal/scheduler"
	"mdtui/internal/tui"
)

var version = "0.1.0-dev"

func main() { os.Exit(run(os.Args[1:])) }
func run(args []string) (code int) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, "内部错误:", r)
			code = 1
		}
	}()
	var file, logFile string
	check, noColor, noMouse := false, false, false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--version":
			fmt.Println("mdtui", version)
			return 0
		case "--help", "-h":
			fmt.Println("mdtui [--no-color] [--no-mouse] [--log-file PATH] PAGE.md\nmdtui check PAGE.md\nmdtui --version\n页面必须来自可信本地来源。Tab 切换，Enter 选择，F2 动作，Z 放大，q 退出。支持鼠标单击定位、双击列表确认、滚轮滚动；--no-mouse 关闭鼠标接管。")
			return 0
		case "check":
			if file != "" || check {
				return usage()
			}
			check = true
		case "--no-color":
			noColor = true
		case "--no-mouse":
			noMouse = true
		case "--log-file":
			i++
			if i == len(args) {
				return usage()
			}
			logFile = args[i]
		default:
			if strings.HasPrefix(args[i], "-") || file != "" {
				return usage()
			}
			file = args[i]
		}
	}
	if file == "" {
		return usage()
	}
	abs, err := filepath.Abs(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if check {
		pages, err := document.Check(abs)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
		panels := 0
		for _, page := range pages {
			panels += len(page.Panels)
		}
		fmt.Printf("%s: 校验通过 (%d 页面, %d panels)\n", file, len(pages), panels)
		return 0
	}
	page, err := document.Load(abs)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Fprintln(os.Stderr, "需要可交互 TTY；仅校验请使用 mdtui check PAGE.md")
		return 3
	}
	engine := scheduler.New(page, runner.OS{}, time.Now)
	if logFile != "" {
		f, e := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			return 1
		}
		defer f.Close()
		engine.Log = log.New(f, "", log.LstdFlags).Printf
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	screen, err := tcell.NewScreen()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 3
	}
	defer screen.Fini()
	if err = screen.Init(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 3
	}
	defer engine.Close()
	ui := tui.New(screen, engine, noColor || os.Getenv("NO_COLOR") != "")
	ui.NoMouse = noMouse
	if err = ui.Run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
func usage() int {
	fmt.Fprintln(os.Stderr, "用法: mdtui [--no-color] [--no-mouse] [--log-file PATH] PAGE.md | mdtui check PAGE.md")
	return 2
}
