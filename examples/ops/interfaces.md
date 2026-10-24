# 网卡列表

本页在 Markdown exec 内调用 Python3，将原生 ip 输出转成 items-json。F3 过滤，Enter / 双击下钻。没有 MDTUI 专用运维子命令。

:::mdtui page
version: 1
id: ops-objects-interface
title: 网卡列表
parent: tasks
max_parallel: 2
:::

:::mdtui panel
id: objects
title: 网卡列表
format: items-json
refresh: 10s
open: ./interface-detail.md
timeout: 12s
exec: ["python3", "-c", "import json, subprocess, sys\nresult=subprocess.run(['ip', '-o', 'link', 'show'],capture_output=True,text=True,timeout=8)\nif result.returncode:\n    print(result.stderr,file=sys.stderr)\n    sys.exit(result.returncode)\nitems=[]\nfor line in result.stdout.splitlines():\n    fields=line.lstrip('● ').split()\n    if len(fields)<3: continue\n    ident=fields[1].rstrip(':').split('@',1)[0]\n    label,description=ident,' '.join(fields[2:])\n    if not ident or ident.startswith('-') or any(ord(ch)<32 for ch in ident): continue\n    items.append({'id':ident,'label':label[:120],'description':description[:240]})\n    if len(items)>=200: break\nprint(json.dumps({'version':1,'items':items},ensure_ascii=False))\n"]
:::
