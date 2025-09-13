# Gamersky 多页爬取解决方案

由于 Colly 无法模拟点击和执行 JavaScript，我们需要使用其他方法来实现多页爬取。

## 当前限制

Colly 是一个**静态HTML爬虫**，不支持：
- ❌ JavaScript 渲染
- ❌ 用户交互（点击、滚动）
- ❌ AJAX 动态加载
- ❌ SPA 应用

## 解决方案

### 方案1：分析AJAX请求 ⭐ (推荐)

1. **使用浏览器开发者工具**分析"点击加载更多"的网络请求
2. **直接调用API接口**获取JSON数据
3. **解析JSON**并存储到数据库

#### 实现步骤：
```bash
# 1. 打开浏览器开发者工具 (F12)
# 2. 切换到 Network 标签
# 3. 点击"加载更多"按钮
# 4. 查看XHR/Fetch请求
# 5. 复制请求URL和参数
```

#### 示例API请求：
```go
func (gnc *GamerskyNewsCrawler) crawlAjaxPage(page int) (int, error) {
    // 分析后可能类似这样的请求
    url := fmt.Sprintf("https://wap.gamersky.com/api/loadmore?page=%d", page)
    
    client := &http.Client{}
    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("X-Requested-With", "XMLHttpRequest")
    
    resp, err := client.Do(req)
    // ... 处理响应JSON
}
```

### 方案2：使用浏览器自动化 🤖

使用 Rod 或 Chromedp 控制真实浏览器：

```go
import "github.com/go-rod/rod"

func (gnc *GamerskyNewsCrawler) crawlWithBrowser(pages int) error {
    browser := rod.New().MustConnect()
    defer browser.MustClose()
    
    page := browser.MustPage("https://wap.gamersky.com/")
    
    for i := 1; i < pages; i++ {
        // 等待加载更多按钮
        loadMore := page.MustElement("a.clickLoadMoreBtn")
        
        // 模拟点击
        loadMore.MustClick()
        
        // 等待新内容加载
        page.MustWaitLoad()
        
        // 提取新闻数据
        // ...
    }
}
```

### 方案3：模拟HTTP请求 📡

手动构造加载更多的HTTP请求：

```go
func (gnc *GamerskyNewsCrawler) loadMoreNews(pageNum int) (int, error) {
    // 需要分析实际的请求参数
    data := url.Values{}
    data.Set("page", strconv.Itoa(pageNum))
    data.Set("templatekey", "wap_index")
    data.Set("nodeid", "21036")
    
    resp, err := http.PostForm("https://wap.gamersky.com/loadmore", data)
    // ... 处理响应
}
```

## 推荐实现步骤

1. **第一步**：使用浏览器分析AJAX请求
   - 打开 https://wap.gamersky.com/
   - F12 开发者工具 → Network
   - 点击"点击加载更多"
   - 找到相关的XHR请求

2. **第二步**：实现API调用
   - 复制请求URL、Headers、参数
   - 用Go HTTP客户端模拟请求
   - 解析返回的JSON/HTML

3. **第三步**：集成到现有爬虫
   - 修改 `crawlAjaxPage` 方法
   - 添加JSON解析逻辑
   - 保持数据库操作不变

## 当前状态

- ✅ 第一页爬取：使用Colly正常工作
- 🚧 多页爬取：需要分析AJAX请求
- ✅ 数据存储：SQLite去重机制完善
- ✅ 命令行工具：参数配置齐全

要实现完整的多页爬取，需要具体分析目标网站的AJAX接口。