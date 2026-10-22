#!/usr/bin/env python3
"""Collect redistribution notices from the pinned local Go module cache."""
import json,subprocess
from pathlib import Path
root=Path(__file__).resolve().parents[1]
raw=subprocess.check_output(['go','list','-m','-json','all'],cwd=root).decode()
decoder=json.JSONDecoder();modules=[]
while raw.strip():
    module,end=decoder.raw_decode(raw.lstrip());raw=raw.lstrip()[end:];modules.append(module)
folder=root/'docs/licenses';folder.mkdir(exist_ok=True)
lines=['# Third-party notices','', 'Pinned dependency licenses accompanying MDTUI binaries. The upstream notices below retain their original terms.','']
for module in modules:
    if module.get('Main') or not module.get('Dir'):continue
    source=Path(module['Dir']);label=module['Path']+'@'+module['Version']
    links=[]
    for name in ['LICENSE','LICENSE.txt','LICENSE.md','NOTICE','COPYING']:
        path=source/name
        if path.exists():
            target=folder/(module['Path'].replace('/','_')+'-'+name.replace('.','_')+'.txt')
            target.write_bytes(path.read_bytes());links.append(f'[{name}](docs/licenses/{target.name})')
    lines.append('- '+label+': '+', '.join(links))
(root/'THIRD_PARTY_NOTICES.md').write_text('\n'.join(lines)+'\n')
