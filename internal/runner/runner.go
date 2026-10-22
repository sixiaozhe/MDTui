package runner

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"mdtui/internal/content"
)

type Spec struct {
	Args    []string
	Dir     string
	Timeout time.Duration
	Stream  bool
}
type Result struct {
	Stdout            []byte
	Stderr            string
	Err               error
	ExitCode          int
	Started, Finished time.Time
}
type Executor interface {
	Run(context.Context, Spec, *content.Buffer) Result
}
type OS struct{}
type limited struct {
	mu             sync.Mutex
	data           []byte
	max            int
	tail, overflow bool
	cancel         context.CancelFunc
}

func (b *limited) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	n := len(p)
	if len(b.data)+n > b.max {
		b.overflow = true
		if b.tail {
			if n >= b.max {
				b.data = append(b.data[:0], p[n-b.max:]...)
			} else {
				b.data = append(b.data[len(b.data)+n-b.max:], p...)
			}
		} else {
			left := b.max - len(b.data)
			b.data = append(b.data, p[:left]...)
			b.cancel()
		}
		return n, nil
	}
	b.data = append(b.data, p...)
	return n, nil
}
func (OS) Run(parent context.Context, s Spec, stream *content.Buffer) Result {
	result := Result{Started: time.Now(), ExitCode: -1}
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	if s.Timeout > 0 {
		var stop context.CancelFunc
		ctx, stop = context.WithTimeout(ctx, s.Timeout)
		defer stop()
	}
	stdout := &limited{max: 4 << 20, cancel: cancel}
	stderr := &limited{max: 64 << 10, tail: true}
	executable := s.Args[0]
	if strings.ContainsAny(executable, "/\\") && !filepath.IsAbs(executable) {
		executable = filepath.Join(s.Dir, executable)
	}
	cmd := exec.Command(executable, s.Args[1:]...)
	cmd.Dir = s.Dir
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	if s.Stream {
		cmd.Stdout = stream
	}
	cmd.Stdin = nil
	configure(cmd)
	if e := cmd.Start(); e != nil {
		result.Err = e
		result.Finished = time.Now()
		return result
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case result.Err = <-done:
		// Even a leader which exits early may leave descendants holding pipes.
		killGroup(cmd)
	case <-ctx.Done():
		terminateGroup(cmd)
		timer := time.NewTimer(time.Second)
		select {
		case <-done:
			if !timer.Stop() {
				<-timer.C
			}
			killGroup(cmd)
		case <-timer.C:
			killGroup(cmd)
			<-done
		}
		result.Err = ctx.Err()
	}
	if s.Stream {
		stream.Finish()
	}
	result.ExitCode = cmd.ProcessState.ExitCode()
	result.Stdout = stdout.data
	result.Stderr = content.Safe(string(stderr.data))
	if stderr.overflow {
		result.Stderr = "[stderr 已截断]\n" + result.Stderr
	}
	if stdout.overflow {
		result.Err = fmt.Errorf("stdout 超过 4 MiB，已终止")
	}
	result.Finished = time.Now()
	return result
}

var _ io.Writer = (*limited)(nil)
