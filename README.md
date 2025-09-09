# B站评论爬虫命令行工具

一个基于 Cobra 的命令行工具，用于爬取B站视频评论数据。

## 功能特性

- 支持爬取一级和二级评论
- 可配置爬取模式（最新/热门评论）
- 支持自定义输出路径
- 可配置请求间隔避免反爬
- 支持二级评论页数限制
- 数据存储到 SQLite 数据库

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

数据存储在SQLite数据库中，表结构如下：

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

# 查看爬取命令帮助
./bili-comment crawl --help

# 查看查询命令帮助
./bili-comment query --help
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

# 运行示例
make example
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
│   └── crawl.go      # 爬取命令
├── crawler/          # 爬虫核心逻辑
│   └── crawler.go    # 爬虫实现
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
