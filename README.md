# 多平台爬虫命令行工具

一个基于 Cobra 的多功能命令行爬虫工具，支持B站评论/视频搜索和Gamersky新闻/评论爬取。

## 功能特性

### B站模块
- 支持爬取一级和二级评论
- 支持根据关键词搜索B站视频
- 可配置爬取模式（最新/热门评论）
- 支持自定义输出路径
- 可配置请求间隔避免反爬
- 支持二级评论页数限制

### Gamersky模块  
- **新闻爬取**: 支持多页新闻爬取，自动去重
- **评论爬取**: 支持文章评论和回复爬取
- **混合架构**: 第一页使用Colly，后续页面使用官方API
- **完整数据**: 包含用户等级、IP位置、设备信息等详细数据

### 通用特性
- 数据存储到 SQLite 数据库
- 支持查询已保存的数据
- 完善的错误处理和日志记录
- 跨平台支持

## 安装

### 方法1：使用 Makefile（推荐）

```bash
# 克隆项目
git clone <repository-url>
cd bili-comment

# 构建项目
make build

# 或安装到系统路径
make install
```

### 方法2：手动构建

```bash
# 克隆项目
git clone <repository-url>
cd bili-comment

# 构建项目（需要CGO支持SQLite）
CGO_ENABLED=1 go build

# 或者直接运行
CGO_ENABLED=1 go run main.go
```

### 方法3：交叉编译

```bash
# 构建所有平台版本
make build-all

# 或单独构建某个平台
make build-linux    # Linux
make build-windows  # Windows
make build-darwin   # macOS
```

## 使用方法

### Gamersky新闻爬取

#### 基本用法

```bash
# 爬取第1页新闻
CGO_ENABLED=1 go run main.go gamersky

# 爬取前5页新闻
CGO_ENABLED=1 go run main.go gamersky --pages=5

# 设置请求延迟
CGO_ENABLED=1 go run main.go gamersky --pages=3 --delay=2s

# 指定输出路径
CGO_ENABLED=1 go run main.go gamersky --output=/tmp/gamersky.db
```

#### 查询新闻

```bash
# 查询所有新闻（默认20条）
CGO_ENABLED=1 go run main.go query-gamersky

# 查询指定数量的新闻
CGO_ENABLED=1 go run main.go query-gamersky --limit=50

# 指定数据库路径查询
CGO_ENABLED=1 go run main.go query-gamersky --output=/tmp/gamersky.db --limit=10
```

### Gamersky评论爬取

#### 基本用法

```bash
# 爬取指定文章的评论
CGO_ENABLED=1 go run main.go gamersky-comments --article-id=2014209

# 爬取前5页评论
CGO_ENABLED=1 go run main.go gamersky-comments --article-id=2014209 --pages=5

# 设置请求延迟
CGO_ENABLED=1 go run main.go gamersky-comments --article-id=2014209 --delay=1s

# 指定输出路径
CGO_ENABLED=1 go run main.go gamersky-comments --article-id=2014209 --output=/tmp/comments.db
```

#### 查询评论

```bash
# 查询所有评论（默认20条）
CGO_ENABLED=1 go run main.go query-gamersky-comments

# 查询指定文章的评论
CGO_ENABLED=1 go run main.go query-gamersky-comments --article-id=2014209

# 查询指定数量的评论
CGO_ENABLED=1 go run main.go query-gamersky-comments --limit=50

# 查询指定文章的前100条评论
CGO_ENABLED=1 go run main.go query-gamersky-comments --article-id=2014209 --limit=100
```

### B站视频搜索

#### 基本用法

```bash
# 搜索关键词相关的视频
./bili-comment search 极氪001

# 搜索指定页数的结果
./bili-comment search 极氪001 --page=2

# 设置每页结果数量
./bili-comment search 极氪001 --page-size=10
```

#### 高级用法

```bash
# 设置请求延迟
./bili-comment search 极氪001 --delay=1s

# 指定输出数据库路径
./bili-comment search 极氪001 --output=/tmp/videos.db

# 指定Cookie文件路径
./bili-comment search 极氪001 --cookie=./my_cookie.txt
```

### B站视频查询

#### 基本查询

```bash
# 查询所有视频记录（默认显示前20条）
./bili-comment query-videos

# 显示前10条视频记录
./bili-comment query-videos --list=10

# 查询特定关键词的视频
./bili-comment query-videos --keyword=极氪001
```

### B站评论爬取

#### 基本用法

```bash
# 爬取指定BV号的评论（默认最新评论，包含二级评论）
./bili-comment crawl BV1HW4y1n7BF
```

#### 高级用法

```bash
# 爬取热门评论
./bili-comment crawl BV1HW4y1n7BF --mode=3

# 不爬取二级评论
./bili-comment crawl BV1HW4y1n7BF --with-replies=false

# 设置请求延迟为1秒（避免反爬）
./bili-comment crawl BV1HW4y1n7BF --delay=1s

# 指定输出数据库路径
./bili-comment crawl BV1HW4y1n7BF --output=/tmp/comments.db

# 指定Cookie文件路径
./bili-comment crawl BV1HW4y1n7BF --cookie=./my_cookie.txt

# 限制二级评论最大页数
./bili-comment crawl BV1HW4y1n7BF --max-pages=5
```

