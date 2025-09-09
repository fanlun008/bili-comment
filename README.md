# B站评论爬虫和视频搜索命令行工具

一个基于 Cobra 的命令行工具，用于爬取B站视频评论数据和搜索B站视频。

## 功能特性

- 支持爬取一级和二级评论
- 支持根据关键词搜索B站视频
- 可配置爬取模式（最新/热门评论）
- 支持自定义输出路径
- 可配置请求间隔避免反爬
- 支持二级评论页数限制
- 数据存储到 SQLite 数据库
- 支持查询已保存的视频和评论数据

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

# 构建项目
go build

# 或者直接运行
go run main.go
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

### 搜索视频

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

### 查询视频

#### 基本查询

```bash
# 查询所有视频记录（默认显示前20条）
./bili-comment query-videos

# 显示前10条视频记录
./bili-comment query-videos --list=10

# 查询特定关键词的视频
./bili-comment query-videos --keyword=极氪001
```

#### 条件查询

```bash
# 查询特定关键词的前5条视频
./bili-comment query-videos --keyword=特斯拉 --list=5

# 指定数据库路径查询
./bili-comment query-videos --output=/path/to/videos.db
```

### 爬取评论

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

#### 组合参数

```bash
# 综合使用多个参数
./bili-comment crawl BV1HW4y1n7BF \
  --mode=3 \
  --with-replies=true \
  --delay=2s \
  --max-pages=3 \
  --output=./custom/path/comments.db
```

### 查询评论

#### 基本查询

```bash
# 统计评论总数
./bili-comment query --count

# 显示前10条评论
./bili-comment query --list=10

# 显示前20条评论
./bili-comment query --list=20
```

#### 条件查询

```bash
# 查询指定视频的评论数量
./bili-comment query --count --bv=BV1HW4y1n7BF

# 查询指定视频的评论列表
./bili-comment query --list=10 --bv=BV1HW4y1n7BF

# 查询指定用户的评论
./bili-comment query --list=5 --user="用户名"

# 指定数据库路径查询
./bili-comment query --count --db=/path/to/comments.db
```

## 参数说明

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--mode` | int | 2 | 爬取模式（2=最新评论，3=热门评论） |
| `--with-replies` | bool | true | 是否爬取二级评论 |
| `--max-pages` | int | 10 | 二级评论最大页数限制（0=无限制） |
| `--output` | string | "./data/crawler.db" | 输出数据库文件路径 |
| `--cookie` | string | "" | Cookie文件路径（为空时自动查找） |
| `--delay` | duration | 500ms | 请求间隔时间 |

## Cookie 配置

工具需要B站的Cookie来访问API。Cookie文件应包含B站的认证信息。

### 自动查找Cookie文件

如果没有指定 `--cookie` 参数，工具会按以下顺序查找：
1. 当前目录下的 `bili_cookie.txt`
2. `py-crawler/bili_cookie.txt`

### Cookie文件格式

Cookie文件应该是纯文本文件，包含完整的Cookie字符串，例如：
```
SESSDATA=xxx; buvid3=xxx; DedeUserID=xxx; ...
```

## 数据库结构

数据存储在SQLite数据库中，包含两个主要表：

### 视频搜索结果表 (bilibili_videos)

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

### 评论表 (bilibili_comments)

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

## 查看帮助

```bash
# 查看主命令帮助
./bili-comment --help

# 查看搜索命令帮助
./bili-comment search --help

# 查看视频查询命令帮助
./bili-comment query-videos --help

# 查看爬取命令帮助
./bili-comment crawl --help

# 查看评论查询命令帮助
./bili-comment query --help
```

## 完整工作流程示例

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

1. 请合理设置请求延迟，避免被B站反爬机制识别
2. 请确保Cookie文件的有效性
3. 大量爬取时建议增加延迟时间
4. 请遵守B站的使用条款和robots.txt规则

## 项目结构

```
bili-comment/
├── main.go           # 程序入口
├── cmd/              # Cobra命令定义
│   ├── root.go       # 根命令
│   ├── crawl.go      # 爬取命令
│   ├── search.go     # 搜索命令
│   ├── query.go      # 评论查询命令
│   └── query_videos.go # 视频查询命令
├── crawler/          # 爬虫核心逻辑
│   └── crawler.go    # 爬虫和搜索实现
├── data/             # 数据存储目录
│   └── crawler.db    # SQLite数据库
├── py-crawler/       # Python版本参考
└── README.md         # 项目文档
```

## 依赖

- [Cobra](https://github.com/spf13/cobra) - CLI框架
- [go-sqlite3](https://github.com/mattn/go-sqlite3) - SQLite驱动

## License

[添加你的许可证信息]
