# 进程列表

本页在 Markdown exec 内调用 Python3，将原生 ps 输出转成 items-json。F3 过滤，Enter / 双击下钻。没有 MDTUI 专用运维子命令。

:::mdtui page
version: 1
id: ops-objects-process
title: 进程列表
parent: tasks
max_parallel: 2
:::

:::mdtui panel
id: objects
title: 进程列表
format: items-json
refresh: 5s
open: ./process-detail.md
timeout: 12s
exec: ["python3", "-c", "import json, subprocess, sys\nresult=subprocess.run(['ps', '-eo', 'pid=,comm=,%cpu=', '--sort=-%cpu'],capture_output=True,text=True,timeout=8)\nif result.returncode:\n    print(result.stderr,file=sys.stderr)\n    sys.exit(result.returncode)\nitems=[]\nfor line in result.stdout.splitlines():\n    fields=line.lstrip('● ').split()\n    if len(fields)<3 or not fields[0].isdigit(): continue\n    ident=fields[0]\n    label='PID '+ident+' · '+' '.join(fields[1:-1])\n    description='CPU '+fields[-1]+'%'\n    if not ident or ident.startswith('-') or any(ord(ch)<32 for ch in ident): continue\n    items.append({'id':ident,'label':label[:120],'description':description[:240]})\n    if len(items)>=200: break\nprint(json.dumps({'version':1,'items':items},ensure_ascii=False))\n"]
:::
