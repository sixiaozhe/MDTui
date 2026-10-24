#!/usr/bin/env python3
"""Exercise SGR mouse reporting through a real xterm-compatible Linux PTY."""
import fcntl
import json
import math
import os
import pty
import struct
import subprocess
import tempfile
import termios
from pathlib import Path
from pty_test import drain

ROOT = Path(__file__).resolve().parents[2]

def scenario(disabled=False):
    with tempfile.TemporaryDirectory(prefix='mdtui-mouse-') as directory:
        directory = Path(directory)
        adapter = str(ROOT/'bin/net-debug')
        command = json.dumps([adapter, 'list'])
        (directory/'root.md').write_text(
            ':::mdtui page\nversion: 1\nid: root\n:::\n'
            ':::mdtui panel\nid: list\nformat: items-json\nexec: '+command+'\nopen: ./child.md\n:::\n'
            ':::mdtui panel\nid: preview\nexec: '+json.dumps([adapter, 'echo', {'ref':'list.cursor.id'}])+'\n:::\n')
        args = [adapter, 'echo', {'ref':'verbose.value'}, {'ref':'protocol.value'}, {'ref':'timeout.value'}, {'ref':'parent.selected.id'}]
        (directory/'child.md').write_text(
            ':::mdtui page\nversion: 1\nid: child\nparent: list\n:::\n'
            ':::mdtui input\nid: verbose\ntype: toggle\n:::\n'
            ':::mdtui input\nid: protocol\ntype: select\noptions: [{value: tcp}, {value: udp}]\n:::\n'
            ':::mdtui input\nid: timeout\ntype: number\ndefault: 5\nmin: 1\nmax: 60\n:::\n'
            ':::mdtui panel\nid: result\nformat: text\nautostart: false\nexec: '+json.dumps(args)+'\n:::\n'
            ':::mdtui action\nid: apply\nlabel: Apply\nbutton: true\nrun: result\nconfirm: Continue?\n:::\n')
        width, height = 120, 30
        master, slave = pty.openpty()
        fcntl.ioctl(slave, termios.TIOCSWINSZ, struct.pack('HHHH', height, width, 0, 0))
        before = termios.tcgetattr(slave)
        def setup():
            os.setsid()
            fcntl.ioctl(0, termios.TIOCSCTTY, 0)
        flags = ['--no-mouse'] if disabled else []
        process = subprocess.Popen(
            [str(ROOT/'bin/mdtui'), '--no-color', *flags, str(directory/'root.md')],
            cwd=directory, stdin=slave, stdout=slave, stderr=slave,
            env={**os.environ, 'TERM':'xterm-256color', 'LANG':'C.UTF-8'}, preexec_fn=setup)
        def send(data, delay=.15):
            os.write(master, data)
            return drain(master, delay)
        def click(x, y):
            return send(f'\x1b[<0;{x+1};{y+1}M\x1b[<0;{x+1};{y+1}m'.encode())
        def report():
            existing = set(directory.glob('mdtui-report-*.md'))
            send(b'\x1bOS', .25)
            new = set(directory.glob('mdtui-report-*.md')) - existing
            assert len(new) == 1, new
            text = new.pop().read_text()
            send(b'\r')
            return text
        try:
            output = drain(master, .5)
            assert (b'\x1b[?1000h' in output) != disabled, 'incorrect mouse reporting mode'
            if not disabled:
                # A single click moves the cursor but does not navigate.
                click(10, 4)
                first = report()
                assert 'PAGE' not in first and '父页对象:' not in first, first
                # Two consecutive clicks on wlan0 confirm and open the child.
                click(10, 4)
                click(10, 4)
                drain(master, .3)
                click(35, 2)  # toggle verbose
                click(width-2, 2)  # next input: protocol
                click(35, 2)  # open dropdown
                menu_y = (height-6)//2
                click(20, menu_y+3)  # udp
                click(width-2, 2)  # next input: timeout
                click(35, 2)  # start numeric edit
                send(f'\x1b[<64;36;3M'.encode())  # wheel up: 5 -> 6
                send(b'\r')  # commit numeric draft
                click(17, 4)  # Apply, opens confirmation
                preview = 'result: ['+' '.join(json.dumps(arg) for arg in [adapter, 'echo', 'true', 'udp', '6', 'wlan0'])+']'
                extra = math.ceil(len(preview)/94)-1
                choice_y = (height-(10+extra))//2+6+extra
                click(15, choice_y)  # default cancel
                cancelled = report()
                assert '["true","udp","6","wlan0"]' not in cancelled, cancelled
                click(17, 4)
                click(27, choice_y)  # explicitly execute
                drain(master, .3)
                result = report()
                assert '父页对象: wlan0' in result and '["true","udp","6","wlan0"]' in result, result
                click(4, 1)  # back to root
                parent = report()
                assert '父页对象:' not in parent and '选中对象: wlan0' in parent, parent
            output += send(b'q', .25)
            process.wait(timeout=3)
            assert process.returncode == 0
            assert before == termios.tcgetattr(slave), 'mouse mode did not restore termios'
            assert b'\x1b[?1049l' in output, 'alternate screen was not restored'
            if not disabled:
                assert b'\x1b[?1000l' in output, 'mouse reporting was not disabled on exit'
            return {'enabled':not disabled, 'sgr_clicks':not disabled, 'termios_restored':True}
        finally:
            if process.poll() is None:
                process.kill()
                process.wait()
            os.close(master)
            os.close(slave)

if __name__ == '__main__':
    print(json.dumps({'mouse':scenario(), 'opt_out':scenario(disabled=True)}, indent=2))
