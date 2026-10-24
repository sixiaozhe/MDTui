#!/usr/bin/env python3
"""Read-only smoke checks and PTY acceptance for the comprehensive ops suite."""
import fcntl
import json
import os
import pty
import re
import signal
import struct
import subprocess
import tempfile
import termios
from pathlib import Path
from pty_test import drain

ROOT = Path(__file__).resolve().parents[2]

def cli_checks():
    pages = ROOT/'examples/ops'
    def commands(name):
        return [json.loads(line[6:]) for line in (pages/name).read_text().splitlines() if line.startswith('exec: ')]
    for page in pages.glob('*.md'):
        for command in commands(page.name):
            assert command[0] not in ('linux-tools', 'mdtui'), (page, command)
    areas = subprocess.run(commands('index.md')[0],capture_output=True,text=True,check=True)
    assert len(json.loads(areas.stdout)['items']) == 15
    objects = subprocess.run(commands('processes.md')[0],capture_output=True,text=True,check=True,timeout=15)
    assert json.loads(objects.stdout)['items']
    changes = (pages/'changes.md').read_text()
    assert changes.splitlines().count('autostart: false') == 6
    assert changes.count('confirm: ') == 6
    assert 'refresh:' not in changes
    return {'areas':15, 'native_commands':True, 'process_objects':True, 'changes_require_confirmation':True}

def pty_scenario():
    with tempfile.TemporaryDirectory(prefix='mdtui-ops-') as directory:
        directory = Path(directory)
        master,slave=pty.openpty()
        fcntl.ioctl(slave,termios.TIOCSWINSZ,struct.pack('HHHH',40,120,0,0))
        before=termios.tcgetattr(slave)
        def setup():
            os.setsid()
            fcntl.ioctl(0,termios.TIOCSCTTY,0)
        process=subprocess.Popen([str(ROOT/'bin/mdtui'),'--no-color',str(ROOT/'examples/ops/index.md')],cwd=directory,stdin=slave,stdout=slave,stderr=slave,env={**os.environ,'TERM':'xterm-256color','LANG':'C.UTF-8'},preexec_fn=setup)
        def send(data,delay=.3):
            os.write(master,data)
            return drain(master,delay)
        def report():
            existing=set(directory.glob('mdtui-report-*.md'))
            send(b'\x1bOS')
            new=set(directory.glob('mdtui-report-*.md'))-existing
            assert len(new)==1,new
            text=new.pop().read_text()
            send(b'\r',.1)
            return text
        try:
            drain(master,.7)
            send(b'\r')  # system
            send(b'\x1b[B\r')  # kernel task
            send(b'r',.5)
            kernel=report()
            assert '父页对象: system.kernel' in kernel and 'Linux' in kernel, kernel
            send(b'q')
            send(b'q')
            send(b'\x1bORprocess\r')
            send(b'\r')
            send(b'\x1bORprocess.detail\r')
            send(b'\r',.5)  # live process browser, third level
            send(b'\r',.5)  # process detail, fourth level
            detail=report()
            assert '进程与性能 [process] → 进程列表 [process.detail] → 进程详情' in detail, detail
            assert re.search(r'父页对象: [0-9]+',detail) and 'PID' in detail,detail
            send(b'q')
            send(b'q')
            send(b'q')
            send(b'q')
            process.wait(timeout=3)
            assert process.returncode==0
            assert before==termios.tcgetattr(slave)
            return {'live_query':True,'four_level_navigation':True,'export':True,'termios_restored':True}
        finally:
            if process.poll() is None:
                process.kill();process.wait()
            os.close(master);os.close(slave)

if __name__=='__main__':
    print(json.dumps({'cli':cli_checks(),'pty':pty_scenario()},indent=2))
