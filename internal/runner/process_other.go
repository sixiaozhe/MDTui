//go:build !linux

package runner

import (
	"os/exec"
	"time"
)

func configure(c *exec.Cmd)      { c.WaitDelay = 100 * time.Millisecond }
func terminateGroup(c *exec.Cmd) { _ = c.Process.Kill() }
func killGroup(c *exec.Cmd)      { _ = c.Process.Kill() }
