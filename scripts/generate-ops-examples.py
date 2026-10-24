#!/usr/bin/env python3
"""Generate ordinary Markdown calling native Linux commands, not MDTUI ops."""
import json
from pathlib import Path

ROOT=Path(__file__).resolve().parents[1]
OUT=ROOT/'examples/ops'
catalog=json.loads((OUT/'catalog.json').read_text(encoding='utf-8'))
areas,tasks=catalog['areas'],catalog['tasks']

def block(kind,body): return ':::mdtui '+kind+'\n'+body.strip()+'\n:::\n\n'
def panel(identifier,title,command,extra=''):
    return block('panel',f'id: {identifier}\ntitle: {title}\n'+extra+'exec: '+json.dumps(command,ensure_ascii=False))
def action(identifier,label,key,target,confirm=''):
    value=f'id: {identifier}\nlabel: {label}\nkey: {key}\nbutton: true\nrun: {target}'
    if confirm: value+='\nconfirm: '+json.dumps(confirm,ensure_ascii=False)
    return block('action',value)
def menu(items): return ['printf','%s\n',json.dumps({'version':1,'items':items},ensure_ascii=False)]
def write(name,text): (OUT/name).write_text(text.rstrip()+'\n',encoding='utf-8')
def slug(task): return task['id'].replace('.','-')+'.md'

def protect(command):
    program,args=command['program'],list(command['args'])
    if '{target}' in args and program in {'systemctl','docker','kubectl','id','getent','ping','traceroute','tracepath'}:
        index=args.index('{target}')
        if '--' not in args[:index]: args.insert(index,'--')
    if program=='curl': args=['--proto','=http,https','--proto-redir','=http,https',*args]
    return {'program':program,'args':args}

def native(command,reference='target.value',fixed_limit=None):
    args=[command['program']]
    for value in command['args']:
        if value=='{target}': args.append({'ref':reference})
        elif value=='{limit}': args.append(str(fixed_limit) if fixed_limit else {'ref':'limit.value'})
        else:
            if '{target}' in value or '{limit}' in value: raise ValueError('Refs must be complete arguments: '+value)
            args.append(value)
    return args

# Full-path inputs avoid requiring string concatenation or expressions in MDTUI.
for task in tasks:
    if task['id'] in {'process.status','process.limits'}:
        suffix=task['id'].split('.')[1]
        task.update(target_kind='path',default_target=f'/proc/1/{suffix}',target_hint=f'完整绝对路径，例如 /proc/1/{suffix}')
        task['commands']=[{'program':'cat','args':['--','{target}']}]
    if task['id'] in {'process.files','services.detail','network.dns'}: task['commands']=task['commands'][:1]
    task['commands']=[protect(command) for command in task['commands']]
    task['follow']=[protect(command) for command in task.get('follow',[])]
lookup={task['id']:task for task in tasks}
special={'process.detail':'processes.md','network.interfaces':'interfaces.md','services.detail':'services-list.md','containers.list':'containers-list.md'}

text='# Linux 运维命令工作台\n\n所有任务在 Markdown exec 中直接调用 Linux 原生命令，没有 MDTUI 运维子命令。Enter / 双击进入，F3 过滤，F4 导出。普通入口只读，变更操作在独立页面确认。\n\n'
text+=block('page','version: 1\nid: ops-home\ntitle: Linux 运维命令工作台\nlayout: master-detail\nsidebar: areas\nmax_parallel: 4')
mapping='open:\n'+''.join(f'  {area["id"]}: ./{area["id"]}.md\n' for area in areas)
text+=panel('areas','运维主题 · Enter / 双击进入',menu(areas),'format: items-json\n'+mapping)
for identifier,title,command,cap in [('kernel','主机内核',['uname','-a'],5),('load','运行时间与负载',['uptime'],5),('memory','内存概览',['free','-h'],7),('filesystems','文件系统',['df','-hT'],None)]:
    text+=panel(identifier,title,command,'format: text\n'+(f'max_height: {cap}\n' if cap else ''))
