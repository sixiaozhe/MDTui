#!/usr/bin/env python3
"""PTY input and incremental stream latency (flush -> terminal bytes, not pixels).

The fixture writes a unique marker only on a file trigger. The terminal frame is
fully drained before every trial; each measured marker has never appeared before.
"""
import os,sys,pty,fcntl,termios,struct,subprocess,select,time,signal,json,tempfile,statistics
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
def run():
    with tempfile.TemporaryDirectory(prefix='mdtui-latency-') as tmp:
        tmp=Path(tmp);trigger=tmp/'trigger';stamp=tmp/'stamp';fixture=tmp/'fixture.py'
        fixture.write_text("import sys,time\nfrom pathlib import Path\nt=Path(sys.argv[1]);s=Path(sys.argv[2]);last=''\nprint(('warmup '+'x'*180+'\\n')*10000,flush=True)\nwhile True:\n if t.exists():\n  value=t.read_text()\n  if value!=last:\n   s.write_text(str(time.monotonic_ns()))\n   last=value\n print(last or 'heartbeat',flush=True)\n time.sleep(.01)\n")
        page=tmp/'page.md';page.write_text(':::mdtui page\nversion: 1\nid: latency\n:::\n:::mdtui panel\nid: log\nmode: stream\nformat: text\nexec: '+json.dumps(['/usr/bin/python3',str(fixture),str(trigger),str(stamp)])+'\n:::\n')
        master,slave=pty.openpty();fcntl.ioctl(slave,termios.TIOCSWINSZ,struct.pack('HHHH',24,100,0,0))
        def setup():os.setsid();fcntl.ioctl(0,termios.TIOCSCTTY,0)
        p=subprocess.Popen([str(ROOT/'bin/mdtui'),str(page)],stdin=slave,stdout=slave,stderr=slave,preexec_fn=setup,env={**os.environ,'TERM':'xterm-256color'})
        def drain(duration):
            until=time.monotonic()+duration
            while time.monotonic()<until:
                if select.select([master],[],[],.01)[0]:os.read(master,65536)
        def until_marker(marker):
            until=time.monotonic()+2;buf=b''
            while time.monotonic()<until:
                if select.select([master],[],[],.01)[0]:
                    buf+=os.read(master,65536)
                    if marker in buf:return time.monotonic_ns()
            raise AssertionError(('marker not rendered',marker,buf[-1000:]))
        stream=[];keys=[]
        try:
            drain(.5)
            for i in range(40):
                marker=f'LATENCY_MARKER_{i:04d}'
                trigger.write_text(marker)
                visible=until_marker(marker.encode());flushed=int(stamp.read_text())
                stream.append((visible-flushed)/1e6)
                drain(.02)
                start=time.monotonic_ns();os.write(master,b'?');visible=until_marker(b'Esc')
                keys.append((visible-start)/1e6)
                os.write(master,b'\r');drain(.02)
            os.write(master,b'q');drain(.2);p.wait(timeout=2)
        finally:
            if p.poll() is None:p.kill();p.wait()
            os.close(master);os.close(slave)
        percentile=lambda xs:round(sorted(xs)[int(len(xs)*.95)-1],3)
        result={'trials':40,'key_to_pty_frame_p95_ms':percentile(keys),'fixture_flush_to_pty_frame_p95_ms':percentile(stream),'key_samples_ms':keys,'stream_samples_ms':stream,'measurement':'local PTY bytes; physical display and SSH network excluded; timestamp immediately before flush'}
        assert result['key_to_pty_frame_p95_ms']<=100,result
        assert result['fixture_flush_to_pty_frame_p95_ms']<=200,result
        return result
if __name__=='__main__':print(json.dumps(run(),indent=2))
