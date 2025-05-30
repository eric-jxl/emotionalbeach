# emotionalBeach
![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/eric-jxl/Go?color=blue&label=go&logo=go)
[![build-go-binary](https://github.com/eric-jxl/emotionalbeach/actions/workflows/go-binary-release.yml/badge.svg)](https://github.com/eric-jxl/emotionalbeach/actions/workflows/go-binary-release.yml)
[![Docker Image CI](https://github.com/eric-jxl/emotionalbeach/actions/workflows/docker-image.yml/badge.svg)](https://github.com/eric-jxl/emotionalbeach/actions/workflows/docker-image.yml)

## Getting started

```shell
# Hot loading
go install github.com/zzwx/fresh@latest  # latest release.
# generate a sample settings file either at "./.fresh.yaml" or at specified by -c location
fresh -generate  # 或者fresh -g
fresh -c .fresh.yaml
```
> [!TIP]  
> *新增docker ci/cd 打包发布到ghcr.io*
> 
> *创建release自动编译跨平台二进制包*

## Add your files

```shell
cd existing_repo
git remote add origin https://gitlab.com/eric-jxl/emotionalbeach.git
git branch -M main
git push -uf origin main

```
## 编译快捷命令
```shell
make all # 打包编译并upx压缩
make gen # 生成Swagger文档
```
## 快速启动
```shell
docker compose up -d
```