write('index.md',text)

for area in areas:
    if area['id']=='changes': continue
    group=[task for task in tasks if task['area']==area['id']]
    text=f'# {area["label"]}\n\n{area["description"]}。选择任务 Enter 查看原生命令和参数；进入普通任务页不会自动采集。\n\n'
    text+=block('page',f'version: 1\nid: ops-{area["id"]}\ntitle: {area["label"]}\nparent: areas\nmax_parallel: 2')
    mapping='open:\n'+''.join(f'  {task["id"]}: ./'+special.get(task['id'],slug(task))+'\n' for task in group)
    items=[{'id':task['id'],'label':task['label'],'description':task['description']} for task in group]
    text+=panel('tasks','任务清单 · F3 过滤',menu(items),'format: items-json\n'+mapping)
    write(area['id']+'.md',text)

for task in tasks:
    text=f'# {task["label"]}\n\n{task["description"]}。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。\n\n'
    for i,command in enumerate(task['commands']):
        text+=('主命令' if i==0 else '备用命令')+'：\n\n```text\n'+json.dumps(native(command),ensure_ascii=False)+'\n```\n\n'
    if task['target_kind']!='none': text+='对象说明：'+task['target_hint']+'。\n\n'
    text+=block('page',f'version: 1\nid: ops-{task["id"].replace(".","-")}\ntitle: {task["label"]}\nparent: tasks\nmax_parallel: 3')
    if task['target_kind']!='none':
        value=f'id: target\nlabel: {task["target_hint"] or "操作对象"}\ndefault: '+json.dumps(task['default_target'],ensure_ascii=False)+'\nrequired: true\nmax_length: 1024'
        if task['target_kind']=='pid': value+='\ntype: number\nmin: 1\nmax: 2147483647'
        text+=block('input',value)
    if any('{limit}' in arg for command in task['commands']+task['follow'] for arg in command['args']):
        text+=block('input','id: limit\nlabel: 日志行数\ntype: number\ndefault: 50\nmin: 1\nmax: 200\nrequired: true')
    for i,command in enumerate(task['commands']):
        identifier='result' if i==0 else 'fallback'
        text+=panel(identifier,'原生命令输出' if i==0 else '备用命令输出',native(command),f'format: text\nautostart: false\ntimeout: {max(10,task["timeout_seconds"]+2)}s\n')
        text+=action('collect' if i==0 else 'collect_fallback','执行主命令' if i==0 else '执行备用命令','r' if i==0 else 'b',identifier)
    for i,command in enumerate(task['follow']):
        identifier='follow' if i==0 else 'follow_fallback'
        text+=panel(identifier,'实时日志' if i==0 else '备用日志跟随',native(command),'mode: stream\nformat: text\nautostart: false\n')
        text+=action('start_follow' if i==0 else 'start_follow_fallback','开始日志跟随' if i==0 else '备用方式跟随','l' if i==0 else 'f',identifier)
    write(slug(task),text)

# Inline formatting adapters only: native tools still provide all real data.
object_commands={
 'process':['ps','-eo','pid=,comm=,%cpu=','--sort=-%cpu'],
 'interface':['ip','-o','link','show'],
 'service':['systemctl','list-units','--all','--type=service','--no-pager','--plain','--no-legend'],
 'container':['docker','ps','-a','--format','{{.ID}}\t{{.Names}}\t{{.Status}}'],
}
object_parsers={
 'process':"""    if len(fields)<3 or not fields[0].isdigit(): continue
    ident=fields[0]
    label='PID '+ident+' · '+' '.join(fields[1:-1])
    description='CPU '+fields[-1]+'%'""",
 'interface':"""    if len(fields)<3: continue
    ident=fields[1].rstrip(':').split('@',1)[0]
    label,description=ident,' '.join(fields[2:])""",
 'service':"""    if len(fields)<4 or not fields[0].endswith('.service'): continue
    ident,label,description=fields[0],fields[0],' '.join(fields[2:])""",
 'container':"""    parts=line.split('\\t',2)
    if len(parts)!=3: continue
    ident,label,description=parts""",
}
def objects(kind):
    source='import json, subprocess, sys\n'
    source+='result=subprocess.run('+repr(object_commands[kind])+',capture_output=True,text=True,timeout=8)\n'
    source+="if result.returncode:\n    print(result.stderr,file=sys.stderr)\n    sys.exit(result.returncode)\nitems=[]\n"
    source+="for line in result.stdout.splitlines():\n    fields=line.lstrip('● ').split()\n"+object_parsers[kind]+'\n'
    source+="    if not ident or ident.startswith('-') or any(ord(ch)<32 for ch in ident): continue\n"
    source+="    items.append({'id':ident,'label':label[:120],'description':description[:240]})\n    if len(items)>=200: break\n"
    source+="print(json.dumps({'version':1,'items':items},ensure_ascii=False))\n"
    return ['python3','-c',source]

