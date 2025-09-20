package gamersky

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

// APIRequest API请求结构体
type APIRequest struct {
	Request APIRequestData `json:"request"`
}

// APIRequestData API请求数据
type APIRequestData struct {
	PageSize  int `json:"pageSize"`
	CacheTime int `json:"cacheTime"`
	PageIndex int `json:"pageIndex"`
}

// APIResponse API响应结构体
type APIResponse struct {
	ErrorCode    int           `json:"errorCode"`
	ErrorMessage string        `json:"errorMessage"`
	WatchTimes   []interface{} `json:"watchTimes"`
	WatchTime    float64       `json:"watchTime"`
	Result       []APINewsItem `json:"result"`
}

// APINewsItem API返回的新闻项
type APINewsItem struct {
	WapShowType              string `json:"WapShowType"`
	ArticleID                int    `json:"ArticleID"`
	Title                    string `json:"Title"`
	WapTopLineTimeTodayLabel string `json:"WapTopLineTimeTodayLabel"`
	WapArticleUrl            string `json:"WapArticleUrl"`
	WapSanTuArticlePic       string `json:"WapSanTuArticlePic"`
	TopLineTime              string `json:"TopLineTime"`
}

// NewsCrawler Gamersky新闻爬虫结构体
type NewsCrawler struct {
	db     *sql.DB
	config *Config
}

// NewNewsCrawler 创建新的Gamersky新闻爬虫实例
func NewNewsCrawler(config *Config) (*NewsCrawler, error) {
	// 初始化数据库
	db, err := GetDBConnection(config.OutputPath)
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %v", err)
	}

	return &NewsCrawler{
		db:     db,
		config: config,
	}, nil
}

// CrawlNews 爬取Gamersky新闻
func (gnc *NewsCrawler) CrawlNews(page int) (int, error) {
	if page == 1 {
		// 第一页使用Colly爬取（保持原有逻辑）
		return gnc.crawlFirstPage()
	} else {
		// 后续页面使用API
		return gnc.crawlAPIPage(page)
	}
}

// crawlAPIPage 使用API爬取指定页面
func (gnc *NewsCrawler) crawlAPIPage(page int) (int, error) {
	// 构造API请求
	requestData := APIRequest{
		Request: APIRequestData{
			PageSize:  15,
			CacheTime: 1,
			PageIndex: page,
		},
	}

	// 序列化请求数据
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return 0, fmt.Errorf("序列化请求数据失败: %v", err)
	}

	// 创建HTTP客户端
	client := &http.Client{Timeout: 30 * time.Second}

	// 创建请求
	req, err := http.NewRequest("POST", "https://appapi2.gamersky.com/v6/GetWapIndex", strings.NewReader(string(jsonData)))
	if err != nil {
		return 0, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("读取响应失败: %v", err)
	}

	log.Printf("API响应状态: %d, 大小: %d bytes", resp.StatusCode, len(body))

	// 解析JSON响应
	var apiResponse APIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return 0, fmt.Errorf("解析JSON失败: %v", err)
	}

	// 检查API错误
	if apiResponse.ErrorCode != 0 {
		return 0, fmt.Errorf("API错误: %s (代码: %d)", apiResponse.ErrorMessage, apiResponse.ErrorCode)
	}

	// 处理新闻数据
	count := 0
	for _, item := range apiResponse.Result {
		// 转换为统一的新闻格式
		news := &NewsInfo{
			SID:         fmt.Sprintf("%d", item.ArticleID),
			Title:       item.Title,
			Time:        item.WapTopLineTimeTodayLabel,
			URL:         item.WapArticleUrl,
			TopLineTime: item.TopLineTime,
			CreateTime:  time.Now().Format("2006-01-02 15:04:05"),
		}

		// 从HTML图片标签中提取图片URL
		if item.WapSanTuArticlePic != "" {
			imgRegex := regexp.MustCompile(`src=['"]([^'"]*?)['"]`)
			if matches := imgRegex.FindStringSubmatch(item.WapSanTuArticlePic); len(matches) > 1 {
				news.ImageURL = matches[1]
			}
		}

		// 保存新闻到数据库（去重）
		if err := gnc.saveNewsToDB(news); err != nil {
			log.Printf("保存新闻失败 (SID: %s): %v", news.SID, err)
			continue
		}

		count++
		log.Printf("API爬取新闻: %s - %s", news.SID, news.Title)
	}

	log.Printf("API第 %d 页完成，共 %d 条新闻", page, len(apiResponse.Result))
	return count, nil
}

