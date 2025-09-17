# GitHub Actions 自动化爬取部署指南

本项目支持通过GitHub Actions进行自动化的周期性爬取，无需本地服务器24小时运行。

## 🚀 快速部署

### 1. 启用GitHub Actions

1. 确保你的代码已推送到GitHub仓库
2. 进入仓库的 `Actions` 标签页
3. 如果是第一次使用，点击 `I understand my workflows, go ahead and enable them`

### 2. 配置工作流

项目包含两个预配置的工作流：

#### 基础工作流 (`gamersky-crawler.yml`)
- ✅ 每5分钟执行一次
- ✅ 支持手动触发
- ✅ 自动上传爬取数据

#### 灵活工作流 (`gamersky-flexible.yml`) 🌟 推荐
- ✅ 智能调度：工作时间高频，非工作时间低频
- ✅ 多种模式：普通模式、深度模式、测试模式
- ✅ 自动清理：定期删除旧数据
- ✅ 详细报告：生成爬取统计报告

## 📋 定时策略说明

### 灵活工作流的定时策略

| 时间段 | 频率 | 页数 | 说明 |
|--------|------|------|------|
| 工作日 9:00-18:00 | 每10分钟 | 3页 | 高频爬取，及时获取热点新闻 |
| 工作日其他时间 | 每30分钟 | 2页 | 中频爬取，保持数据更新 |
| 周末全天 | 每30分钟 | 2页 | 低频爬取，节省资源 |
| 每天凌晨2点 | 每天1次 | 10页 | 深度爬取，获取更多历史数据 |

### 自定义定时策略

修改 `.github/workflows/gamersky-flexible.yml` 文件中的 `schedule` 部分：

```yaml
schedule:
  # 每5分钟（高频）
  - cron: '*/5 * * * *'
  
  # 每15分钟（中频）
  - cron: '*/15 * * * *'
  
  # 每小时（低频）
  - cron: '0 * * * *'
  
  # 每天8点和20点
  - cron: '0 8,20 * * *'
  
  # 仅工作日执行
  - cron: '0 9 * * 1-5'
```

## 🎮 手动执行

### 通过GitHub网页界面

1. 进入仓库的 `Actions` 标签页
2. 选择 `Gamersky News Crawler (Flexible Schedule)`
3. 点击右侧的 `Run workflow` 按钮
4. 选择参数：
   - **爬取模式**: `normal`（3页）、`deep`（10页）、`test`（1页）
   - **爬取页数**: 可以覆盖默认页数
   - **请求延迟**: 如 `1s`、`2s`、`500ms`

### 通过GitHub CLI

```bash
# 安装 GitHub CLI
gh auth login

# 手动触发普通模式
gh workflow run "gamersky-flexible.yml" \
  --field crawl_mode=normal \
  --field pages=3 \
  --field delay=1s

# 手动触发深度模式
gh workflow run "gamersky-flexible.yml" \
  --field crawl_mode=deep \
  --field pages=10 \
  --field delay=2s
```

## 📊 数据管理

### 查看爬取结果

1. 进入 `Actions` 标签页
2. 点击任意一次运行记录
3. 在页面底部找到 `Artifacts` 部分
4. 下载 `gamersky-xxx-xxx` 压缩包
5. 解压后获得 SQLite 数据库文件和报告

### 数据文件说明

```
gamersky-normal-123.zip
├── gamersky_normal_20240918_143022.db  # SQLite数据库
└── crawl_report.txt                    # 爬取报告
```

### 自动清理策略

- **Artifacts保留期**: 30天自动删除
- **旧数据清理**: 每次定时任务会清理7天前的artifacts
- **存储限制**: GitHub免费账户有存储配额限制

## ⚙️ 高级配置

### 1. 设置Secrets（敏感信息）

如果需要配置数据库连接、API密钥等：

1. 进入仓库 `Settings` > `Secrets and variables` > `Actions`
2. 点击 `New repository secret`
3. 添加所需的环境变量

```yaml
# 在工作流中使用
env:
  DATABASE_URL: ${{ secrets.DATABASE_URL }}
  API_KEY: ${{ secrets.API_KEY }}
```

