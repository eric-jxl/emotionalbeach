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

- [ ] [Create](https://docs.gitlab.com/ee/user/project/repository/web_editor.html#create-a-file)
  or [upload](https://docs.gitlab.com/ee/user/project/repository/web_editor.html#upload-a-file) files
- [ ] [Add files using the command line](https://docs.gitlab.com/ee/gitlab-basics/add-file.html#add-a-file-using-the-command-line)
  or push an existing Git repository with the following command:

```shell
cd existing_repo
git remote add origin https://gitlab.com/eric-jxl/emotionalbeach.git
git branch -M main
git push -uf origin main
```

### 第一期目标实现登陆注册,用户表等接口