resources=[('process','processes.md','process-detail.md','进程列表','进程详情','process.detail','process.files','5s'),('interface','interfaces.md','interface-detail.md','网卡列表','网卡详情','network.interface','network.statistics','10s'),('service','services-list.md','service-detail.md','服务列表','服务详情','services.detail','logs.unit','10s'),('container','containers-list.md','container-detail.md','容器列表','容器详情','containers.inspect','containers.logs','10s')]
for kind,browser,leaf,label,title,main,extra,refresh in resources:
    text=f'# {label}\n\n本页在 Markdown exec 内调用 Python3，将原生 {object_commands[kind][0]} 输出转成 items-json。F3 过滤，Enter / 双击下钻。没有 MDTUI 专用运维子命令。\n\n'
    text+=block('page',f'version: 1\nid: ops-objects-{kind}\ntitle: {label}\nparent: tasks\nmax_parallel: 2')
    text+=panel('objects',label,objects(kind),f'format: items-json\nrefresh: {refresh}\nopen: ./{leaf}\ntimeout: 12s\n')
    write(browser,text)
    text=f'# {title}\n\n引用父页对象 ID，执行原生命令查询状态。r 刷新，e 读取补充信息，F4 导出，Esc / q 逐层返回。\n\n'
    text+=block('page',f'version: 1\nid: ops-object-{kind}\ntitle: {title}\nparent: objects\nmax_parallel: 3')
    text+=panel('detail','对象状态',native(lookup[main]['commands'][0],'parent.selected.id',80),'format: text\ntimeout: 15s\n')
    text+=action('refresh_object','刷新对象状态','r','detail')
    text+=panel('extra','补充诊断 · e 执行',native(lookup[extra]['commands'][0],'parent.selected.id',80),'format: text\nautostart: false\ntimeout: 15s\n')
    text+=action('collect_extra','采集补充信息','e','extra')
    if lookup[extra]['follow']:
        text+=panel('follow','对象实时日志',native(lookup[extra]['follow'][0],'parent.selected.id',50),'mode: stream\nformat: text\nautostart: false\n')
        text+=action('start_follow','开始日志跟随','l','follow')
    write(leaf,text)

text='# 显式变更操作\n\nexec 直接调用 systemctl、kill 或 docker。变更面板全部为 autostart: false，仅通过带确认框的动作执行；确认框显示原生 argv。浏览不会修改系统，不会自动 sudo。\n\n'
text+=block('page','version: 1\nid: ops-changes\ntitle: 变更操作\nparent: areas\nmax_parallel: 2')
text+=block('input','id: service\nlabel: 明确的服务名\nrequired: true\nmax_length: 128')
text+=block('input','id: pid\nlabel: 终止的进程 PID（至少为 2）\ntype: number\nmin: 2\nmax: 2147483647\nrequired: true')
text+=block('input','id: container\nlabel: 明确的 Docker 容器名 / ID\nrequired: true\nmax_length: 128')
for operation,label,key in [('start','启动服务','a'),('stop','停止服务','s'),('restart','重启服务','r'),('reload','重载服务','l')]:
    identifier='service_'+operation
    text+=panel(identifier,label,['systemctl','--no-ask-password',operation,'--',{'ref':'service.value'}],'format: text\nautostart: false\ntimeout: 20s\n')
    text+=action('do_'+operation,label,key,identifier,'确认'+label+'？请核对下方原生命令中的对象。')
