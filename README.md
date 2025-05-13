# emotionalBeach

## Getting started

```shell
#Hot loading
go install github.com/zzwx/fresh@latest  # latest release.
#generate a sample settings file either at "./.fresh.yaml" or at specified by -c location
fresh -generate  # 或者fresh -g
fresh -c .fresh.yaml
``` 

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
