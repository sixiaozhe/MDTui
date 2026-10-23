// Package linuxops provides a fixed, read-only catalog of Linux diagnostics.
// User values are passed as individual argv entries after validation, never
// interpolated into a shell command.
package linuxops

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"mdtui/internal/content"
)

type Executor interface {
	Run(context.Context, string, ...string) (string, error)
}

type OS struct{}

type boundedOutput struct {
	mu   sync.Mutex
	data bytes.Buffer
}

func (b *boundedOutput) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	const limit = 256 << 10
	if remaining := limit - b.data.Len(); remaining > 0 {
		b.data.Write(p[:min(len(p), remaining)])
	}
	return len(p), nil
}

func (OS) Run(ctx context.Context, program string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, program, args...)
	cmd.Stdin = nil
	cmd.Env = append(os.Environ(), "TERM=dumb", "NO_COLOR=1", "SYSTEMD_COLORS=0", "PAGER=cat")
	output := &boundedOutput{}
	cmd.Stdout, cmd.Stderr = output, output
	err := cmd.Run()
	return content.Safe(output.data.String()), err
}

type invocation struct {
	program string
	args    []string
}

type section struct {
	title    string
	attempts []invocation
	lines    int
}

func task(title, program string, args ...string) section {
	return section{title: title, attempts: []invocation{{program: program, args: args}}, lines: 35}
}

var topics = []struct{ id, label, description string }{
	{"system", "系统与资源", "内核、运行时间、内存和负载"},
	{"network", "网络与端口", "接口、路由和监听端口"},
	{"process", "进程与负载", "CPU、内存和进程列表"},
	{"storage", "磁盘与挂载", "块设备、文件系统和挂载点"},
	{"service", "服务状态", "运行和失败的服务"},
	{"logs", "系统日志", "近期系统事件与错误"},
}

func Topics() string {
	type item struct {
		ID          string `json:"id"`
		Label       string `json:"label"`
		Description string `json:"description"`
	}
	items := make([]item, 0, len(topics))
	for _, topic := range topics {
		items = append(items, item{topic.id, topic.label, topic.description})
	}
	data, _ := json.Marshal(struct {
		Version int    `json:"version"`
		Items   []item `json:"items"`
	}{1, items})
	return string(data)
}

func Explain(id string) (string, error) {
	if !validTopic(id) {
		return "", fmt.Errorf("未知类别 %q", id)
	}
	title, sections := topicSections(id)
	var out strings.Builder
	fmt.Fprintf(&out, "# %s · 采集计划\n\n按 Enter 后执行下列命令。\n\n| 步骤 | 首选命令 | 备选命令 |\n| --- | --- | --- |\n", title)
	for _, section := range sections {
		primary, fallback := "—", "—"
		if len(section.attempts) > 0 {
			primary = commandText(section.attempts[0])
		}
		if len(section.attempts) > 1 {
			var alternatives []string
			for _, attempt := range section.attempts[1:] {
				alternatives = append(alternatives, commandText(attempt))
			}
			fallback = strings.Join(alternatives, "；")
		}
		fmt.Fprintf(&out, "| %s | `%s` | `%s` |\n", section.title, primary, fallback)
	}
	return out.String(), nil
}

func commandText(attempt invocation) string {
	return strings.Join(append([]string{attempt.program}, attempt.args...), " ")
}

func validTopic(id string) bool {
	for _, topic := range topics {
		if topic.id == id {
			return true
		}
	}
	return false
}

var serviceName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.@-]{0,127}$`)
var hostLabel = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9-]{0,61}[A-Za-z0-9])?$`)

func ValidTarget(value string) bool {
	if net.ParseIP(value) != nil {
		return true
	}
	if len(value) == 0 || len(value) > 253 || strings.HasSuffix(value, ".") {
		return false
	}
	for _, label := range strings.Split(value, ".") {
		if !hostLabel.MatchString(label) {
			return false
		}
	}
	return true
}

func Render(ctx context.Context, executor Executor, operation, value string) (string, error) {
	var sections []section
	title := ""
	switch operation {
	case "overview":
		title = "系统概览"
		sections = []section{
			task("内核", "uname", "-a"),
			task("运行时间与负载", "uptime"),
			task("内存", "free", "-h"),
			task("文件系统", "df", "-h"),
		}
	case "topic":
		if !validTopic(value) {
			return "", fmt.Errorf("未知类别 %q", value)
		}
		title, sections = topicSections(value)
	case "diagnose":
		if !ValidTarget(value) {
			return "", errors.New("目标只能是有效的主机名或 IP 地址")
		}
		title = "连通性诊断 · " + value
		sections = []section{
			task("名称解析", "getent", "ahosts", value),
			{title: "ICMP 连通性", attempts: []invocation{{"ping", []string{"-c", "3", "-W", "2", value}}, {"ping", []string{"-c", "3", value}}}, lines: 25},
		}
	case "service":
		if !serviceName.MatchString(value) {
			return "", errors.New("服务名只能包含字母、数字、点、下划线、@ 和连字符，且不能以符号开头")
		}
		title = "服务详情 · " + value
		sections = []section{{title: "服务状态", attempts: []invocation{{"systemctl", []string{"status", "--no-pager", "--full", value}}, {"service", []string{value, "status"}}}, lines: 65}}
	default:
		return "", fmt.Errorf("未知操作 %q", operation)
	}
	var out strings.Builder
	if operation == "overview" {
		host, _ := os.Hostname()
		fmt.Fprintf(&out, "# %s\n\n| 项目 | 当前值 |\n| :--- | :--- |\n| 主机 | %s |\n| 架构 | %s/%s |\n| 采集时间 | %s |\n\n", title, host, runtime.GOOS, runtime.GOARCH, time.Now().Format("2006-01-02 15:04:05"))
	} else {
		fmt.Fprintf(&out, "# %s\n\n", title)
	}
	for _, section := range sections {
		writeSection(ctx, &out, executor, section)
	}
	return out.String(), nil
}