text+=panel('terminate','发送 TERM',['kill','--signal','TERM','--',{'ref':'pid.value'}],'format: text\nautostart: false\ntimeout: 10s\n')
text+=action('do_term','终止指定进程','t','terminate','确认向下方 PID 发送 TERM？该进程可能承担关键业务。')
text+=panel('restart_container','重启容器',['docker','restart','--time','10','--',{'ref':'container.value'}],'format: text\nautostart: false\ntimeout: 25s\n')
text+=action('do_container','重启指定容器','c','restart_container','确认重启下方 Docker 容器？')
write('changes.md',text)

lines=['# Linux 运维 Markdown 示例集','','每个任务都在普通 MDTUI Markdown 的 exec 中直接声明 Linux 原生命令；MDTUI 核心没有增加运维子命令或专用分发器。','','```sh','bash scripts/build.sh','./bin/mdtui check examples/ops/index.md','./bin/mdtui examples/ops/index.md','```','','## 结构与操作','','- 首页 → 主题 → 独立任务页。进程、网卡、服务和容器支持主题 → 对象列表 → 对象详情，共四层。','- 静态导航通过 GNU coreutils 的 printf 输出 items-json，不依赖 linux-tools 或 net-debug。','- 四种动态对象列表只在 Markdown 中使用 Python3 小片段，把 ps/ip/systemctl/docker 输出转换成既有 items-json；不是 MDTUI 内置功能。','- 普通任务页先显示原生命令及参数，r/按钮执行，b 使用备用命令；备用命令不自动运行。对象详情页自动查询所选对象状态。','- 日志页 l 启动跟随，Ctrl+C 取消，Space 暂停显示跟随。F3 过滤列表，F4 导出。','','## 环境与边界','','需要 GNU/Linux 及各任务对应的原生工具。systemd、iproute2、procps、sysstat、lsof、smartmontools、Docker、kubectl、数据库客户端按主题使用，不会自动安装或提权。仅四种动态对象列表需要 Python3。缺少程序、权限不足、daemon 不可用或对象消失时展示原生错误。','','Kubernetes 使用当前 kubeconfig；数据库使用现有客户端环境或认证配置，页面不收集密码。进程参数、Docker inspect 和配置可能包含敏感内容，分享导出前请检查。快照和日志沿用 MDTUI 的有界输出缓存；日志行数由数字输入传给原生命令。','','变更页面直接执行原生命令，但所有面板禁止自动启动和刷新，操作必须通过确认框。PID 输入至少为 2；服务和容器对象前用 -- 结束选项。自动化验证不会实际执行变更。','','## 主题目录','','| 主题 | 任务数 | 内容 |','|---|---|---|']
for area in areas:
    group=[task for task in tasks if task['area']==area['id']]
    lines.append(f'| [{area["label"]}]({area["id"]}.md) | {len(group) if group else 6} | {area["description"]} |')
lines+=['','## 任务和原生命令','','| 页面 | 首选原生命令 |','|---|---|']
for task in tasks:
    command=' '.join([task['commands'][0]['program'],*task['commands'][0]['args']]).replace('|','\\|')
    lines.append(f'| [{task["label"]}]({slug(task)}) | `{command}` |')
lines+=['','## 维护','','这些 Markdown 可以直接编辑和部署。catalog.json 只是生成器的开发期数据，不会被 MDTUI 加载；批量重建运行 `python3 scripts/generate-ops-examples.py`，会覆盖生成页面。部署运行不需要生成器、catalog.json 或项目适配器。','']
write('README.md','\n'.join(lines))
print(f'{len(areas)} topics, {len(tasks)} native diagnostic tasks, {len(list(OUT.glob("*.md")))} Markdown files')
