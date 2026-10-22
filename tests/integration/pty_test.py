#!/usr/bin/env python3
"""Real Linux PTY lifecycle and keyboard acceptance; no external Python modules."""
import os, pty, fcntl, termios, struct, subprocess, select, time, signal, json
from pathlib import Path
ROOT = Path(__file__).resolve().parents[2]

def drain(fd, seconds=.2):
    end = time.monotonic()+seconds
    data = b''
    while time.monotonic()<end:
        if select.select([fd], [], [], min(.03, max(0,end-time.monotonic())))[0]:
            try: data += os.read(fd, 65536)
            except OSError: break
    return data

def scenario(sig=None):
    master, slave = pty.openpty()
    fcntl.ioctl(slave, termios.TIOCSWINSZ, struct.pack('HHHH',40,120,0,0))
    before=termios.tcgetattr(slave)
    def setup():
        os.setsid()
        fcntl.ioctl(0, termios.TIOCSCTTY, 0)
    p=subprocess.Popen([str(ROOT/'bin/mdtui'), '--no-color', str(ROOT/'examples/network.md')],stdin=slave,stdout=slave,stderr=slave,env={**os.environ,'TERM':'xterm-256color','LANG':'C.UTF-8'},preexec_fn=setup)
    output=drain(master,.8)
    assert p.poll() is None, output.decode(errors='replace')
    os.write(master,b'\r');output+=drain(master,.4)
    assert b'eth0' in output, 'selection did not drive detail output'
    # Start a stream through explicit confirmation (default cancel, Tab toggles).
    os.write(master,b'd');output+=drain(master,.15)
    os.write(master,b'\t\r');output+=drain(master,.5)
    for rows,cols in [(24,80),(15,59),(40,120)]:
        fcntl.ioctl(slave,termios.TIOCSWINSZ,struct.pack('HHHH',rows,cols,0,0))
        os.kill(p.pid,signal.SIGWINCH);output+=drain(master,.1)
    started=time.monotonic()
    if sig: os.kill(p.pid,sig)
    else: os.write(master,b'q')
    output+=drain(master,.3);p.wait(timeout=3)
    elapsed=time.monotonic()-started
    assert p.returncode==0,(p.returncode,output[-1000:])
    assert before==termios.tcgetattr(slave),'terminal settings not restored'
    assert b'\x1b[?1049l' in output,'alternate screen not restored'
    assert elapsed<2,elapsed
    os.close(master);os.close(slave)
    return {'exit':signal.Signals(sig).name if sig else 'q','seconds':round(elapsed,3),'output_bytes':len(output),'termios_restored':True}

def input_search_scenario():
    master,slave=pty.openpty()
    fcntl.ioctl(slave,termios.TIOCSWINSZ,struct.pack('HHHH',24,80,0,0))
    before=termios.tcgetattr(slave)
    def setup():
        os.setsid();fcntl.ioctl(0,termios.TIOCSCTTY,0)
    p=subprocess.Popen([str(ROOT/'bin/mdtui'),'--no-color',str(ROOT/'examples/input.md')],stdin=slave,stdout=slave,stderr=slave,env={**os.environ,'TERM':'xterm-256color','LANG':'C.UTF-8'},preexec_fn=setup)
    try:
        output=drain(master,.4)
        # Initial focus is the panel. Two Tabs reach the input field.
        os.write(master,b'\t\t\rabcd')
        output+=drain(master,.2)
        os.write(master,b'\x1b[H\x1b[C\x1b[3~Z\x1b[F\x7f\r')
        output+=drain(master,.6)
        assert b'aZc' in output,'cursor editing did not submit the expected value'
        os.write(master,b'\t/aZc\r')
        output+=drain(master,.2)
        os.write(master,b'nNq')
        output+=drain(master,.2)
        p.wait(timeout=3)
        assert p.returncode==0,p.returncode
        assert before==termios.tcgetattr(slave),'input scenario did not restore terminal'
        assert b'\x1b[?1049l' in output,'input scenario did not restore alternate screen'
        return {'submitted_value':'aZc','search_keys':True,'termios_restored':True}
    finally:
        if p.poll() is None:
            p.kill();p.wait()
        os.close(master);os.close(slave)

if __name__=='__main__':
    results=[scenario(),scenario(signal.SIGTERM),scenario(signal.SIGINT)]
    p=subprocess.run([str(ROOT/'bin/mdtui'),str(ROOT/'examples/network.md')],capture_output=True)
    assert p.returncode==3 and b'TTY' in p.stderr
    print(json.dumps({'pty':results,'input_search':input_search_scenario(),'no_tty_exit':p.returncode},indent=2))
