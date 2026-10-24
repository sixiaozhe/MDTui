# 软件包

Debian/RPM 包、文件归属和本地更新清单。选择任务 Enter 查看原生命令和参数；进入普通任务页不会自动采集。

:::mdtui page
version: 1
id: ops-packages
title: 软件包
parent: areas
max_parallel: 2
:::

:::mdtui panel
id: tasks
title: 任务清单 · F3 过滤
format: items-json
open:
  packages.installed: ./packages-installed.md
  packages.upgradable: ./packages-upgradable.md
  packages.detail: ./packages-detail.md
  packages.files: ./packages-files.md
  packages.owner: ./packages-owner.md
  packages.history: ./packages-history.md
exec: ["printf", "%s\n", "{\"version\": 1, \"items\": [{\"id\": \"packages.installed\", \"label\": \"已安装软件包\", \"description\": \"Debian/RPM 软件包清单\"}, {\"id\": \"packages.upgradable\", \"label\": \"本地可升级清单\", \"description\": \"仅使用已有元数据，不执行 apt update 或安装\"}, {\"id\": \"packages.detail\", \"label\": \"软件包详情\", \"description\": \"指定软件包的版本、描述和依赖\"}, {\"id\": \"packages.files\", \"label\": \"包内文件\", \"description\": \"指定软件包安装的文件\"}, {\"id\": \"packages.owner\", \"label\": \"文件所属软件包\", \"description\": \"查找绝对路径文件属于哪个软件包\"}, {\"id\": \"packages.history\", \"label\": \"包管理日志\", \"description\": \"读取 Debian dpkg 日志尾部\"}]}"]
:::
