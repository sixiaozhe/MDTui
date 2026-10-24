#!/usr/bin/env python3
"""Exercise filtering, child context and report export through a real PTY."""
import json
import re
import os
import pty
import fcntl
import termios
import struct
import subprocess
import tempfile
from pathlib import Path
from pty_test import drain

ROOT = Path(__file__).resolve().parents[2]

def scenario(example=False):
    with tempfile.TemporaryDirectory(prefix='mdtui-workspace-') as directory:
        directory = Path(directory)
        items = json.dumps({'version': 1, 'items': [
            {'id': 'A', 'label': 'alpha'}, {'id': 'B', 'label': 'beta'}]})
        command = json.dumps(['/usr/bin/printf', '%s', items])
        (directory/'root.md').write_text(
            ':::mdtui page\nversion: 1\nid: root\ntitle: Root\n:::\n'
            ':::mdtui panel\nid: list\nformat: items-json\nexec: '+command+
            '\nopen: ./child.md\n:::\n')
        (directory/'child.md').write_text(
            ':::mdtui page\nversion: 1\nid: child\ntitle: Child\nparent: list\n:::\n'
            ':::mdtui panel\nid: detail\nexec: [/usr/bin/printf, "%s", {ref: parent.selected.id}]\n:::\n')
        master, slave = pty.openpty()
        fcntl.ioctl(slave, termios.TIOCSWINSZ, struct.pack('HHHH', 24, 80, 0, 0))
        before = termios.tcgetattr(slave)
        def setup():
            os.setsid()
            fcntl.ioctl(0, termios.TIOCSCTTY, 0)
        page = ROOT/'examples/showcase.md' if example else directory/'root.md'
        selected = 'wlan0' if example else 'B'
        process = subprocess.Popen(
            [str(ROOT/'bin/mdtui'), '--no-color', str(page)],
            cwd=directory, stdin=slave, stdout=slave, stderr=slave,
            env={**os.environ, 'TERM': 'xterm-256color', 'LANG': 'C.UTF-8'}, preexec_fn=setup)
        try:
            output = drain(master, .5)
            os.write(master, b'\x1bOR'+(b'wlan0' if example else b'beta')+b'\r')  # F3, apply filter
            output += drain(master, .2)
            os.write(master, b'\r')  # open the filtered B item
            child_output = drain(master, .5)
            text = re.sub(rb'\x1b\[[0-?]*[ -/]*[@-~]', b'', child_output)
            assert '父页对象'.encode() in text, 'child context was not displayed'
            if example:
                os.write(master, b'f')  # include the simulated error in the snapshot
                output += drain(master, .3)
            os.write(master, b'\x1bOS')  # F4 export child
            output += drain(master, .3)
            reports = list(directory.glob('mdtui-report-*.md'))
            assert len(reports) == 1, reports
            child = reports[0].read_text()
            breadcrumb = '功能体验台 → 接口诊断' if example else 'Root → Child'
            assert breadcrumb in child and '父页对象: '+selected in child, child
            if example:
                for expected in ['演示采集', '模拟故障: unavailable', '模拟诊断正常', 'wlan0']:
                    assert expected in child, (expected, child)
            else:
                assert '\nB\n' in child, child
            os.write(master, b'\rq')  # close export result, return to parent
            output += drain(master, .3)
            os.write(master, b'\x1bOS')
            output += drain(master, .3)
            reports = list(directory.glob('mdtui-report-*.md'))
            assert len(reports) == 2, reports
            root = next(path.read_text() for path in reports if breadcrumb not in path.read_text())
            assert '选中对象: '+selected in root, root
            assert ('eth0\t' in root and 'wlan0\t' in root) if example else ('A\talpha' in root and 'B\tbeta' in root), root
            os.write(master, b'\rq')  # close export result, exit root
            output += drain(master, .2)
            process.wait(timeout=3)
            assert process.returncode == 0, output
            assert before == termios.tcgetattr(slave), 'terminal not restored'
            return {'filter_selection': selected, 'navigation_context': True, 'reports': 2, 'termios_restored': True}
        finally:
            if process.poll() is None:
                process.kill()
                process.wait()
            os.close(master)
            os.close(slave)

def controls_scenario():
    with tempfile.TemporaryDirectory(prefix='mdtui-controls-') as directory:
        directory = Path(directory)
        master, slave = pty.openpty()
        fcntl.ioctl(slave, termios.TIOCSWINSZ, struct.pack('HHHH', 24, 80, 0, 0))
        before = termios.tcgetattr(slave)
        def setup():
            os.setsid()
            fcntl.ioctl(0, termios.TIOCSCTTY, 0)
        process = subprocess.Popen(
            [str(ROOT/'bin/mdtui'), '--no-color', str(ROOT/'examples/controls.md')],
            cwd=directory, stdin=slave, stdout=slave, stderr=slave,
            env={**os.environ, 'TERM': 'xterm-256color', 'LANG': 'C.UTF-8'}, preexec_fn=setup)
        try:
            drain(master, .4)
            # Initial focus is the result panel; Shift+Tab reaches the number.
            os.write(master, b'\x1b[Z\r\x01\x1b[3~61\r')
            output = drain(master, .2)
            clean = re.sub(rb'\x1b\[[0-?]*[ -/]*[@-~]', b'', output)
            assert '数值不得大于'.encode() in clean, 'out-of-range number was not rejected'
            os.write(master, b'\x01\x1b[3~\x1b[3~12\r')
            drain(master, .1)
            # Cancel a selection, reopen and commit UDP.
            os.write(master, b'\x1b[Z\r\x1b[B\x1b')
            drain(master, .15)
            os.write(master, b'\r\x1b[B\r')
            drain(master, .15)
            os.write(master, b'\x1b[Z ')  # toggle verbose with Space
            drain(master, .1)
            os.write(master, b'\x1b[Z\r\x01'+b'\x1b[3~'*4+'设备甲'.encode()+b'\r')
            drain(master, .1)
            # Five Tabs reach the primary button from the first input.
            os.write(master, b'\t'*5+b'\r')
            drain(master, .4)
            os.write(master, b'\x1bOS')
            drain(master, .25)
            reports = list(directory.glob('mdtui-report-*.md'))
            assert len(reports) == 1, reports
            report = reports[0].read_text()
            for expected in ['"--note","设备甲"', '"--verbose","true"', '"--protocol","udp"', '"--timeout","12"']:
                assert expected in report, (expected, report)
            # The second button uses the existing confirmation dialog.
            os.write(master, b'\r\t\r')
            output = drain(master, .15)
            clean = re.sub(rb'\x1b\[[0-?]*[ -/]*[@-~]', b'', output)
            assert '确认操作'.encode() in clean, 'button did not open confirmation'
            os.write(master, b'\r')  # default cancel
            drain(master, .1)
            os.write(master, b'\r\t\r')  # reopen, select execute and confirm
            drain(master, .3)
            os.write(master, b'q')
            drain(master, .2)
            process.wait(timeout=3)
            assert process.returncode == 0
            assert before == termios.tcgetattr(slave), 'controls scenario did not restore terminal'
            return {'toggle': True, 'select': 'udp', 'number': 12, 'button': True, 'confirmation': True, 'export': True}
        finally:
            if process.poll() is None:
                process.kill()
                process.wait()
            os.close(master)
            os.close(slave)

if __name__ == '__main__':
    print(json.dumps({'fixture': scenario(), 'showcase': scenario(example=True), 'controls': controls_scenario()}, indent=2))
