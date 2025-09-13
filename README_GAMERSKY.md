# Gamersky 新闻爬虫

这是一个基于 Go 和 Colly 库开发的 Gamersky 游戏天空网站新闻爬虫工具。

## 功能特性

- 🚀 使用 Colly 高性能爬虫框架
- 📰 爬取 Gamersky 网站新闻列表
- 🗄️ 数据存储到 SQLite 数据库
- 🔄 自动去重（基于 data-sid）
- 📊 支持查询已爬取的新闻
- ⏱️ 可配置请求延迟

## 安装和使用

### 构建项目

```bash
cd /path/to/bili-comment
CGO_ENABLED=1 go build -o bili-comment .
```

### 基本用法

```bash
# 爬取第一页新闻
./bili-comment gamersky

# 爬取多页新闻（目前仅支持第一页）
./bili-comment gamersky --pages=1

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

## 命令行参数

### gamersky 命令

- `--pages int`: 爬取页数 (默认: 1)
- `--output string`: 输出数据库文件路径 (默认: "./data/gamersky.db")
- `--delay duration`: 请求间隔时间 (默认: 1s)

### query-gamersky 命令

- `--limit int`: 限制查询结果数量 (默认: 20, 0=无限制)
- `--output string`: 数据库文件路径 (默认: "./data/gamersky.db")

## 数据结构

爬虫会提取以下新闻信息：

- **SID**: 新闻唯一标识符 (data-sid)
- **Title**: 新闻标题
- **Time**: 发布时间
- **CommentNum**: 评论数量
- **URL**: 新闻链接
- **ImageURL**: 新闻图片链接
- **TopLineTime**: 置顶时间
- **CreateTime**: 记录创建时间

## 数据库表结构

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

## 技术栈

- **Go**: 编程语言
- **Colly**: 网页爬虫框架
- **SQLite**: 数据库
- **Cobra**: 命令行工具框架

## 注意事项

1. **请求频率**: 默认请求延迟为1秒，请适当设置以避免对目标网站造成过大压力
2. **去重机制**: 基于新闻的 `data-sid` 属性进行去重
3. **多页支持**: 目前仅支持第一页爬取，多页功能需要进一步开发
4. **评论数**: 评论数可能是动态加载的，初始值可能为0

## 示例输出

```
2025/09/13 10:46:34 爬取新闻: 2013313 - 全球十大畅销手机：iPhone16霸榜前三 小米国产之光
2025/09/13 10:46:34 爬取新闻: 2013306 - 筋疲力尽！《小丑牌》开发者道歉：无法准时更新游戏
2025/09/13 10:46:34 爬取新闻: 2013118 - 有人在2025年掏出一个2013年的游戏
...
2025/09/13 10:46:34 第 1 页爬取完成，新增 11 条新闻
2025/09/13 10:46:34 爬取完成！总共爬取 11 条新闻，已保存到 SQLite 数据库
```

## 扩展功能

如果需要添加更多功能，可以考虑：

1. **多页支持**: 分析网站的 AJAX 请求，实现真正的多页爬取
2. **评论爬取**: 进一步爬取每条新闻的评论内容
3. **内容提取**: 爬取新闻的完整内容
4. **数据导出**: 支持导出为 JSON、CSV 等格式
5. **定时任务**: 支持定时自动爬取最新新闻

## 许可证

请遵守目标网站的 robots.txt 和使用条款，合理使用爬虫工具。