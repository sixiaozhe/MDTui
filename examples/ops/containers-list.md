# 容器列表

本页在 Markdown exec 内调用 Python3，将原生 docker 输出转成 items-json。F3 过滤，Enter / 双击下钻。没有 MDTUI 专用运维子命令。

:::mdtui page
version: 1
id: ops-objects-container
title: 容器列表
parent: tasks
max_parallel: 2
:::

:::mdtui panel
id: objects
title: 容器列表
format: items-json
refresh: 10s
open: ./container-detail.md
timeout: 12s
exec: ["python3", "-c", "import json, subprocess, sys\nresult=subprocess.run(['docker', 'ps', '-a', '--format', '{{.ID}}\\t{{.Names}}\\t{{.Status}}'],capture_output=True,text=True,timeout=8)\nif result.returncode:\n    print(result.stderr,file=sys.stderr)\n    sys.exit(result.returncode)\nitems=[]\nfor line in result.stdout.splitlines():\n    fields=line.lstrip('● ').split()\n    parts=line.split('\\t',2)\n    if len(parts)!=3: continue\n    ident,label,description=parts\n    if not ident or ident.startswith('-') or any(ord(ch)<32 for ch in ident): continue\n    items.append({'id':ident,'label':label[:120],'description':description[:240]})\n    if len(items)>=200: break\nprint(json.dumps({'version':1,'items':items},ensure_ascii=False))\n"]
:::
