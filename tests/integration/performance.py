#!/usr/bin/env python3
"""Record 8-panel RSS and CPU from /proc while draining a real PTY.

Usage: python3 tests/integration/performance.py --seconds 1800 --output docs/evidence/performance.json
The stream fixture is rate limited to ~2 MB/s, enough to evict the entire window repeatedly.
"""
import argparse,os,pty,fcntl,termios,struct,subprocess,select,time,signal,json,tempfile,statistics
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]

def run(seconds,stream,content_format='text',extra_args=()):
    with tempfile.TemporaryDirectory(prefix='mdtui-perf-') as directory:
        directory=Path(directory)
        script=directory/'source.py'
        line = ('- ' if content_format == 'markdown' else '') + '中文 diagnostic ' + 'x'*180 + '\n'
        script.write_text("import sys,time\nwhile True:\n sys.stdout.write((" + repr(line) + ")*100)\n sys.stdout.flush()\n time.sleep(.01)\n")
        page=':::mdtui page\nversion: 1\nid: performance\nmax_parallel: 2\n:::\n'
        for i in range(8):
            mode='stream' if stream and i==0 else 'snapshot'
            args=['/usr/bin/python3',str(script)] if mode=='stream' else ['/bin/echo','reference snapshot']
            panel_format = content_format if mode == 'stream' else 'text'
            page+=f':::mdtui panel\nid: p{i}\nmode: {mode}\nformat: {panel_format}\nexec: {json.dumps(args)}\n:::\n'
        path=directory/'page.md';path.write_text(page)
        master,slave=pty.openpty();fcntl.ioctl(slave,termios.TIOCSWINSZ,struct.pack('HHHH',40,120,0,0))
        def setup():
            os.setsid();fcntl.ioctl(0,termios.TIOCSCTTY,0)
        p=subprocess.Popen([str(ROOT/'bin/mdtui'),*extra_args,str(path)],stdin=slave,stdout=slave,stderr=slave,preexec_fn=setup,env={**os.environ,'TERM':'xterm-256color','LANG':'C.UTF-8'})
        def sample():
            fields=Path(f'/proc/{p.pid}/stat').read_text().split()
            rss=int(Path(f'/proc/{p.pid}/statm').read_text().split()[1])*os.sysconf('SC_PAGE_SIZE')
            return (int(fields[13])+int(fields[14]))/os.sysconf('SC_CLK_TCK'),rss
        started=time.monotonic();samples=[];next_sample=started+1;cpu0=None
        try:
            while time.monotonic()-started<seconds:
                assert p.poll() is None,'workbench exited unexpectedly'
                if select.select([master],[],[],.03)[0]:os.read(master,65536)
                now=time.monotonic()
                if now>=next_sample:
                    cpu,rss=sample();samples.append({'seconds':round(now-started,2),'cpu_seconds':cpu,'rss_bytes':rss})
                    next_sample=now+1
                    if cpu0 is None:cpu0=(now,cpu)
            cpu,rss=sample();cpu_percent=(cpu-cpu0[1])/(time.monotonic()-cpu0[0])*100
            exit_start=time.monotonic();os.kill(p.pid,signal.SIGTERM)
            while p.poll() is None and time.monotonic()-exit_start<3:
                if select.select([master],[],[],.02)[0]:os.read(master,65536)
            p.wait(timeout=1)
            assert p.returncode==0
            return {'mode':content_format+'-stream' if stream else 'idle','duration_seconds':seconds,'panels':8,'cpu_percent_one_core':round(cpu_percent,3),'rss_max_bytes':max(s['rss_bytes'] for s in samples),'rss_final_bytes':rss,'exit_seconds':round(time.monotonic()-exit_start,3),'samples':samples}
        finally:
            if p.poll() is None:p.kill();p.wait()
            os.close(master);os.close(slave)

if __name__=='__main__':
    parser=argparse.ArgumentParser();parser.add_argument('--seconds',type=int,default=1800);parser.add_argument('--idle-seconds',type=int,default=30);parser.add_argument('--format',choices=['text','markdown'],default='text');parser.add_argument('--output',default='docs/evidence/performance.json');parser.add_argument('--low-resource',action='store_true');args=parser.parse_args()
    if min(args.seconds,args.idle_seconds)<2: parser.error('durations must be at least 2 seconds')
    extra_args=['--low-resource'] if args.low_resource else []
    report={'environment':os.uname()._asdict() if hasattr(os.uname(),'_asdict') else str(os.uname()),'profile':'low-resource' if args.low_resource else 'default','idle':run(args.idle_seconds,False,extra_args=extra_args)}
    Path(args.output).parent.mkdir(parents=True,exist_ok=True)
    Path(args.output).write_text(json.dumps(report,indent=2))
    report['stream']=run(args.seconds,True,args.format,extra_args)
    Path(args.output).write_text(json.dumps(report,indent=2))
    print(json.dumps({k:({a:b for a,b in v.items() if a!='samples'} if isinstance(v,dict) else v) for k,v in report.items()},indent=2))