### B站评论查询

```bash
# 统计评论总数
./bili-comment query --count

# 显示前10条评论
./bili-comment query --list=10

# 查询指定视频的评论
./bili-comment query --list=10 --bv=BV1HW4y1n7BF

# 查询指定用户的评论
./bili-comment query --list=5 --user="用户名"
```

## 参数说明

### Gamersky模块参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--pages` | int | 1 | 爬取页数 |
| `--delay` | duration | 1s | 请求间隔时间 |
| `--output` | string | "./data/gamersky.db" | 输出数据库文件路径 |
| `--article-id` | string | "" | 文章ID（评论爬取必需） |
| `--limit` | int | 20 | 查询结果限制数量 |

### B站模块参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--mode` | int | 2 | 爬取模式（2=最新评论，3=热门评论） |
| `--with-replies` | bool | true | 是否爬取二级评论 |
| `--max-pages` | int | 10 | 二级评论最大页数限制（0=无限制） |
| `--output` | string | "./data/crawler.db" | 输出数据库文件路径 |
| `--cookie` | string | "" | Cookie文件路径（为空时自动查找） |
| `--delay` | duration | 500ms | 请求间隔时间 |

## 数据库结构

### Gamersky数据库 (gamersky.db)

#### 新闻表 (gamersky_news)

```sql
CREATE TABLE gamersky_news (
    sid TEXT PRIMARY KEY,           -- 新闻ID (data-sid)
    title TEXT NOT NULL,            -- 新闻标题
    time TEXT,                      -- 发布时间
    comment_num INTEGER DEFAULT 0,  -- 评论数
    url TEXT,                       -- 新闻链接
    image_url TEXT,                 -- 图片链接
    topline_time TEXT,              -- 置顶时间
    create_time TEXT DEFAULT CURRENT_TIMESTAMP -- 记录创建时间
);
```

#### 评论表 (gamersky_comments)

```sql
CREATE TABLE gamersky_comments (
    id INTEGER PRIMARY KEY,                    -- 评论ID
    article_id TEXT NOT NULL,                  -- 文章ID
    user_id INTEGER,                           -- 用户ID
    username TEXT,                             -- 用户名
    content TEXT,                              -- 评论内容
    comment_time TEXT,                         -- 评论时间
    support_count INTEGER DEFAULT 0,          -- 点赞数
    reply_count INTEGER DEFAULT 0,            -- 回复数
    parent_id INTEGER DEFAULT 0,              -- 父评论ID (0表示一级评论)
    user_avatar TEXT,                          -- 用户头像
    user_level INTEGER DEFAULT 0,             -- 用户等级
    ip_location TEXT,                          -- IP位置
    device_name TEXT,                          -- 设备名称
    floor_number INTEGER DEFAULT 0,           -- 楼层号
    is_tuijian BOOLEAN DEFAULT FALSE,         -- 是否推荐
    is_author BOOLEAN DEFAULT FALSE,          -- 是否作者
    is_best BOOLEAN DEFAULT FALSE,            -- 是否最佳
    user_authentication TEXT,                 -- 用户认证
    user_group_id INTEGER DEFAULT 0,          -- 用户组ID
    third_platform_bound TEXT,                -- 第三方平台绑定
    create_time TEXT DEFAULT CURRENT_TIMESTAMP, -- 记录创建时间
    UNIQUE(id, article_id)                     -- 防重复索引
);
```

### B站数据库 (crawler.db)

#### 视频搜索结果表 (bilibili_videos)

```sql
CREATE TABLE bilibili_videos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    keyword TEXT NOT NULL,           -- 搜索关键词
    bvid TEXT NOT NULL,             -- 视频BV号
    title TEXT,                     -- 视频标题
    author TEXT,                    -- 作者
    play INTEGER,                   -- 播放量
    video_review INTEGER,           -- 评论数
    favorites INTEGER,              -- 收藏数
    pubdate INTEGER,                -- 发布时间戳
    duration TEXT,                  -- 视频时长
    like_count INTEGER,             -- 点赞数
    danmaku INTEGER,                -- 弹幕数
    description TEXT,               -- 视频描述
    pic TEXT,                       -- 视频封面
    create_time TEXT,               -- 记录创建时间
    UNIQUE(keyword, bvid)           -- 防重复索引
);
```

#### 评论表 (bilibili_comments)

```sql
CREATE TABLE bilibili_comments (
    序号 INTEGER,
    上级评论ID INTEGER,
    评论ID INTEGER PRIMARY KEY,
    用户ID INTEGER,
    用户名 TEXT,
    用户等级 INTEGER,
    性别 TEXT,
    评论内容 TEXT,
    评论时间 TEXT,
    回复数 INTEGER,
    点赞数 INTEGER,
    个性签名 TEXT,
    IP属地 TEXT,
    是否是大会员 TEXT,
    头像 TEXT,
    视频BV号 TEXT,
    视频标题 TEXT
);
```

## Cookie 配置（仅B站模块）

