# Kubernetes

节点、Pod、工作负载、事件和日志。选择任务 Enter 查看原生命令和参数；进入普通任务页不会自动采集。

:::mdtui page
version: 1
id: ops-kubernetes
title: Kubernetes
parent: areas
max_parallel: 2
:::

:::mdtui panel
id: tasks
title: 任务清单 · F3 过滤
format: items-json
open:
  kubernetes.contexts: ./kubernetes-contexts.md
  kubernetes.nodes: ./kubernetes-nodes.md
  kubernetes.pods: ./kubernetes-pods.md
  kubernetes.deployments: ./kubernetes-deployments.md
  kubernetes.services: ./kubernetes-services.md
  kubernetes.events: ./kubernetes-events.md
  kubernetes.describe: ./kubernetes-describe.md
  kubernetes.logs: ./kubernetes-logs.md
exec: ["printf", "%s\n", "{\"version\": 1, \"items\": [{\"id\": \"kubernetes.contexts\", \"label\": \"集群上下文\", \"description\": \"读取 kubeconfig 上下文，不切换上下文\"}, {\"id\": \"kubernetes.nodes\", \"label\": \"节点清单\", \"description\": \"查询当前上下文的节点状态\"}, {\"id\": \"kubernetes.pods\", \"label\": \"命名空间 Pod\", \"description\": \"查询指定命名空间 Pod\"}, {\"id\": \"kubernetes.deployments\", \"label\": \"工作负载\", \"description\": \"查询指定命名空间 Deployment\"}, {\"id\": \"kubernetes.services\", \"label\": \"Service 清单\", \"description\": \"查询指定命名空间的 Service\"}, {\"id\": \"kubernetes.events\", \"label\": \"近期事件\", \"description\": \"查询命名空间事件，按时间排序\"}, {\"id\": \"kubernetes.describe\", \"label\": \"资源详情\", \"description\": \"查询当前命名空间的资源，例如 pod/example\"}, {\"id\": \"kubernetes.logs\", \"label\": \"Pod 日志\", \"description\": \"读取当前命名空间 Pod 日志，可跟随\"}]}"]
:::
