#!/usr/bin/env python3
"""Exercise the real Linux workbench page through a PTY."""
import fcntl
import json
import os
import pty
import select
import struct
import subprocess
import termios
import time
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]


def drain(fd, seconds):
    output = b""
    end = time.monotonic() + seconds
    while time.monotonic() < end:
        if select.select([fd], [], [], min(.03, end - time.monotonic()))[0]:
            output += os.read(fd, 65536)
    return output


def main():
    master, slave = pty.openpty()
    fcntl.ioctl(slave, termios.TIOCSWINSZ, struct.pack("HHHH", 40, 120, 0, 0))

    def setup():
        os.setsid()
        fcntl.ioctl(0, termios.TIOCSCTTY, 0)

    proc = subprocess.Popen(
        [str(ROOT / "bin/mdtui"), "--no-color", str(ROOT / "examples/linux-workbench.md")],
        stdin=slave, stdout=slave, stderr=slave, preexec_fn=setup,
        env={**os.environ, "TERM": "xterm-256color", "LANG": "C.UTF-8"},
    )
    try:
        output = drain(master, .8)
        assert proc.poll() is None, output.decode(errors="replace")
        assert b"MDTUI" in output
        os.write(master, b"\x1b[B\r")  # Select the network topic.
        output += drain(master, .8)
        assert b"ip" in output, "topic selection did not render network commands"
        os.write(master, b"d")  # Run the default loopback diagnosis.
        output += drain(master, 3)
        assert b"PING" in output, "diagnosis did not render"
        os.write(master, b"q")
        proc.wait(timeout=3)
        assert proc.returncode == 0, proc.returncode
        print(json.dumps({"page": "linux-workbench", "selection": True, "diagnosis": True, "exit": proc.returncode}))
    finally:
        if proc.poll() is None:
            proc.kill()
            proc.wait()
        os.close(master)
        os.close(slave)


if __name__ == "__main__":
    main()
