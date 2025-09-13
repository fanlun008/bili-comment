# Gamersky 新闻爬虫

这是一个基于 Go 和 Colly + API 的 Gamersky 游戏天空网站新闻爬虫工具。

## ✨ 功能特性

- 🚀 **混合爬取策略**: 第一页使用 Colly 爬虫框架，后续页面使用官方API
- 📰 **完整多页支持**: 真正的多页爬取功能，支持任意页数
- 🗄️ **SQLite数据库**: 数据持久化存储
- 🔄 **智能去重**: 基于文章ID (data-sid) 自动去重
- 📊 **数据查询**: 支持查询已爬取的新闻
- ⏱️ **可配置延迟**: 防止对服务器造成压力
- 🎯 **精确数据提取**: 提取标题、时间、链接、图片等完整信息

## 🏗️ 技术架构

### 爬取策略
- **第1页**: 使用 Colly 爬取静态HTML页面
- **第2页+**: 直接调用官方API `https://appapi2.gamersky.com/v6/GetWapIndex`

### API请求格式
```json
{
    "request": {
        "pageSize": 15,
        "cacheTime": 1,
        "pageIndex": 2
    }
}
```

## 🚀 安装和使用

### 构建项目

```bash
cd /path/to/bili-comment
CGO_ENABLED=1 go build -o bili-comment .
```

### 基本用法

```bash
# 爬取第一页新闻
./bili-comment gamersky

# 爬取多页新闻 (真正的多页支持!)
./bili-comment gamersky --pages=5

# 设置请求延迟
./bili-comment gamersky --delay=2s

# 指定输出数据库路径
./bili-comment gamersky --output=/path/to/custom.db
```

### 查询已爬取的新闻

```bash
# 查询所有新闻
./bili-comment query-gamersky

# 限制查询结果数量
./bili-comment query-gamersky --limit=10

# 指定数据库路径
./bili-comment query-gamersky --output=/path/to/custom.db
```

## 📋 命令行参数

### gamersky 命令

- `--pages int`: 爬取页数 (默认: 1, 支持多页)
- `--output string`: 输出数据库文件路径 (默认: "./data/gamersky.db")
- `--delay duration`: 请求间隔时间 (默认: 1s)

### query-gamersky 命令

- `--limit int`: 限制查询结果数量 (默认: 20, 0=无限制)
- `--output string`: 数据库文件路径 (默认: "./data/gamersky.db")

## 📊 数据结构

爬虫会提取以下新闻信息：

### 从第一页 (Colly)
- **SID**: 新闻唯一标识符 (data-sid)
- **Title**: 新闻标题
- **Time**: 发布时间
- **CommentNum**: 评论数量
- **URL**: 新闻链接
- **ImageURL**: 新闻图片链接
- **TopLineTime**: 置顶时间
- **CreateTime**: 记录创建时间

### 从API页面
- **ArticleID**: 文章ID (转换为SID)
- **Title**: 新闻标题
- **WapTopLineTimeTodayLabel**: 今日时间标签
- **WapArticleUrl**: 文章链接
- **WapSanTuArticlePic**: 图片HTML (自动提取URL)
- **TopLineTime**: 完整置顶时间

## 🗄️ 数据库表结构

```sql
CREATE TABLE gamersky_news (
    sid TEXT PRIMARY KEY,           -- 新闻ID
    title TEXT NOT NULL,            -- 标题
    time TEXT,                      -- 发布时间
    comment_num INTEGER DEFAULT 0,  -- 评论数
    url TEXT,                       -- 链接
    image_url TEXT,                 -- 图片链接
    topline_time TEXT,              -- 置顶时间
    create_time TEXT DEFAULT CURRENT_TIMESTAMP  -- 创建时间
);
```

## 🛠️ 技术栈

- **Go**: 编程语言
- **Colly**: 网页爬虫框架 (第一页)
- **HTTP Client**: 原生HTTP客户端 (API调用)
- **SQLite**: 轻量级数据库
- **Cobra**: 命令行工具框架

## 📝 实际运行示例

```bash
$ ./bili-comment gamersky --pages=3

2025/09/13 11:08:25 Gamersky新闻爬虫启动...
2025/09/13 11:08:25 开始爬取Gamersky新闻，页数：3
2025/09/13 11:08:25 请求延迟：1s
2025/09/13 11:08:25 正在爬取第 1 页...
2025/09/13 11:08:26 收到响应: https://wap.gamersky.com/, 状态码: 200, 大小: 40390 bytes
2025/09/13 11:08:26 爬取新闻: 2013307 - 库克回应iPhone Air续航质疑：非常出色 你会爱上它
...
2025/09/13 11:08:26 第 1 页爬取完成，新增 12 条新闻
2025/09/13 11:08:27 正在爬取第 2 页...
2025/09/13 11:08:28 API响应状态: 200, 大小: 7428 bytes
2025/09/13 11:08:28 API爬取新闻: 2013069 - 《LOL》新英雄曝光：带翼暗裔 被动叠满或可复活
...
2025/09/13 11:08:28 第 2 页爬取完成，新增 15 条新闻
2025/09/13 11:08:29 第 3 页爬取完成，新增 15 条新闻
2025/09/13 11:08:29 爬取完成！总共爬取 42 条新闻，已保存到 SQLite 数据库
```

## ⚡ 性能特点

- **高效去重**: 使用 `INSERT OR IGNORE` 避免重复数据
- **智能策略**: 第一页保持完整信息，后续页面使用高效API
- **错误处理**: 单页失败不影响整体爬取进程
- **请求控制**: 可配置延迟防止服务器压力

## 🔧 注意事项

1. **请求频率**: 默认请求延迟为1秒，请适当设置
2. **去重机制**: 基于文章ID进行去重，同一文章不会重复存储
3. **API稳定性**: 后续页面依赖官方API，如API变更需要更新代码
4. **数据一致性**: 第一页和API页面的数据格式略有不同，已做统一处理

## 🎯 功能完成度

- ✅ **第一页爬取**: Colly静态页面爬取 
- ✅ **多页爬取**: 官方API调用
- ✅ **数据去重**: 基于SID的智能去重
- ✅ **数据存储**: SQLite持久化
- ✅ **查询功能**: 灵活的数据查询
- ✅ **错误处理**: 完善的异常处理
- ✅ **命令行工具**: 用户友好的CLI

这个爬虫已经是一个**生产就绪**的完整解决方案！ 🚀