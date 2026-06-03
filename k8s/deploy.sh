#!/usr/bin/env bash
# Kubernetes 滚动部署脚本
# 用法:
#   ./k8s/deploy.sh [IMAGE_TAG]
# 示例:
#   ./k8s/deploy.sh v1.2.3
#   ./k8s/deploy.sh                # 默认使用 git 短 commit 作为 tag
set -euo pipefail

NAMESPACE="emotionalbeach"
IMAGE_REPO="ghcr.io/eric-jxl/emotionalbeach"
K8S_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 1. 计算镜像 tag
TAG="${1:-$(git rev-parse --short HEAD 2>/dev/null || echo latest)}"
IMAGE="${IMAGE_REPO}:${TAG}"
echo "▶ 使用镜像: ${IMAGE}"

# 2. 构建并推送镜像
echo "▶ 构建镜像…"
cd ..
docker build -t "${IMAGE}" -f Dockerfile .
echo "▶ 推送镜像…"
docker push "${IMAGE}"
cd -

# 3. 应用基础资源（kustomize 一次性 apply 全部）
echo "▶ 应用 Kubernetes 清单…"
kubectl apply -k "${K8S_DIR}"

# 4. 滚动更新到新镜像（触发 RollingUpdate）
echo "▶ 滚动更新 Deployment 到 ${IMAGE}…"
kubectl -n "${NAMESPACE}" set image deployment/emotionalbeach \
  emotionalbeach="${IMAGE}" --record

# 5. 等待滚动完成；失败自动回滚
echo "▶ 等待滚动更新完成…"
if ! kubectl -n "${NAMESPACE}" rollout status deployment/emotionalbeach --timeout=180s; then
  echo "✗ 滚动更新失败，自动回滚到上一个版本"
  kubectl -n "${NAMESPACE}" rollout undo deployment/emotionalbeach
  kubectl -n "${NAMESPACE}" rollout status deployment/emotionalbeach --timeout=120s
  exit 1
fi

echo "✅ 部署完成：${IMAGE}"
kubectl -n "${NAMESPACE}" get pods -l app.kubernetes.io/name=emotionalbeach

