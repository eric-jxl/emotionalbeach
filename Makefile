GOBUILD=go build
GOARCH=$(shell go env GOARCH)
GOOS=$(shell go env GOOS )

BASE_PAH := $(shell pwd)
BUILD_PATH = $(BASE_PAH)/cmd
SERVER_PATH=$(BASE_PAH)/backend
MAIN= $(BASE_PAH)/main.go
APP_NAME=emotionalBeach

.PHONY: upx_bin build_backend clean build_backend_on_linux
upx_bin:
	upx $(BUILD_PATH)/$(APP_NAME)

build_backend:
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) -trimpath -ldflags '-s -w' -o $(BUILD_PATH)/$(APP_NAME) $(MAIN)
build_backend_on_linux:
	GOOS=linux GOARCH=$(GOARCH) $(GOBUILD) -trimpath -ldflags '-s -w' -o $(BUILD_PATH)/$(APP_NAME) $(MAIN)

clean:
	rm -rf cmd/*