### 2. 配置通知

#### 企业微信通知示例

```yaml
- name: 发送企业微信通知
  if: success()
  run: |
    curl -X POST "${{ secrets.WECHAT_WEBHOOK }}" \
      -H "Content-Type: application/json" \
      -d '{
        "msgtype": "text",
        "text": {
          "content": "✅ Gamersky爬取成功\n时间: $(date)\n页数: ${{ needs.determine-params.outputs.pages }}"
        }
      }'
```

#### 邮件通知示例

```yaml
- name: 发送邮件通知
  uses: dawidd6/action-send-mail@v3
  with:
    server_address: smtp.gmail.com
    server_port: 587
    username: ${{ secrets.EMAIL_USERNAME }}
    password: ${{ secrets.EMAIL_PASSWORD }}
    subject: "Gamersky爬取报告"
    body: file://crawl_report.txt
    to: your-email@example.com
```

### 3. 数据同步到外部存储

#### 同步到阿里云OSS

```yaml
- name: 同步到阿里云OSS
  env:
    OSS_ACCESS_KEY_ID: ${{ secrets.OSS_ACCESS_KEY_ID }}
    OSS_ACCESS_KEY_SECRET: ${{ secrets.OSS_ACCESS_KEY_SECRET }}
  run: |
    # 安装ossutil
    wget https://gosspublic.alicdn.com/ossutil/1.7.15/ossutil64
    chmod +x ossutil64
    
    # 配置OSS
    ./ossutil64 config -e oss-cn-beijing.aliyuncs.com \
      -i $OSS_ACCESS_KEY_ID \
      -k $OSS_ACCESS_KEY_SECRET
    
    # 上传数据
    ./ossutil64 cp data/ oss://your-bucket/gamersky/ -r
```

## 🛠️ 故障排除

### 常见问题

1. **工作流不执行**
   - 检查仓库是否启用了Actions
   - 确认cron表达式格式正确
   - 检查仓库是否有最近的活动（GitHub可能暂停不活跃仓库的定时任务）

2. **爬取失败**
   - 查看Actions运行日志
   - 检查网络连接问题
   - 确认目标网站是否可访问

3. **存储空间不足**
   - 检查GitHub存储配额
   - 调整数据保留策略
   - 考虑同步到外部存储

### 调试技巧

1. **启用详细日志**
   ```yaml
   env:
     ACTIONS_RUNNER_DEBUG: true
     ACTIONS_STEP_DEBUG: true
   ```

2. **测试模式运行**
   ```bash
   # 手动触发测试模式
   gh workflow run "gamersky-flexible.yml" --field crawl_mode=test
   ```

3. **本地测试**
   ```bash
   # 本地测试单次执行命令
   CGO_ENABLED=1 go run main.go gamersky-once --pages=1
   ```

## 📈 监控和优化

### 性能监控

在工作流中添加性能监控：

```yaml
- name: 性能监控
  run: |
    echo "=== 系统资源使用情况 ==="
    df -h
    free -h
    echo "=== 爬取任务资源使用 ==="
    time go run main.go gamersky-once --pages=1
```

### 成本优化

1. **调整执行频率**: 根据需求降低执行频率
2. **使用条件执行**: 仅在有新内容时执行
3. **优化数据存储**: 压缩数据库，删除重复内容

## 🔄 升级和维护

### 依赖更新

```yaml
# 添加依赖更新检查
- name: 检查Go模块更新
  run: |
    go list -u -m all
    go mod tidy
```

### 定期维护任务

```yaml
# 每周运行维护任务
schedule:
  - cron: '0 0 * * 0'  # 每周日午夜

jobs:
  maintenance:
    runs-on: ubuntu-latest
    steps:
    - name: 数据库优化
      run: |
        # 数据库VACUUM，去重等操作
        sqlite3 data/gamersky.db "VACUUM;"
```

通过以上配置，你的GitHub Actions爬虫将能够：
- 🔄 自动化周期性执行
- 📊 生成详细的爬取报告  
- 💾 安全存储和管理数据
- 🔔 及时通知爬取状态
- 🛠️ 便于调试和维护

现在你可以享受完全自动化的新闻爬取服务了！