func topicSections(id string) (string, []section) {
	switch id {
	case "system":
		return "系统与资源", []section{task("内核", "uname", "-a"), task("运行时间", "uptime"), task("内存", "free", "-h"), task("进程概览", "ps", "-eo", "pid,ppid,comm,%cpu,%mem", "--sort=-%cpu")}
	case "network":
		return "网络与端口", []section{
			{title: "接口与地址", attempts: []invocation{{"ip", []string{"-br", "addr"}}, {"ip", []string{"addr"}}, {"ifconfig", []string{"-a"}}}, lines: 60},
			{title: "路由", attempts: []invocation{{"ip", []string{"route"}}, {"route", []string{"-n"}}}, lines: 35},
			{title: "监听端口", attempts: []invocation{{"ss", []string{"-tuln"}}, {"netstat", []string{"-tuln"}}}, lines: 55},
		}
	case "process":
		return "进程与负载", []section{
			task("负载", "uptime"),
			{title: "进程列表（优先显示高 CPU）", attempts: []invocation{{"ps", []string{"-eo", "pid,ppid,user,%cpu,%mem,comm", "--sort=-%cpu"}}, {"ps", []string{"aux"}}, {"ps", nil}}, lines: 70},
		}
	case "storage":
		return "磁盘与挂载", []section{
			{title: "块设备", attempts: []invocation{{"lsblk", []string{"-o", "NAME,SIZE,FSTYPE,MOUNTPOINT"}}, {"lsblk", nil}}, lines: 55},
			{title: "文件系统容量", attempts: []invocation{{"df", []string{"-hT"}}, {"df", []string{"-h"}}}, lines: 55},
			{title: "挂载点", attempts: []invocation{{"findmnt", []string{"-r"}}, {"mount", nil}}, lines: 55},
		}
	case "service":
		return "服务状态", []section{
			{title: "正在运行", attempts: []invocation{{"systemctl", []string{"list-units", "--type=service", "--state=running", "--no-pager", "--plain", "--no-legend"}}, {"service", []string{"--status-all"}}}, lines: 65},
			{title: "失败项", attempts: []invocation{{"systemctl", []string{"--failed", "--no-pager", "--plain"}}}, lines: 35},
		}
	case "logs":
		return "系统日志", []section{{title: "最近事件", attempts: []invocation{{"journalctl", []string{"-n", "80", "--no-pager", "-o", "short-iso"}}, {"dmesg", []string{"-T"}}, {"tail", []string{"-n", "80", "/var/log/syslog"}}}, lines: 85}}
	}
	return "", nil
}

func writeSection(parent context.Context, out *strings.Builder, executor Executor, section section) {
	fmt.Fprintf(out, "## %s\n\n", section.title)
	var lastError error
	var lastOutput string
	var lastProgram string
	for _, attempt := range section.attempts {
		timeout := 4 * time.Second
		if attempt.program == "ping" {
			timeout = 9 * time.Second
		}
		ctx, cancel := context.WithTimeout(parent, timeout)
		value, err := executor.Run(ctx, attempt.program, attempt.args...)
		cancel()
		if err != nil {
			lastError = err
			if strings.TrimSpace(value) != "" {
				lastOutput = value
				lastProgram = attempt.program
			}
			continue
		}
		fmt.Fprintf(out, "`%s`\n\n", attempt.program)
		lines := strings.Split(strings.TrimSpace(value), "\n")
		if len(lines) == 1 && lines[0] == "" {
			out.WriteString("_命令未返回内容。_\n\n")
			return
		}
		for i, line := range lines {
			if i >= section.lines {
				out.WriteString("    … 后续输出已省略\n")
				break
			}
			out.WriteString("    " + line + "\n")
		}
		out.WriteByte('\n')
		return
	}
	if lastOutput != "" {
		fmt.Fprintf(out, "`%s` 返回非零状态：%s\n\n", lastProgram, content.Safe(lastError.Error()))
		for i, line := range strings.Split(strings.TrimSpace(lastOutput), "\n") {
			if i >= section.lines {
				out.WriteString("    … 后续输出已省略\n")
				break
			}
			out.WriteString("    " + line + "\n")
		}
		out.WriteByte('\n')
		return
	}
	if lastError != nil {
		fmt.Fprintf(out, "> 当前环境暂不可用：%s。请检查对应工具或权限。\n\n", content.Safe(lastError.Error()))
	} else {
		out.WriteString("> 当前环境没有可用命令。\n\n")
	}
}

func Follow(ctx context.Context, stdout, stderr io.Writer) error {
	commands := []invocation{
		{"journalctl", []string{"-f", "-n", "0", "--no-pager"}},
		{"tail", []string{"-F", "/var/log/syslog"}},
	}
	for _, command := range commands {
		if _, err := exec.LookPath(command.program); err != nil {
			continue
		}
		cmd := exec.CommandContext(ctx, command.program, command.args...)
		cmd.Stdout, cmd.Stderr = stdout, stderr
		cmd.Stdin = nil
		return cmd.Run()
	}
	return errors.New("未找到 journalctl 或 tail，无法跟随系统日志")
}
