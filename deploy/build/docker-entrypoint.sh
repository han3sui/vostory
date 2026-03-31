#!/bin/sh
set -e

echo "========================================="
echo "Vostory 容器启动"
echo "========================================="

CONF_FILE="config/docker.yml"
VERSION_FILE="/app/storage/.version"
CURRENT_VERSION="${APP_VERSION:-unknown}"

echo "配置文件: $CONF_FILE"
echo "当前版本: $CURRENT_VERSION"
echo ""

# 获取已初始化的版本
if [ -f "$VERSION_FILE" ]; then
    INSTALLED_VERSION=$(cat "$VERSION_FILE")
else
    INSTALLED_VERSION=""
fi

echo "已安装版本: ${INSTALLED_VERSION:-无}"

# 判断是否需要初始化
if [ -z "$INSTALLED_VERSION" ]; then
    echo ""
    echo "首次启动，执行数据库迁移..."

    if ./migration -conf "$CONF_FILE"; then
        echo ""
        echo "执行创建管理员..."
        echo "y" | ./createadmin -conf "$CONF_FILE" -user admin || echo "⚠ 创建管理员跳过（可能已存在）"

        mkdir -p /app/storage
        echo "$CURRENT_VERSION" > "$VERSION_FILE"
        echo ""
        echo "✓ 初始化成功 (版本: $CURRENT_VERSION)"
    else
        echo ""
        echo "❌ 数据库迁移失败"
        echo "  手动重试: docker exec <container> ./migration -conf $CONF_FILE"
        exit 1
    fi

elif [ "$INSTALLED_VERSION" != "$CURRENT_VERSION" ]; then
    echo ""
    echo "检测到版本升级: $INSTALLED_VERSION -> $CURRENT_VERSION"
    echo "执行数据库迁移..."

    if ./migration -conf "$CONF_FILE"; then
        echo "$CURRENT_VERSION" > "$VERSION_FILE"
        echo ""
        echo "✓ 数据库迁移成功 (版本: $CURRENT_VERSION)"
    else
        echo ""
        echo "❌ 数据库迁移失败"
        echo "  手动重试: docker exec <container> ./migration -conf $CONF_FILE"
        exit 1
    fi

else
    echo ""
    echo "版本一致，跳过初始化"
    echo "  强制重新初始化: docker exec <container> rm $VERSION_FILE && docker restart <container>"
fi

# 同步 API 数据
echo ""
echo "同步 API 数据..."
if [ -f "docs/swagger.json" ]; then
    if ./syncmenu -conf "$CONF_FILE"; then
        echo "✓ API 数据同步成功"
    else
        echo "⚠ API 数据同步失败（非致命错误，继续启动）"
    fi
else
    echo "⚠ docs/swagger.json 不存在，跳过 API 同步"
fi

echo ""
echo "启动服务..."
exec /usr/bin/supervisord -c /etc/supervisord.conf
