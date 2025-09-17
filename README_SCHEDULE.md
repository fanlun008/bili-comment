# Gamersky 定时爬取任务

本项目现在支持定时爬取Gamersky新闻的功能，可以设置周期性任务来持续爬取新闻数据。

## 功能特性

- ✅ 支持自定义Cron表达式定时执行
- ✅ 支持自定义爬取页数
- ✅ 支持自定义请求延迟
- ✅ 支持优雅关闭
- ✅ 详细的日志记录
- ✅ 立即执行 + 定时执行
- ✅ 防重复爬取（基于数据库去重）

## 快速开始

### 基本用法

```bash
# 每5分钟爬取1页新闻（默认设置）
CGO_ENABLED=1 go run main.go gamersky-schedule

# 每5分钟爬取3页新闻
CGO_ENABLED=1 go run main.go gamersky-schedule --pages=3

# 设置自定义输出路径
CGO_ENABLED=1 go run main.go gamersky-schedule --output=/path/to/your/database.db
```

### 自定义时间间隔

```bash
# 每1分钟执行一次
CGO_ENABLED=1 go run main.go gamersky-schedule --cron="0 */1 * * * *"

# 每10分钟执行一次
CGO_ENABLED=1 go run main.go gamersky-schedule --cron="0 */10 * * * *"

# 每小时执行一次
CGO_ENABLED=1 go run main.go gamersky-schedule --cron="0 0 * * * *"

# 每天凌晨2点执行
CGO_ENABLED=1 go run main.go gamersky-schedule --cron="0 0 2 * * *"

# 工作日每小时执行（周一到周五）
CGO_ENABLED=1 go run main.go gamersky-schedule --cron="0 0 * * * 1-5"
```

## Cron表达式格式

本项目使用6位Cron表达式（支持秒级精度）：

```
秒  分  时  日  月  周
*   *   *   *   *   *
```

### 常用表达式示例

| 表达式 | 说明 |
|--------|------|
| `0 */5 * * * *` | 每5分钟 |
| `0 */10 * * * *` | 每10分钟 |
| `0 */30 * * * *` | 每30分钟 |
| `0 0 * * * *` | 每小时 |
| `0 0 */2 * * *` | 每2小时 |
| `0 0 */6 * * *` | 每6小时 |
| `0 0 0 * * *` | 每天午夜 |
| `0 0 2 * * *` | 每天凌晨2点 |
| `0 0 0 * * 1` | 每周一午夜 |
| `0 0 0 1 * *` | 每月1日午夜 |

### 特殊字符说明

- `*` : 匹配任何值
- `?` : 匹配任何值（仅用于日期和星期）
- `-` : 范围（如 1-5 表示1到5）
- `,` : 枚举（如 1,3,5 表示1、3、5）
- `/` : 增量（如 */5 表示每5个单位）

## 命令行参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--pages` | 1 | 每次爬取的页数 |
| `--output` | `./data/gamersky.db` | 输出数据库文件路径 |
| `--delay` | `1s` | 请求间隔时间 |
| `--cron` | `0 */5 * * * *` | Cron表达式（每5分钟） |

## 使用示例

### 1. 开发测试（每分钟执行）

```bash
CGO_ENABLED=1 go run main.go gamersky-schedule \
  --cron="0 */1 * * * *" \
  --pages=1 \
  --output="./data/test.db"
```

### 2. 生产环境（每5分钟，爬取3页）

```bash
CGO_ENABLED=1 go run main.go gamersky-schedule \
  --cron="0 */5 * * * *" \
  --pages=3 \
  --delay=2s \
  --output="./data/production.db"
```

### 3. 夜间批量爬取（每晚2点，爬取10页）

```bash
CGO_ENABLED=1 go run main.go gamersky-schedule \
  --cron="0 0 2 * * *" \
  --pages=10 \
  --delay=1s
```

## 停止任务

定时任务支持优雅关闭，可以通过以下方式停止：

- 按 `Ctrl+C` 发送 SIGINT 信号
- 发送 SIGTERM 信号：`kill <pid>`

程序会等待当前正在执行的爬取任务完成后再退出。

## 日志输出

程序会输出详细的日志信息，包括：

- 定时任务配置信息
- 每次任务执行的开始和结束时间
- 爬取页面的详细信息
- 新增新闻的数量
- 任务执行耗时
- 错误信息（如果有）

## 部署建议

### 1. 使用 systemd（推荐）

创建 systemd 服务文件：

```ini
[Unit]
Description=Gamersky News Crawler
After=network.target

[Service]
Type=simple
User=your-user
WorkingDirectory=/path/to/your/project
Environment=CGO_ENABLED=1
ExecStart=/usr/local/go/bin/go run main.go gamersky-schedule --pages=3
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

### 2. 使用 Docker

```dockerfile
FROM golang:1.21-alpine

WORKDIR /app
COPY . .
RUN CGO_ENABLED=1 go build -o gamersky-crawler main.go

CMD ["./gamersky-crawler", "gamersky-schedule", "--pages=3"]
```

### 3. 使用 nohup

```bash
nohup CGO_ENABLED=1 go run main.go gamersky-schedule --pages=3 > gamersky.log 2>&1 &
```

## 注意事项

1. **数据库去重**: 程序会自动去重，相同ID的新闻不会重复插入
2. **网络延迟**: 建议设置适当的请求延迟，避免给目标网站造成过大压力
3. **磁盘空间**: 定期检查数据库文件大小，必要时进行数据清理
4. **错误处理**: 单个页面爬取失败不会影响其他页面，程序会继续执行
5. **资源消耗**: 定时任务会持续运行，注意监控CPU和内存使用情况

## 数据查询

爬取的数据可以使用以下命令查询：

```bash
# 查询所有新闻
CGO_ENABLED=1 go run main.go query-gamersky

# 查询指定数量的最新新闻
CGO_ENABLED=1 go run main.go query-gamersky --limit=10
```