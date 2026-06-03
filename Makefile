GOBUILD=go build
GOARCH=$(shell go env GOARCH)
GOOS=$(shell go env GOOS )

BASE_PAH := $(shell pwd)
BUILD_PATH = $(BASE_PAH)/cmd
MAIN= $(BASE_PAH)/main.go
APP_NAME=emotionalBeach

K8S_DIR=$(BASE_PAH)/k8s
K8S_NS=emotionalbeach

.PHONY: upx_bin build_backend clean build_backend_on_linux gen fmt \
        k8s-deploy k8s-apply k8s-status k8s-rollback k8s-history k8s-delete
all: build_backend upx_bin

gen:
	@go generate ./...
fmt:
	@go list -f {{.Dir}} ./... | xargs -I{} gofmt -w -s {}

upx_bin:
	upx $(BUILD_PATH)/$(APP_NAME)

build_backend:
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) -trimpath -ldflags '-s -w' -o $(BUILD_PATH)/$(APP_NAME) $(MAIN)
build_backend_on_linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -trimpath -ldflags '-s -w' -o $(BUILD_PATH)/$(APP_NAME) $(MAIN)

clean:
	rm -rf cmd/*

# ── Kubernetes 滚动部署 ───────────────────────────────────────────────────
# 一键应用全部清单
k8s-apply:
	kubectl apply -k $(K8S_DIR)
# 构建/推送/滚动更新/失败回滚（可传 TAG: make k8s-deploy TAG=v1.0.0）
k8s-deploy:
	$(K8S_DIR)/deploy.sh $(TAG)
# 查看滚动更新状态
k8s-status:
	kubectl -n $(K8S_NS) rollout status deployment/emotionalbeach
# 回滚到上一个版本
k8s-rollback:
	kubectl -n $(K8S_NS) rollout undo deployment/emotionalbeach
# 查看发布历史
k8s-history:
	kubectl -n $(K8S_NS) rollout history deployment/emotionalbeach
# 删除全部资源
k8s-delete:
	kubectl delete -k $(K8S_DIR)

