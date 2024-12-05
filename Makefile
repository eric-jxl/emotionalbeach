GOBUILD=go build
GOARCH=$(shell go env GOARCH)
GOOS=$(shell go env GOOS )

BASE_PAH := $(shell pwd)
BUILD_PATH = $(BASE_PAH)/cmd
SERVER_PATH=$(BASE_PAH)/backend
MAIN= $(BASE_PAH)/main.go
APP_NAME=emotionalBeach
.PHONY: build_on_local
upx_bin:
	upx $(BUILD_PATH)/$(APP_NAME)

build_backend_on_linux:
	GOOS=linux GOARCH=$(GOARCH) $(GOBUILD) -trimpath -ldflags '-s -w' -o $(BUILD_PATH)/$(APP_NAME) $(MAIN)

build_backend_on_darwin:
	GOOS=darwin GOARCH=$(GOARCH) $(GOBUILD) -trimpath -ldflags '-s -w'  -o $(BUILD_PATH)/$(APP_NAME) $(MAIN)

build_backend_on_windows:
	GOOS=windows GOARCH=$(GOARCH) $(GOBUILD) -trimpath -ldflags '-s -w'  -o $(BUILD_PATH)/$(APP_NAME) $(MAIN)

build_all: build_backend_on_linux build_backend_on_darwin build_backend_on_windows

build_on_local: build_backend_on_darwin
