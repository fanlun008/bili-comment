#!/bin/bash

# GitHub仓库信息
REPO="fanlun008/bili-comment"
WORKFLOW_NAME="Gamersky News Crawler"

echo "🔍 正在查看最近的工作流运行..."

# 检查是否安装了 gh CLI
if ! command -v gh &> /dev/null; then
    echo "❌ GitHub CLI 未安装，请先安装："
    echo "brew install gh"
    exit 1
fi

# 检查是否已登录
if ! gh auth status &> /dev/null; then
    echo "❌ 请先登录 GitHub CLI："
    echo "gh auth login"
    exit 1
fi

# 显示最近的运行记录
echo "📋 最近的工作流运行："
gh run list --repo "$REPO" --workflow "$WORKFLOW_NAME" --limit 10

echo ""
read -p "🔢 请输入要下载的运行ID (或按回车下载最新的): " RUN_ID

# 创建下载目录
DOWNLOAD_DIR="./downloads/$(date +%Y%m%d_%H%M%S)"
mkdir -p "$DOWNLOAD_DIR"

if [ -z "$RUN_ID" ]; then
    echo "📥 正在下载最新的 artifacts..."
    gh run download --repo "$REPO" --dir "$DOWNLOAD_DIR"
else
    echo "📥 正在下载运行 ID: $RUN_ID 的 artifacts..."
    gh run download "$RUN_ID" --repo "$REPO" --dir "$DOWNLOAD_DIR"
fi

echo "✅ 下载完成！文件保存在: $DOWNLOAD_DIR"
echo "📁 查看下载的文件："
ls -la "$DOWNLOAD_DIR"

# 如果下载了数据库文件，显示基本信息
if ls "$DOWNLOAD_DIR"/*/*.db &> /dev/null; then
    echo ""
    echo "📊 数据库文件信息："
    for db in "$DOWNLOAD_DIR"/*/*.db; do
        echo "文件: $(basename "$db")"
        echo "大小: $(du -h "$db" | cut -f1)"
        if command -v sqlite3 &> /dev/null; then
            echo "记录数: $(sqlite3 "$db" "SELECT COUNT(*) FROM news;" 2>/dev/null || echo "无法查询")"
        fi
        echo "---"
    done
fi