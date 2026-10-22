#!/usr/bin/env python3
"""Validation must not execute commands, including invalid later declarations."""
import json,subprocess,tempfile
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
BIN=ROOT/'bin/mdtui'
def check():
    with tempfile.TemporaryDirectory(prefix='mdtui-cli-') as tmp:
        tmp=Path(tmp);marker=tmp/'executed';page=tmp/'page.md'
        header=':::mdtui page\nversion: 1\nid: validation\n:::\n'
        command=':::mdtui panel\nid: operation\nexec: '+json.dumps(['/usr/bin/touch',str(marker)])+'\n:::\n'
        page.write_text(header+command)
        p=subprocess.run([str(BIN),'check',str(page)],capture_output=True)
        assert p.returncode==0,p.stderr
        assert not marker.exists(),'check executed operation'
        page.write_text(header+command+':::mdtui panel\nid: typo\nexec: [echo]\nunknown: true\n:::\n')
        p=subprocess.run([str(BIN),str(page)],capture_output=True)
        assert p.returncode==2,p.stderr
        assert b'unknown' in p.stderr and not marker.exists()
        # Fenced declaration cannot execute, even if it contains an actual command.
        page.write_text(header+'```\n'+command+'```\n')
        p=subprocess.run([str(BIN),'check',str(page)],capture_output=True)
        assert p.returncode==0 and b'0 panels' in p.stdout and not marker.exists()
        for args in [['--version'],['--help']]:assert subprocess.run([str(BIN),*args],capture_output=True).returncode==0
        assert subprocess.run([str(BIN),'--unknown'],capture_output=True).returncode==2
    return {'check_zero_execution':True,'invalid_page_zero_execution':True,'fenced_example_inert':True,'cli_exit_codes':True}
if __name__=='__main__':print(json.dumps(check(),indent=2))
