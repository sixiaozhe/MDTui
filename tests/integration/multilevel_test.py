#!/usr/bin/env python3
"""Real PTY acceptance for the three-level device navigation example."""
import fcntl
import json
import os
import pty
import struct
import subprocess
import tempfile
import termios
from pathlib import Path
from pty_test import drain

ROOT = Path(__file__).resolve().parents[2]

def scenario():
    with tempfile.TemporaryDirectory(prefix='mdtui-multilevel-') as directory:
        directory = Path(directory)
        master, slave = pty.openpty()
        fcntl.ioctl(slave, termios.TIOCSWINSZ, struct.pack('HHHH', 30, 120, 0, 0))
        before = termios.tcgetattr(slave)
        def setup():
            os.setsid()
            fcntl.ioctl(0, termios.TIOCSCTTY, 0)
        process = subprocess.Popen(
            [str(ROOT/'bin/mdtui'), '--no-color', str(ROOT/'examples/multilevel.md')],
            cwd=directory, stdin=slave, stdout=slave, stderr=slave,
            env={**os.environ, 'TERM':'xterm-256color', 'LANG':'C.UTF-8'}, preexec_fn=setup)
        def send(keys, seconds=.3):
            os.write(master, keys)
            return drain(master, seconds)
        def report():
            existing = set(directory.glob('mdtui-report-*.md'))
            send(b'\x1bOS')
            created = set(directory.glob('mdtui-report-*.md')) - existing
            assert len(created) == 1, created
            text = created.pop().read_text()
            send(b'\r', .1)
            return text
        try:
            drain(master, .5)
            send(b'\x1bORlab-b\r')
            send(b'\r', .5)
            send(b'\x1bORwlan0\r')
            send(b'\r', .5)
            send(b'd', .5)
            leaf = report()
            assert '设备导航 → 接口列表 [lab-b] → 接口诊断' in leaf, leaf
            assert '父页对象: wlan0' in leaf and '模拟诊断正常' in leaf, leaf
            send(b'q')
            middle = report()
            assert '父页对象: lab-b' in middle and '选中对象: wlan0' in middle, middle
            assert 'lab-b · DOWN' in middle, middle
            # Re-enter the same interface without changing the selection.
            send(b'\r', .5)
            revisited = report()
            assert '父页对象: wlan0' in revisited, revisited
            send(b'q')
            send(b'q')
            root = report()
            assert '选中对象: lab-b' in root and 'lab-a\t' in root, root
            # Return to the same device, then unwind normally.
            send(b'\r', .5)
            same_device = report()
            assert '父页对象: lab-b' in same_device, same_device
            send(b'q')
            send(b'q')
            process.wait(timeout=3)
            assert process.returncode == 0
            assert before == termios.tcgetattr(slave), 'terminal was not restored'
            return {'depth':3, 'device':'lab-b', 'interface':'wlan0', 'same_item_reentry':True, 'reports':5, 'termios_restored':True}
        finally:
            if process.poll() is None:
                process.kill()
                process.wait()
            os.close(master)
            os.close(slave)

if __name__ == '__main__':
    print(json.dumps(scenario(), indent=2))
