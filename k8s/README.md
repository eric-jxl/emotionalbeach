# Kubernetes 滚动部署

本目录提供 EmotionalBeach 的 Kubernetes 部署清单，默认采用 **RollingUpdate（滚动更新）** 策略实现零停机发布。

## 文件说明

| 文件                   | 作用                                                     |
|----------------------|--------------------------------------------------------|
| `namespace.yaml`     | 独立命名空间 `emotionalbeach`                                |
| `secret.yaml`        | 敏感配置（JWT / OAuth / DB / Redis 密码），通过环境变量覆盖 viper 配置    |
| `configmap.yaml`     | 非敏感 `config.yaml`，挂载到 `/app/config/config.yaml`        |
| `postgres.yaml`      | 集群内 PostgreSQL（StatefulSet + Service + PVC），生产建议用托管数据库 |
| `deployment.yaml`    | 应用 Deployment，含滚动更新策略、探针、优雅停机                          |
| `service.yaml`       | ClusterIP Service（80 → 8080）                           |
| `hpa.yaml`           | 基于 CPU/内存的水平自动扩缩（3~10 副本）                              |
| `pdb.yaml`           | PodDisruptionBudget，保证维护期间至少 2 副本可用                    |
| `ingress.yaml`       | Nginx Ingress 入口（按需替换域名）                               |
| `kustomization.yaml` | 聚合所有资源，便于一键 apply                                      |
| `deploy.sh`          | 构建镜像 + 推送 + 滚动更新 + 失败自动回滚                              |

## 滚动更新关键设计

- **策略**：`maxSurge: 1`、`maxUnavailable: 0`，升级时先起新 Pod、旧 Pod 全程可用，做到零停机。
- **就绪探针** `/health`：检查数据库等依赖，只有就绪的 Pod 才接收流量；这是滚动期间不丢请求的核心。
- **存活探针** `/ping`：仅探活进程，失败重启容器。
- **启动探针** `/ping`：保护慢启动，期间不触发存活/就绪。
- **优雅停机**：应用监听 `SIGTERM` 调用 `server.Shutdown`，配合 `preStop sleep 5`（等待 Service 摘流）与
  `terminationGracePeriodSeconds: 30`，避免连接被强杀。
- **PDB + 多副本 + 拓扑分散**：保证节点维护、自愈、扩缩容时的高可用。

## 快速开始

```bash
# 1. 一键部署全部资源
kubectl apply -k k8s/

# 2. 查看滚动状态
kubectl -n emotionalbeach rollout status deployment/emotionalbeach

# 3. 发布新版本（构建/推送/滚动/失败回滚）
./k8s/deploy.sh v1.0.0

# 4. 仅更新镜像触发滚动更新
kubectl -n emotionalbeach set image deployment/emotionalbeach \
  emotionalbeach=ghcr.io/eric-jxl/emotionalbeach:v1.0.0

# 5. 回滚到上一个版本
kubectl -n emotionalbeach rollout undo deployment/emotionalbeach

# 6. 查看历史
kubectl -n emotionalbeach rollout history deployment/emotionalbeach
```

## 首次部署：数据库迁移

应用支持 `-migrate` 自动迁移。首次部署可执行一次性 Job：

```bash
kubectl -n emotionalbeach run db-migrate --restart=Never --rm -it \
  --image=ghcr.io/eric-jxl/emotionalbeach:latest \
  --overrides='{"spec":{"containers":[{"name":"db-migrate","image":"ghcr.io/eric-jxl/emotionalbeach:latest","args":["-migrate"],"envFrom":[{"secretRef":{"name":"emotionalbeach-secret"}}],"volumeMounts":[{"name":"config","mountPath":"/app/config"}]}],"volumes":[{"name":"config","configMap":{"name":"emotionalbeach-config","items":[{"key":"config.yaml","path":"config.yaml"}]}}]}}'
```

## 安全提示

`secret.yaml` 中为示例明文，**请勿把真实密钥提交到仓库**。生产环境请使用 SealedSecrets / HashiCorp Vault / External
Secrets Operator 管理。

