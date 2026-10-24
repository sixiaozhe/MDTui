# 用户与会话

在线会话、账户、组和登录历史。选择任务 Enter 查看原生命令和参数；进入普通任务页不会自动采集。

:::mdtui page
version: 1
id: ops-users
title: 用户与会话
parent: areas
max_parallel: 2
:::

:::mdtui panel
id: tasks
title: 任务清单 · F3 过滤
format: items-json
open:
  users.sessions: ./users-sessions.md
  users.accounts: ./users-accounts.md
  users.groups: ./users-groups.md
  users.identity: ./users-identity.md
  users.last: ./users-last.md
  users.failed: ./users-failed.md
exec: ["printf", "%s\n", "{\"version\": 1, \"items\": [{\"id\": \"users.sessions\", \"label\": \"在线会话\", \"description\": \"查看在线用户和登录来源\"}, {\"id\": \"users.accounts\", \"label\": \"账户清单\", \"description\": \"读取 NSS 账户信息，不读取 shadow\"}, {\"id\": \"users.groups\", \"label\": \"组清单\", \"description\": \"读取 NSS 组信息\"}, {\"id\": \"users.identity\", \"label\": \"用户身份\", \"description\": \"UID、GID 和补充组\"}, {\"id\": \"users.last\", \"label\": \"最近登录\", \"description\": \"读取登录历史，不修改记录\"}, {\"id\": \"users.failed\", \"label\": \"失败登录\", \"description\": \"读取 btmp，通常需要额外权限\"}]}"]
:::