B站模块需要Cookie来访问API。Cookie文件应包含B站的认证信息。

### 自动查找Cookie文件

如果没有指定 `--cookie` 参数，工具会按以下顺序查找：
1. 当前目录下的 `bili_cookie.txt`
2. `py-crawler/bili_cookie.txt`

### Cookie文件格式

Cookie文件应该是纯文本文件，包含完整的Cookie字符串，例如：
```
SESSDATA=xxx; buvid3=xxx; DedeUserID=xxx; ...
```

## 查看帮助

```bash
# 查看主命令帮助
./bili-comment --help

# 查看Gamersky新闻爬取帮助
./bili-comment gamersky --help

# 查看Gamersky评论爬取帮助
./bili-comment gamersky-comments --help

# 查看Gamersky新闻查询帮助
./bili-comment query-gamersky --help

# 查看Gamersky评论查询帮助
./bili-comment query-gamersky-comments --help

# 查看B站搜索命令帮助
./bili-comment search --help

# 查看B站视频查询命令帮助
./bili-comment query-videos --help

# 查看B站爬取命令帮助
./bili-comment crawl --help

# 查看B站评论查询命令帮助
./bili-comment query --help
```

## 完整工作流程示例

### Gamersky工作流程

```bash
# 1. 爬取Gamersky新闻
CGO_ENABLED=1 go run main.go gamersky --pages=3

# 2. 查看爬取的新闻
CGO_ENABLED=1 go run main.go query-gamersky --limit=10

# 3. 选择感兴趣的文章，爬取评论（使用查询结果中的SID作为article-id）
CGO_ENABLED=1 go run main.go gamersky-comments --article-id=2014209 --pages=5

# 4. 查看爬取的评论
CGO_ENABLED=1 go run main.go query-gamersky-comments --article-id=2014209 --limit=20
```

### B站工作流程

```bash
# 1. 搜索相关视频
./bili-comment search 极氪001 --page-size=5

# 2. 查看搜索结果
./bili-comment query-videos --keyword=极氪001

# 3. 选择感兴趣的视频，爬取评论（使用查询结果中的BVID）
./bili-comment crawl BV1QhWJeoEww

# 4. 查看爬取的评论
./bili-comment query --list=10 --bv=BV1QhWJeoEww
```

## 开发工具

项目提供了 Makefile 来简化开发流程：

```bash
# 查看所有可用命令
make help

# 构建项目
make build

# 清理构建文件
make clean

# 格式化代码
make fmt

# 检查代码
make vet

# 更新依赖
make deps

# 运行搜索示例
make search-example

# 运行查询示例
make query-example
```

## 注意事项

### 通用注意事项
1. 请合理设置请求延迟，避免被反爬机制识别
2. 大量爬取时建议增加延迟时间
3. 请遵守各网站的使用条款和robots.txt规则

### Gamersky模块
1. 第一页使用Colly爬取静态内容，后续页面使用官方API
2. 自动跳过无效的评论数据（ID为0或用户名为空）
3. 支持一级评论和二级回复的完整层级结构

### B站模块
1. 请确保Cookie文件的有效性
2. 评论爬取需要有效的B站登录状态

## 项目结构

```
bili-comment/
├── main.go                      # 程序入口
├── cmd/                         # Cobra命令定义
│   ├── root.go                  # 根命令
│   ├── crawl.go                 # B站评论爬取命令
│   ├── search.go                # B站视频搜索命令
│   ├── query.go                 # B站评论查询命令
│   ├── query_videos.go          # B站视频查询命令
│   ├── gamersky.go              # Gamersky新闻爬取命令
│   ├── gamersky_comments.go     # Gamersky评论爬取命令
│   ├── query_gamersky.go        # Gamersky新闻查询命令
│   └── query_gamersky_comments.go # Gamersky评论查询命令
├── crawler/                     # 爬虫核心逻辑
│   └── crawler.go               # 所有爬虫实现
├── data/                        # 数据存储目录
│   ├── crawler.db               # B站数据SQLite数据库
│   └── gamersky.db              # Gamersky数据SQLite数据库
├── py-crawler/                  # Python版本参考
└── README.md                    # 项目文档
```

## 技术架构

### Gamersky爬虫架构
- **混合爬取策略**: 第一页使用Colly框架爬取静态HTML，后续页面通过官方API获取JSON数据
- **API集成**: 使用 `https://appapi2.gamersky.com/v6/GetWapIndex` 获取新闻列表
- **评论系统**: 通过 `https://cm.gamersky.com/appapi/GetArticleCommentWithClubStyle` 获取文章评论
- **数据去重**: 基于SID和评论ID的数据库唯一约束

### B站爬虫架构  
- **API访问**: 使用B站官方Web API
- **认证机制**: 基于Cookie的用户认证
- **数据处理**: JSON响应解析和数据清洗
- **反爬策略**: 可配置请求延迟和用户代理

## 依赖

- [Cobra](https://github.com/spf13/cobra) - CLI框架
- [Colly](https://github.com/gocolly/colly) - Web爬虫框架
- [go-sqlite3](https://github.com/mattn/go-sqlite3) - SQLite驱动

## License

[添加你的许可证信息]