// crawlFirstPage 爬取第一页
func (gnc *NewsCrawler) crawlFirstPage() (int, error) {
	// 创建 Colly 收集器
	c := colly.NewCollector(
		// 设置允许的域名
		colly.AllowedDomains("wap.gamersky.com"),
	)

	// 设置用户代理和请求头
	c.UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36"

	// 设置请求延迟
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 1,
		Delay:       gnc.config.RequestDelay,
	})

	count := 0

	// 查找新闻列表项
	c.OnHTML("li[data-id]", func(e *colly.HTMLElement) {
		// 提取新闻信息
		news := &NewsInfo{
			SID:         e.Attr("data-id"),
			TopLineTime: e.Attr("data-toplinetime"),
			CreateTime:  time.Now().Format("2006-01-02 15:04:05"),
		}

		// 提取标题 - 根据当前li元素的布局类型来决定
		var newsType string

		// 检查是否为 titleAndTime 布局
		titleElement := e.DOM.Find(".titleAndTime h5")
		if titleElement.Length() > 0 {
			news.Title = strings.TrimSpace(titleElement.Text())
			newsType = "titleAndTime"
		}

		// 检查是否为 sanTu 布局
		sanTuElement := e.DOM.Find(".sanTu h5")
		if sanTuElement.Length() > 0 {
			news.Title = strings.TrimSpace(sanTuElement.Text())
			newsType = "sanTu"
		}

		// 提取时间 - 从 time 标签
		timeElement := e.DOM.Find("time")
		if timeElement.Length() > 0 {
			news.Time = strings.TrimSpace(timeElement.Text())
		}

		// 提取评论数 - 从 commentNum 类的 span
		commentElement := e.DOM.Find(".commentNum")
		if commentElement.Length() > 0 {
			if commentText := strings.TrimSpace(commentElement.Text()); commentText != "" {
				if commentNum, err := strconv.Atoi(commentText); err == nil {
					news.CommentNum = commentNum
				}
			}
		}

		// 提取链接 - 根据布局类型选择合适的选择器
		if newsType == "sanTu" {
			// sanTu 布局：优先使用 .sanTu 内的链接
			sanTuLinkElement := e.DOM.Find(".sanTu a")
			if sanTuLinkElement.Length() > 0 {
				news.URL = sanTuLinkElement.AttrOr("href", "")
			}
		} else {
			// titleAndTime 布局或其他：使用通用的 a 标签
			linkElement := e.DOM.Find("a")
			if linkElement.Length() > 0 {
				news.URL = linkElement.AttrOr("href", "")
			}
		}

		// 提取图片链接 - 从 img 标签的 src
		imgElement := e.DOM.Find("img")
		if imgElement.Length() > 0 {
			news.ImageURL = imgElement.AttrOr("src", "")
		}

		// 只处理有标题的新闻
		if news.Title != "" {
			// 保存新闻到数据库（去重）
			if err := gnc.saveNewsToDB(news); err != nil {
				log.Printf("保存新闻失败 (SID: %s): %v", news.SID, err)
			} else {
				count++
				log.Printf("爬取新闻 [%s]: %s - %s", newsType, news.SID, news.Title)
			}
		}
	})

	// 查找"点击加载更多"按钮，用于获取后续页面
	var nextPageFound bool
	c.OnHTML("a.clickLoadMoreBtn", func(e *colly.HTMLElement) {
		nextPageFound = true
		dataNum := e.Attr("data-num")
		log.Printf("找到加载更多按钮，当前页数：%s", dataNum)
	})

	// 错误处理
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("请求失败: %s, 错误: %v", r.Request.URL, err)
	})

	// 请求完成处理
	c.OnResponse(func(r *colly.Response) {
		log.Printf("收到响应: %s, 状态码: %d, 大小: %d bytes",
			r.Request.URL, r.StatusCode, len(r.Body))
	})

	// 开始爬取
	err := c.Visit("https://wap.gamersky.com/")
	if err != nil {
		return 0, fmt.Errorf("访问页面失败: %v", err)
	}

	// 等待所有请求完成
	c.Wait()

	log.Printf("页面是否有加载更多按钮: %t", nextPageFound)
	return count, nil
}

// saveNewsToDB 保存新闻到数据库
func (gnc *NewsCrawler) saveNewsToDB(news *NewsInfo) error {
	// 使用 INSERT OR IGNORE 来实现去重
	sql := `
	INSERT OR IGNORE INTO gamersky_news 
	(sid, title, time, comment_num, url, image_url, topline_time, create_time)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := gnc.db.Exec(sql,
		news.SID, news.Title, news.Time, news.CommentNum,
		news.URL, news.ImageURL, news.TopLineTime, news.CreateTime)

	return err
}

// QueryNews 查询数据库中的新闻
func (gnc *NewsCrawler) QueryNews(limit int) ([]NewsInfo, error) {
	var query string
	var args []interface{}

	if limit > 0 {
		query = `
		SELECT sid, title, time, comment_num, url, image_url, topline_time, create_time 
		FROM gamersky_news 
		ORDER BY create_time DESC 
		LIMIT ?`
		args = append(args, limit)
	} else {
		query = `
		SELECT sid, title, time, comment_num, url, image_url, topline_time, create_time 
		FROM gamersky_news 
		ORDER BY create_time DESC`
	}

	rows, err := gnc.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var news []NewsInfo
	for rows.Next() {
		var item NewsInfo
		err := rows.Scan(
			&item.SID, &item.Title, &item.Time, &item.CommentNum,
			&item.URL, &item.ImageURL, &item.TopLineTime, &item.CreateTime,
		)
		if err != nil {
			return nil, err
		}
		news = append(news, item)
	}

	return news, nil
}

// Close 关闭数据库连接
func (gnc *NewsCrawler) Close() error {
	if gnc.db != nil {
		return gnc.db.Close()
	}
	return nil
}
