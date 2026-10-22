//go:build linux

package runner

import (
	"context"
	"fmt"
	"mdtui/internal/content"
	"os"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestHelperProcess(t *testing.T) {
	if os.Getenv("MDTUI_HELPER") != "1" {
		return
	}
	for i, a := range os.Args {
		if a == "--" {
			switch os.Args[i+1] {
			case "echo":
				fmt.Print(strings.Join(os.Args[i+2:], "|"))
			case "huge":
				fmt.Print(strings.Repeat("x", (4<<20)+1))
			case "stderr":
				fmt.Fprint(os.Stderr, strings.Repeat("e", 70<<10))
				os.Exit(7)
			case "slow":
				for {
					time.Sleep(time.Second)
				}
			case "split":
				for _, b := range []byte("中文\x1b]52;c;unsafe\x07完成\n") {
					os.Stdout.Write([]byte{b})
					time.Sleep(time.Millisecond)
				}
			}
			os.Exit(0)
		}
	}
}
func helper(t *testing.T, mode string, args ...string) Spec {
	t.Helper()
	t.Setenv("MDTUI_HELPER", "1")
	a := []string{os.Args[0], "-test.run=TestHelperProcess", "--", mode}
	return Spec{Args: append(a, args...), Timeout: 2 * time.Second}
}
func TestLiteralArguments(t *testing.T) {
	value := `a b ' ";$(touch /tmp/MDTUI_SHOULD_NOT_EXIST)`
	r := OS{}.Run(context.Background(), helper(t, "echo", value), &content.Buffer{})
	if r.Err != nil || string(r.Stdout) != value {
		t.Fatal(r)
	}
}
func TestOutputLimitAndFailure(t *testing.T) {
	r := OS{}.Run(context.Background(), helper(t, "huge"), &content.Buffer{})
	if r.Err == nil || len(r.Stdout) != 4<<20 {
		t.Fatal(r.Err, len(r.Stdout))
	}
	r = OS{}.Run(context.Background(), helper(t, "stderr"), &content.Buffer{})
	if r.ExitCode != 7 || !strings.HasPrefix(r.Stderr, "[stderr 已截断]") {
		t.Fatal(r.ExitCode)
	}
}
func TestTimeout(t *testing.T) {
	s := helper(t, "slow")
	s.Timeout = 30 * time.Millisecond
	start := time.Now()
	r := OS{}.Run(context.Background(), s, &content.Buffer{})
	if r.Err == nil || time.Since(start) > 1500*time.Millisecond {
		t.Fatal(r.Err, time.Since(start))
	}
}
func TestProcessGroupCleanup(t *testing.T) {
	dir := t.TempDir()
	pidfile := dir + "/pid"
	script := dir + "/child.sh"
	os.WriteFile(script, []byte("#!/bin/sh\ntrap '' TERM\nsleep 100 &\necho $! > '"+pidfile+"'\nwait\n"), 0700)
	r := OS{}.Run(context.Background(), Spec{Args: []string{script}, Timeout: 100 * time.Millisecond}, &content.Buffer{})
	if r.Err == nil {
		t.Fatal("timeout missing")
	}
	b, err := os.ReadFile(pidfile)
	if err != nil {
		t.Fatal(err)
	}
	pid, _ := strconv.Atoi(strings.TrimSpace(string(b)))
	for until := time.Now().Add(time.Second); ; {
		err = syscall.Kill(pid, 0)
		if err != nil {
			break
		}
		stat, _ := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
		parts := strings.Fields(string(stat))
		if len(parts) >= 3 && parts[2] == "Z" {
			break
		}
		if time.Now().After(until) {
			t.Fatalf("descendant %d remains alive", pid)
		}
		time.Sleep(time.Millisecond)
	}
}
func TestRelativeExecutableAndWorkingDirectory(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(dir+"/fixture", []byte("#!/bin/sh\npwd\n"), 0700)
	r := OS{}.Run(context.Background(), Spec{Args: []string{"./fixture"}, Dir: dir, Timeout: time.Second}, &content.Buffer{})
	if r.Err != nil || strings.TrimSpace(string(r.Stdout)) != dir {
		t.Fatal(r)
	}
}

func TestExitedLeaderCleansDescendants(t *testing.T) {
	dir := t.TempDir()
	pidfile := dir + "/pid"
	script := dir + "/parent.sh"
	os.WriteFile(script, []byte("#!/bin/sh\nsleep 100 &\necho $! > '"+pidfile+"'\nexit 0\n"), 0700)
	OS{}.Run(context.Background(), Spec{Args: []string{script}, Timeout: time.Second}, &content.Buffer{})
	b, err := os.ReadFile(pidfile)
	if err != nil {
		t.Fatal(err)
	}
	pid, _ := strconv.Atoi(strings.TrimSpace(string(b)))
	for until := time.Now().Add(time.Second); ; {
		stat, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
		fields := strings.Fields(string(stat))
		if err != nil || (len(fields) > 2 && fields[2] == "Z") {
			break
		}
		if time.Now().After(until) {
			t.Fatal("live descendant after leader exit")
		}
		time.Sleep(time.Millisecond)
	}
}

func TestRealStreamSplitWrites(t *testing.T) {
	s := helper(t, "split")
	s.Stream = true
	b := &content.Buffer{}
	r := OS{}.Run(context.Background(), s, b)
	w, _ := b.Read(0)
	if r.Err != nil || w.Text != "中文完成\n" {
		t.Fatalf("%v %q", r.Err, w.Text)
	}
}
