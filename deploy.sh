#!/usr/bin/env bash
make gen build_backend_on_linux upx_bin
scp ./cmd/emotionalBeach test:/root/emo
echo "上传成功"
SERVICE_NAME="emotionalbeach:emotionalbeach_00"
ssh test "supervisorctl stop  '$SERVICE_NAME'; mv /root/emo /root/emotionalBeach; supervisorctl start '$SERVICE_NAME';"

SSH_EXIT_CODE=$?
if [ $SSH_EXIT_CODE -ne 0 ]; then
    echo "错误: 远程 SSH 命令执行失败，退出码: $SSH_EXIT_CODE"
    exit 1
fi

echo "所有步骤已完成。"
exit