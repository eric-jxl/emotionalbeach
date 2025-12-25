#!/bin/sh
# 如果 CONFIG_PATH 环境变量被设置，则使用它，否则使用默认值
set -e
if [ -z "$CONFIG_PATH" ]; then
  CONFIG_PATH="/app/config/config.yaml"
fi

# 判断文件是否存在
if [ ! -f "$CONFIG_PATH" ]; then
  echo "\x1b[31m错误：配置文件 $CONFIG_PATH 不存在\x1b[0m"
  exit 1
fi

exec /app/emnotonalBeach
