#!/bin/sh
# 如果 CONFIG_PATH 环境变量被设置，则使用它，否则使用默认值
if [ -z "$CONFIG_PATH" ]; then
  CONFIG_PATH="/app/config/.env"
fi

exec /app/emnotonalBeach -e "$CONFIG_PATH"