package crawler

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	_ "github.com/mattn/go-sqlite3"
)

// CommentInfo 评论信息结构体
type CommentInfo struct {
	SerialNumber int    `json:"serial_number"` // 序号
	ParentID     int64  `json:"parent_id"`     // 上级评论ID
	CommentID    int64  `json:"comment_id"`    // 评论ID
	UserID       int64  `json:"user_id"`       // 用户ID
	Username     string `json:"username"`      // 用户名
	UserLevel    int    `json:"user_level"`    // 用户等级
	Gender       string `json:"gender"`        // 性别
	Content      string `json:"content"`       // 评论内容
	CommentTime  string `json:"comment_time"`  // 评论时间
	ReplyCount   int    `json:"reply_count"`   // 回复数
	LikeCount    int    `json:"like_count"`    // 点赞数
	Signature    string `json:"signature"`     // 个性签名
	IPLocation   string `json:"ip_location"`   // IP属地
	IsVIP        string `json:"is_vip"`        // 是否是大会员
	Avatar       string `json:"avatar"`        // 头像
	BV           string `json:"bv"`            // 视频BV号
	VideoTitle   string `json:"video_title"`   // 视频标题
}

// VideoInfo 视频信息结构体
type VideoInfo struct {
	Keyword     string `json:"keyword"`      // 搜索关键词
	BVID        string `json:"bvid"`         // 视频BV号
	Title       string `json:"title"`        // 视频标题
	Author      string `json:"author"`       // 作者
	Play        int64  `json:"play"`         // 播放量
	VideoReview int    `json:"video_review"` // 评论数
	Favorites   int    `json:"favorites"`    // 收藏数
	PubDate     int64  `json:"pubdate"`      // 发布时间戳
	Duration    string `json:"duration"`     // 视频时长
	Like        int    `json:"like"`         // 点赞数
	Danmaku     int    `json:"danmaku"`      // 弹幕数
	Description string `json:"description"`  // 视频描述
	Pic         string `json:"pic"`          // 视频封面
	CreateTime  string `json:"create_time"`  // 记录创建时间
}

// BilibiliCommentCrawler B站评论爬虫结构体
type BilibiliCommentCrawler struct {
	db       *sql.DB
	cookie   string
	comments []CommentInfo
	config   *Config
}

// BilibiliVideoSearcher B站视频搜索结构体
type BilibiliVideoSearcher struct {
	db     *sql.DB
	cookie string
	config *Config
}

// Config 爬虫配置
type Config struct {
	BV           string        // BV号
	Mode         int           // 爬取模式 (2=最新, 3=热门)
	WithReplies  bool          // 是否爬取二级评论
	MaxPages     int           // 最大页数限制
	OutputPath   string        // 输出数据库路径
	CookiePath   string        // Cookie文件路径
	RequestDelay time.Duration // 请求间隔
}

// CommentResponse API响应结构体
type CommentResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Replies []struct {
			Parent int64 `json:"parent"`
			Rpid   int64 `json:"rpid"`
			Mid    int64 `json:"mid"`
			Member struct {
				Uname     string `json:"uname"`
				Sex       string `json:"sex"`
				Avatar    string `json:"avatar"`
				Sign      string `json:"sign"`
				LevelInfo struct {
					CurrentLevel int `json:"current_level"`
				} `json:"level_info"`
				Vip struct {
					VipStatus int `json:"vipStatus"`
				} `json:"vip"`
			} `json:"member"`
			Content struct {
				Message string `json:"message"`
			} `json:"content"`
			Ctime        int64 `json:"ctime"`
			Like         int   `json:"like"`
			ReplyControl struct {
				SubReplyEntryText string `json:"sub_reply_entry_text"`
				Location          string `json:"location"`
			} `json:"reply_control"`
		} `json:"replies"`
		Cursor struct {
			PaginationReply struct {
				NextOffset string `json:"next_offset"`
			} `json:"pagination_reply"`
		} `json:"cursor"`
	} `json:"data"`
}

// SecondCommentResponse 二级评论响应结构体
type SecondCommentResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Replies []struct {
			Parent int64 `json:"parent"`
			Rpid   int64 `json:"rpid"`
			Mid    int64 `json:"mid"`
			Member struct {
				Uname     string `json:"uname"`
				Sex       string `json:"sex"`
				Avatar    string `json:"avatar"`
				Sign      string `json:"sign"`
				LevelInfo struct {
					CurrentLevel int `json:"current_level"`
				} `json:"level_info"`
				Vip struct {
					VipStatus int `json:"vipStatus"`
				} `json:"vip"`
			} `json:"member"`
			Content struct {
				Message string `json:"message"`
			} `json:"content"`
			Ctime        int64 `json:"ctime"`
			Like         int   `json:"like"`
			ReplyControl struct {
				SubReplyEntryText string `json:"sub_reply_entry_text"`
				Location          string `json:"location"`
			} `json:"reply_control"`
		} `json:"replies"`
	} `json:"data"`
}

// SearchResponse 搜索API响应结构体
type SearchResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Result []struct {
			ResultType string `json:"result_type"`
			Data       []struct {
				Type        string `json:"type"`
				ID          int64  `json:"id"`
				Author      string `json:"author"`
				MID         int64  `json:"mid"`
				TypeID      string `json:"typeid"`
				TypeName    string `json:"typename"`
				ArcURL      string `json:"arcurl"`
				AID         int64  `json:"aid"`
				BVID        string `json:"bvid"`
				Title       string `json:"title"`
				Description string `json:"description"`
				Pic         string `json:"pic"`
				Play        int64  `json:"play"`
				VideoReview int    `json:"video_review"`
				Favorites   int    `json:"favorites"`
				Tag         string `json:"tag"`
				Review      int    `json:"review"`
				PubDate     int64  `json:"pubdate"`
				SendDate    int64  `json:"senddate"`
				Duration    string `json:"duration"`
				Like        int    `json:"like"`
				Danmaku     int    `json:"danmaku"`
			} `json:"data"`
		} `json:"result"`
	} `json:"data"`
}

// NewBilibiliCommentCrawler 创建新的B站评论爬虫实例
func NewBilibiliCommentCrawler(config *Config) (*BilibiliCommentCrawler, error) {
	// 初始化数据库
	db, err := getDBConnection(config.OutputPath)
	if err != nil {
		return nil, fmt.Errorf("初始化数据库失败: %v", err)
	}

	// 读取cookie
	cookie, err := readCookie(config.CookiePath)
	if err != nil {
		return nil, fmt.Errorf("读取cookie失败: %v", err)
	}

	return &BilibiliCommentCrawler{
		db:       db,
		cookie:   cookie,
		comments: make([]CommentInfo, 0),
		config:   config,
	}, nil
}

// getDBConnection 获取数据库连接
func getDBConnection(dbPath string) (*sql.DB, error) {
	if dbPath == "" {
		dbPath = "./data/crawler.db"
	}

	// 创建目录
	dir := strings.TrimSuffix(dbPath, "/crawler.db")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// 连接数据库
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// 创建评论表
	createCommentTableSQL := `
	CREATE TABLE IF NOT EXISTS bilibili_comments (
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
	)`

	_, err = db.Exec(createCommentTableSQL)
	if err != nil {
		return nil, err
	}

	// 创建视频搜索结果表
	createVideoTableSQL := `
	CREATE TABLE IF NOT EXISTS bilibili_videos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		keyword TEXT NOT NULL,
		bvid TEXT NOT NULL,
		title TEXT,
		author TEXT,
		play INTEGER,
		video_review INTEGER,
		favorites INTEGER,
		pubdate INTEGER,
		duration TEXT,
		like_count INTEGER,
		danmaku INTEGER,
		description TEXT,
		pic TEXT,
		create_time TEXT,
		UNIQUE(keyword, bvid)
	)`

	_, err = db.Exec(createVideoTableSQL)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// readCookie 读取cookie文件
func readCookie(cookiePath string) (string, error) {
	// 如果没有指定路径，使用默认路径
	if cookiePath == "" {
		// 首先尝试当前目录
		cookiePath = "bili_cookie.txt"
		if _, err := os.Stat(cookiePath); os.IsNotExist(err) {
			// 如果当前目录不存在，尝试py-crawler目录
			cookiePath = "py-crawler/bili_cookie.txt"
		}
	}

	cookieBytes, err := os.ReadFile(cookiePath)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(cookieBytes)), nil
}

// insertCommentToDB 插入评论到数据库
func (bcc *BilibiliCommentCrawler) insertCommentToDB(comment CommentInfo) error {
	sql := `
	INSERT OR IGNORE INTO bilibili_comments 
	(序号, 上级评论ID, 评论ID, 用户ID, 用户名, 用户等级, 性别, 评论内容, 评论时间, 回复数, 点赞数, 个性签名, IP属地, 是否是大会员, 头像, 视频BV号, 视频标题)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := bcc.db.Exec(sql,
		comment.SerialNumber, comment.ParentID, comment.CommentID, comment.UserID,
		comment.Username, comment.UserLevel, comment.Gender, comment.Content,
		comment.CommentTime, comment.ReplyCount, comment.LikeCount, comment.Signature,
		comment.IPLocation, comment.IsVIP, comment.Avatar, comment.BV, comment.VideoTitle)

	return err
}

// getHeader 获取HTTP请求头
func (bcc *BilibiliCommentCrawler) getHeader() map[string]string {
	return map[string]string{
		"Cookie":     bcc.cookie,
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:135.0) Gecko/20100101 Firefox/135.0",
	}
}

// GetVideoInfo 通过BV号获取视频的OID和标题
func (bcc *BilibiliCommentCrawler) GetVideoInfo(bv string) (string, string, error) {
	url := fmt.Sprintf("https://www.bilibili.com/video/%s/?p=14&spm_id_from=pageDriver&vd_source=cd6ee6b033cd2da64359bad72619ca8a", bv)

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", err
	}

	// 设置请求头
	for key, value := range bcc.getHeader() {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	content := string(body)

	// 提取视频OID
	oidRegex := regexp.MustCompile(fmt.Sprintf(`"aid":(\d+),"bvid":"%s"`, bv))
	oidMatches := oidRegex.FindStringSubmatch(content)
	if len(oidMatches) < 2 {
		return "", "", fmt.Errorf("无法提取视频OID")
	}
	oid := oidMatches[1]

	// 提取视频标题
	titleRegex := regexp.MustCompile(`<title>(.*?)</title>`)
	titleMatches := titleRegex.FindStringSubmatch(content)
	title := "未识别"
	if len(titleMatches) >= 2 {
		title = titleMatches[1]
	}

	return oid, title, nil
}

// md5Hash MD5加密
func md5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// cleanHTMLTags 清理HTML标签
func cleanHTMLTags(text string) string {
	// 移除HTML标签
	re := regexp.MustCompile(`<[^>]*>`)
	cleaned := re.ReplaceAllString(text, "")
	return cleaned
}

// CrawlComments 爬取评论的主要函数
func (bcc *BilibiliCommentCrawler) CrawlComments(bv, oid, pageID string, count int, title string, isSecond bool) (string, int, error) {
	// 参数
	mode := bcc.config.Mode // 使用配置中的模式
	plat := 1
	commentType := 1
	webLocation := 1315875

	// 获取当前时间戳
	wts := time.Now().Unix()

	var paginationStr string
	if pageID != "" {
		paginationStr = fmt.Sprintf(`{"offset":"%s"}`, pageID)
	} else {
		paginationStr = `{"offset":""}`
	}

	// 构建签名
	encodedPagination := url.QueryEscape(paginationStr)
	code := fmt.Sprintf("mode=%d&oid=%s&pagination_str=%s&plat=%d&type=%d&web_location=%d&wts=%d",
		mode, oid, encodedPagination, plat, commentType, webLocation, wts) + "ea1db124af3c7062474693fa704f4ff8"
	wRid := md5Hash(code)

	// 构建URL
	requestURL := fmt.Sprintf("https://api.bilibili.com/x/v2/reply/wbi/main?oid=%s&type=%d&mode=%d&pagination_str=%s&plat=1&web_location=1315875&w_rid=%s&wts=%d",
		oid, commentType, mode, url.QueryEscape(paginationStr), wRid, wts)

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return "", count, err
	}

	// 设置请求头
	for key, value := range bcc.getHeader() {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", count, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", count, err
	}

	// 解析JSON响应
	var commentResp CommentResponse
	if err := json.Unmarshal(body, &commentResp); err != nil {
		return "", count, err
	}

	// 处理评论
	for _, reply := range commentResp.Data.Replies {
		count++

		if count%1000 == 0 {
			time.Sleep(bcc.config.RequestDelay * 10) // 大批量时增加延迟
		}

		// 构建评论信息
		comment := CommentInfo{
			SerialNumber: count,
			ParentID:     reply.Parent,
			CommentID:    reply.Rpid,
			UserID:       reply.Mid,
			Username:     reply.Member.Uname,
			UserLevel:    reply.Member.LevelInfo.CurrentLevel,
			Gender:       reply.Member.Sex,
			Content:      reply.Content.Message,
			CommentTime:  time.Unix(reply.Ctime, 0).Format("2006-01-02 15:04:05"),
			LikeCount:    reply.Like,
			Signature:    reply.Member.Sign,
			Avatar:       reply.Member.Avatar,
			BV:           bv,
			VideoTitle:   title,
		}

		// 处理VIP状态
		if reply.Member.Vip.VipStatus == 0 {
			comment.IsVIP = "否"
		} else {
			comment.IsVIP = "是"
		}

		// 处理IP属地
		if len(reply.ReplyControl.Location) > 5 {
			comment.IPLocation = reply.ReplyControl.Location[5:]
		} else {
			comment.IPLocation = "未知"
		}

		// 处理回复数
		if reply.ReplyControl.SubReplyEntryText != "" {
			re := regexp.MustCompile(`\d+`)
			matches := re.FindStringSubmatch(reply.ReplyControl.SubReplyEntryText)
			if len(matches) > 0 {
				if replyCount, err := strconv.Atoi(matches[0]); err == nil {
					comment.ReplyCount = replyCount
				}
			}
		}

		// 插入数据库
		if err := bcc.insertCommentToDB(comment); err != nil {
			log.Printf("插入评论失败: %v", err)
		}

		// 处理二级评论
		if isSecond && comment.ReplyCount > 0 {
			if err := bcc.crawlSecondComments(oid, reply.Rpid, comment.ReplyCount, &count, bv, title); err != nil {
				log.Printf("爬取二级评论失败: %v", err)
			}
		}
	}

	// 获取下一页的pageID
	nextPageID := commentResp.Data.Cursor.PaginationReply.NextOffset

	return nextPageID, count, nil
}

// crawlSecondComments 爬取二级评论
func (bcc *BilibiliCommentCrawler) crawlSecondComments(oid string, rootID int64, replyCount int, count *int, bv, title string) error {
	pages := replyCount/10 + 1
	maxPages := bcc.config.MaxPages
	if maxPages > 0 && pages > maxPages { // 使用配置中的最大页数限制
		pages = maxPages
	} else if maxPages == 0 && pages > 10 { // 默认限制
		pages = 10
	}

	for page := 1; page <= pages; page++ {
		secondURL := fmt.Sprintf("https://api.bilibili.com/x/v2/reply/reply?oid=%s&type=1&root=%d&ps=10&pn=%d&web_location=333.788",
			oid, rootID, page)

		client := &http.Client{Timeout: 30 * time.Second}
		req, err := http.NewRequest("GET", secondURL, nil)
		if err != nil {
			return err
		}

		// 设置请求头
		for key, value := range bcc.getHeader() {
			req.Header.Set(key, value)
		}

		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return err
		}

		// 解析JSON响应
		var secondResp SecondCommentResponse
		if err := json.Unmarshal(body, &secondResp); err != nil {
			return err
		}

		// 处理二级评论
		for _, second := range secondResp.Data.Replies {
			*count++

			// 构建二级评论信息
			comment := CommentInfo{
				SerialNumber: *count,
				ParentID:     second.Parent,
				CommentID:    second.Rpid,
				UserID:       second.Mid,
				Username:     second.Member.Uname,
				UserLevel:    second.Member.LevelInfo.CurrentLevel,
				Gender:       second.Member.Sex,
				Content:      second.Content.Message,
				CommentTime:  time.Unix(second.Ctime, 0).Format("2006-01-02 15:04:05"),
				LikeCount:    second.Like,
				Signature:    second.Member.Sign,
				Avatar:       second.Member.Avatar,
				BV:           bv,
				VideoTitle:   title,
			}

			// 处理VIP状态
			if second.Member.Vip.VipStatus == 0 {
				comment.IsVIP = "否"
			} else {
				comment.IsVIP = "是"
			}

			// 处理IP属地
			if len(second.ReplyControl.Location) > 5 {
				comment.IPLocation = second.ReplyControl.Location[5:]
			} else {
				comment.IPLocation = "未知"
			}

			// 处理回复数
			if second.ReplyControl.SubReplyEntryText != "" {
				re := regexp.MustCompile(`\d+`)
				matches := re.FindStringSubmatch(second.ReplyControl.SubReplyEntryText)
				if len(matches) > 0 {
					if replyCount, err := strconv.Atoi(matches[0]); err == nil {
						comment.ReplyCount = replyCount
					}
				}
			}

			// 插入数据库
			if err := bcc.insertCommentToDB(comment); err != nil {
				log.Printf("插入二级评论失败: %v", err)
			}
		}

		// 防止请求过快
		time.Sleep(bcc.config.RequestDelay)
	}

	return nil
}

// Close 关闭数据库连接
func (bcc *BilibiliCommentCrawler) Close() error {
	if bcc.db != nil {
		return bcc.db.Close()
	}
	return nil
}

// NewBilibiliVideoSearcher 创建新的B站视频搜索器实例
func NewBilibiliVideoSearcher(config *Config) (*BilibiliVideoSearcher, error) {
	// 初始化数据库
	db, err := getDBConnection(config.OutputPath)
	if err != nil {
		return nil, fmt.Errorf("初始化数据库失败: %v", err)
	}

	// 读取cookie
	cookie, err := readCookie(config.CookiePath)
	if err != nil {
		return nil, fmt.Errorf("读取cookie失败: %v", err)
	}

	return &BilibiliVideoSearcher{
		db:     db,
		cookie: cookie,
		config: config,
	}, nil
}

// getSearchHeader 获取搜索请求头
func (bvs *BilibiliVideoSearcher) getSearchHeader() map[string]string {
	return map[string]string{
		"Cookie":     bvs.cookie,
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:135.0) Gecko/20100101 Firefox/135.0",
		"Referer":    "https://search.bilibili.com/",
	}
}

// SearchVideos 搜索视频
func (bvs *BilibiliVideoSearcher) SearchVideos(keyword string, page int, pageSize int) ([]VideoInfo, error) {
	// 构建搜索URL
	searchURL := fmt.Sprintf("https://api.bilibili.com/x/web-interface/wbi/search/all/v2?keyword=%s&page=%d&page_size=%d&platform=pc",
		url.QueryEscape(keyword), page, pageSize)

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	for key, value := range bvs.getSearchHeader() {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 解析JSON响应
	var searchResp SearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, err
	}

	// 处理搜索结果
	var videos []VideoInfo
	for _, result := range searchResp.Data.Result {
		if result.ResultType == "video" {
			for _, videoData := range result.Data {
				if videoData.Type == "video" {
					video := VideoInfo{
						Keyword:     keyword,
						BVID:        videoData.BVID,
						Title:       cleanHTMLTags(videoData.Title),
						Author:      videoData.Author,
						Play:        videoData.Play,
						VideoReview: videoData.VideoReview,
						Favorites:   videoData.Favorites,
						PubDate:     videoData.PubDate,
						Duration:    videoData.Duration,
						Like:        videoData.Like,
						Danmaku:     videoData.Danmaku,
						Description: cleanHTMLTags(videoData.Description),
						Pic:         videoData.Pic,
						CreateTime:  time.Now().Format("2006-01-02 15:04:05"),
					}
					videos = append(videos, video)
				}
			}
		}
	}

	return videos, nil
}

// SaveVideoToDB 保存视频信息到数据库
func (bvs *BilibiliVideoSearcher) SaveVideoToDB(video VideoInfo) error {
	sql := `
	INSERT OR IGNORE INTO bilibili_videos 
	(keyword, bvid, title, author, play, video_review, favorites, pubdate, duration, like_count, danmaku, description, pic, create_time)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := bvs.db.Exec(sql,
		video.Keyword, video.BVID, video.Title, video.Author,
		video.Play, video.VideoReview, video.Favorites, video.PubDate,
		video.Duration, video.Like, video.Danmaku, video.Description,
		video.Pic, video.CreateTime)

	return err
}

// Close 关闭数据库连接
func (bvs *BilibiliVideoSearcher) Close() error {
	if bvs.db != nil {
		return bvs.db.Close()
	}
	return nil
}

// QueryVideos 查询数据库中的视频信息
func (bvs *BilibiliVideoSearcher) QueryVideos(keyword string, limit int) ([]VideoInfo, error) {
	var sql string
	var args []interface{}

	if keyword != "" {
		sql = `SELECT keyword, bvid, title, author, play, video_review, favorites, pubdate, duration, like_count, danmaku, description, pic, create_time 
			   FROM bilibili_videos WHERE keyword = ? ORDER BY play DESC LIMIT ?`
		args = []interface{}{keyword, limit}
	} else {
		sql = `SELECT keyword, bvid, title, author, play, video_review, favorites, pubdate, duration, like_count, danmaku, description, pic, create_time 
			   FROM bilibili_videos ORDER BY play DESC LIMIT ?`
		args = []interface{}{limit}
	}

	rows, err := bvs.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []VideoInfo
	for rows.Next() {
		var video VideoInfo
		err := rows.Scan(
			&video.Keyword, &video.BVID, &video.Title, &video.Author,
			&video.Play, &video.VideoReview, &video.Favorites, &video.PubDate,
			&video.Duration, &video.Like, &video.Danmaku, &video.Description,
			&video.Pic, &video.CreateTime,
		)
		if err != nil {
			return nil, err
		}
		videos = append(videos, video)
	}

	return videos, nil
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
type NewsInfo struct {
	SID         string `json:"sid"`          // 新闻ID (data-sid)
	Title       string `json:"title"`        // 新闻标题
	Time        string `json:"time"`         // 发布时间
	CommentNum  int    `json:"comment_num"`  // 评论数
	URL         string `json:"url"`          // 新闻链接
	ImageURL    string `json:"image_url"`    // 图片链接
	CreateTime  string `json:"create_time"`  // 记录创建时间
	TopLineTime string `json:"topline_time"` // 置顶时间
}

// GamerskyComment 游戏天空评论信息结构体
type GamerskyComment struct {
	ID                 int64  `json:"id"`                   // 评论ID
	ArticleID          string `json:"article_id"`           // 文章ID
	UserID             int    `json:"user_id"`              // 用户ID
	Username           string `json:"username"`             // 用户名
	Content            string `json:"content"`              // 评论内容
	CommentTime        string `json:"comment_time"`         // 评论时间
	SupportCount       int    `json:"support_count"`        // 点赞数
	ReplyCount         int    `json:"reply_count"`          // 回复数
	ParentID           int64  `json:"parent_id"`            // 父评论ID (0表示一级评论)
	UserAvatar         string `json:"user_avatar"`          // 用户头像
	UserLevel          int    `json:"user_level"`           // 用户等级
	IPLocation         string `json:"ip_location"`          // IP位置
	DeviceName         string `json:"device_name"`          // 设备名称
	FloorNumber        int    `json:"floor_number"`         // 楼层号
	IsTuijian          bool   `json:"is_tuijian"`           // 是否推荐
	IsAuthor           bool   `json:"is_author"`            // 是否作者
	IsBest             bool   `json:"is_best"`              // 是否最佳
	UserAuthentication string `json:"user_authentication"`  // 用户认证
	UserGroupID        int    `json:"user_group_id"`        // 用户组ID
	ThirdPlatformBound string `json:"third_platform_bound"` // 第三方平台绑定
	CreateTime         string `json:"create_time"`          // 记录创建时间
}

// CommentAPIRequest 评论API请求结构体
type CommentAPIRequest struct {
	ArticleID       string `json:"articleId"`
	MinPraisesCount int    `json:"minPraisesCount"`
	RepliesMaxCount int    `json:"repliesMaxCount"`
	PageIndex       int    `json:"pageIndex"`
	PageSize        int    `json:"pageSize"`
	Order           string `json:"order"`
}

// CommentAPIResponse 评论API响应结构体
type CommentAPIResponse struct {
	ErrorCode    int    `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
	Result       struct {
		CommentsCount int `json:"commentsCount"`
		IsUpdateImage int `json:"isUpdateImage"`
		Comments      []struct {
			CommentID          int64         `json:"comment_id"`
			CreateTime         int64         `json:"create_time"`
			LastJoinTime       int64         `json:"last_join_time"`
			IsTuijian          bool          `json:"is_tuijian"`
			IsAuthor           bool          `json:"is_author"`
			IsBest             bool          `json:"is_best"`
			BeAuthorPraise     bool          `json:"beAuthorPraise"`
			From               int           `json:"from"`
			Content            string        `json:"content"`
			SupportCount       int           `json:"support_count"`
			IPLocation         string        `json:"ip_location"`
			UserID             int           `json:"user_id"`
			Nickname           string        `json:"nickname"`
			ImgURL             string        `json:"img_url"`
			DeviceName         string        `json:"deviceName"`
			UserLevel          int           `json:"userLevel"`
			UserAuthentication string        `json:"userAuthentication"`
			UserGroupID        int           `json:"userGroupId"`
			FloorNumber        int           `json:"floorNumber"`
			ImageInfes         []interface{} `json:"imageInfes"`
			ThirdPlatformBound string        `json:"thirdPlatformBound"`
			Comments           interface{}   `json:"comments"`
			Replies            []struct {
				CommentID          int64         `json:"comment_id"`
				CreateTime         int64         `json:"create_time"`
				LastJoinTime       int64         `json:"last_join_time"`
				IsTuijian          bool          `json:"is_tuijian"`
				IsAuthor           bool          `json:"is_author"`
				IsBest             bool          `json:"is_best"`
				BeAuthorPraise     bool          `json:"beAuthorPraise"`
				From               int           `json:"from"`
				Content            string        `json:"content"`
				SupportCount       int           `json:"support_count"`
				IPLocation         string        `json:"ip_location"`
				UserID             int           `json:"user_id"`
				Nickname           string        `json:"nickname"`
				ImgURL             string        `json:"img_url"`
				DeviceName         string        `json:"deviceName"`
				UserLevel          int           `json:"userLevel"`
				UserAuthentication string        `json:"userAuthentication"`
				UserGroupID        int           `json:"userGroupId"`
				FloorNumber        int           `json:"floorNumber"`
				ImageInfes         []interface{} `json:"imageInfes"`
				ThirdPlatformBound string        `json:"thirdPlatformBound"`
			} `json:"replies"`
			RepliesCount int `json:"repliesCount"`
		} `json:"comments"`
	} `json:"result"`
}

// GamerskyNewsCrawler Gamersky新闻爬虫结构体
type GamerskyNewsCrawler struct {
	db     *sql.DB
	config *Config
}

// GamerskyCommentCrawler Gamersky评论爬虫结构体
type GamerskyCommentCrawler struct {
	db     *sql.DB
	config *Config
}

// NewGamerskyNewsCrawler 创建新的Gamersky新闻爬虫实例
func NewGamerskyNewsCrawler(config *Config) (*GamerskyNewsCrawler, error) {
	// 初始化数据库
	db, err := getGamerskyDBConnection(config.OutputPath)
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %v", err)
	}

	return &GamerskyNewsCrawler{
		db:     db,
		config: config,
	}, nil
}

// getGamerskyDBConnection 获取Gamersky数据库连接
func getGamerskyDBConnection(dbPath string) (*sql.DB, error) {
	if dbPath == "" {
		dbPath = "./data/gamersky.db"
	}

	// 创建目录
	dir := strings.TrimSuffix(dbPath, "/gamersky.db")
	if dir != dbPath {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return nil, fmt.Errorf("创建目录失败: %v", err)
		}
	}

	// 连接数据库
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %v", err)
	}

	// 创建新闻表
	createNewsTableSQL := `
	CREATE TABLE IF NOT EXISTS gamersky_news (
		sid TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		time TEXT,
		comment_num INTEGER DEFAULT 0,
		url TEXT,
		image_url TEXT,
		topline_time TEXT,
		create_time TEXT DEFAULT CURRENT_TIMESTAMP
	)`

	_, err = db.Exec(createNewsTableSQL)
	if err != nil {
		return nil, fmt.Errorf("创建表失败: %v", err)
	}

	// 创建评论表
	createCommentTableSQL := `
	CREATE TABLE IF NOT EXISTS gamersky_comments (
		id INTEGER PRIMARY KEY,
		article_id TEXT NOT NULL,
		user_id INTEGER,
		username TEXT,
		content TEXT,
		comment_time TEXT,
		support_count INTEGER DEFAULT 0,
		reply_count INTEGER DEFAULT 0,
		parent_id INTEGER DEFAULT 0,
		user_avatar TEXT,
		user_level INTEGER DEFAULT 0,
		ip_location TEXT,
		device_name TEXT,
		floor_number INTEGER DEFAULT 0,
		is_tuijian BOOLEAN DEFAULT FALSE,
		is_author BOOLEAN DEFAULT FALSE,
		is_best BOOLEAN DEFAULT FALSE,
		user_authentication TEXT,
		user_group_id INTEGER DEFAULT 0,
		third_platform_bound TEXT,
		create_time TEXT DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(id, article_id)
	)`

	_, err = db.Exec(createCommentTableSQL)
	if err != nil {
		return nil, fmt.Errorf("创建评论表失败: %v", err)
	}

	return db, nil
}

// CrawlNews 爬取Gamersky新闻
func (gnc *GamerskyNewsCrawler) CrawlNews(page int) (int, error) {
	if page == 1 {
		// 第一页使用Colly爬取（保持原有逻辑）
		return gnc.crawlFirstPage()
	} else {
		// 后续页面使用API
		return gnc.crawlAPIPage(page)
	}
}

// crawlAPIPage 使用API爬取指定页面
func (gnc *GamerskyNewsCrawler) crawlAPIPage(page int) (int, error) {
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
func (gnc *GamerskyNewsCrawler) crawlFirstPage() (int, error) {
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

		// 提取标题 - 从 titleAndTime 内的 h5 标签
		titleElement := e.DOM.Find(".titleAndTime h5")
		if titleElement.Length() > 0 {
			news.Title = strings.TrimSpace(titleElement.Text())
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

		// 提取链接 - 从 a 标签的 href
		linkElement := e.DOM.Find("a")
		if linkElement.Length() > 0 {
			news.URL = linkElement.AttrOr("href", "")
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
				log.Printf("爬取新闻: %s - %s", news.SID, news.Title)
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
func (gnc *GamerskyNewsCrawler) saveNewsToDB(news *NewsInfo) error {
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

// Close 关闭数据库连接
func (gnc *GamerskyNewsCrawler) Close() error {
	if gnc.db != nil {
		return gnc.db.Close()
	}
	return nil
}

// QueryNews 查询数据库中的新闻
func (gnc *GamerskyNewsCrawler) QueryNews(limit int) ([]NewsInfo, error) {
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

// NewGamerskyCommentCrawler 创建新的Gamersky评论爬虫实例
func NewGamerskyCommentCrawler(config *Config) (*GamerskyCommentCrawler, error) {
	// 初始化数据库
	db, err := getGamerskyDBConnection(config.OutputPath)
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %v", err)
	}

	return &GamerskyCommentCrawler{
		db:     db,
		config: config,
	}, nil
}

// CrawlComments 爬取指定文章的评论
func (gcc *GamerskyCommentCrawler) CrawlComments(articleID string, maxPages int) (int, error) {
	totalCount := 0

	for page := 1; page <= maxPages; page++ {
		log.Printf("正在爬取文章 %s 第 %d 页评论...", articleID, page)

		count, hasMore, err := gcc.crawlCommentsPage(articleID, page)
		if err != nil {
			log.Printf("爬取第 %d 页评论失败: %v", page, err)
			continue
		}

		totalCount += count
		log.Printf("第 %d 页爬取完成，新增 %d 条评论", page, count)

		// 如果没有更多评论了，停止爬取
		if !hasMore || count == 0 {
			log.Printf("第 %d 页没有更多评论，停止爬取", page)
			break
		}

		// 延迟
		if page < maxPages {
			time.Sleep(gcc.config.RequestDelay)
		}
	}

	return totalCount, nil
}

// crawlCommentsPage 爬取指定页面的评论
func (gcc *GamerskyCommentCrawler) crawlCommentsPage(articleID string, pageIndex int) (int, bool, error) {
	// 构造API请求
	requestData := CommentAPIRequest{
		ArticleID:       articleID,
		MinPraisesCount: 0,
		RepliesMaxCount: 10,
		PageIndex:       pageIndex,
		PageSize:        20,
		Order:           "tuiJian",
	}

	// 序列化请求数据为JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return 0, false, fmt.Errorf("序列化请求数据失败: %v", err)
	}

	// URL编码JSON数据
	encodedRequest := url.QueryEscape(string(jsonData))

	// 构造完整的API URL
	apiURL := fmt.Sprintf("https://cm.gamersky.com/appapi/GetArticleCommentWithClubStyle?request=%s", encodedRequest)

	// 创建HTTP客户端
	client := &http.Client{Timeout: 30 * time.Second}

	// 创建请求
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return 0, false, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://www.gamersky.com/")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return 0, false, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, false, fmt.Errorf("读取响应失败: %v", err)
	}

	log.Printf("评论API响应状态: %d, 大小: %d bytes", resp.StatusCode, len(body))

	// 解析JSON响应
	var apiResponse CommentAPIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return 0, false, fmt.Errorf("解析JSON失败: %v, 响应内容: %s", err, string(body[:min(200, len(body))]))
	}

	// 检查API错误
	if apiResponse.ErrorCode != 0 {
		return 0, false, fmt.Errorf("API错误: %s (代码: %d)", apiResponse.ErrorMessage, apiResponse.ErrorCode)
	}

	// 处理评论数据
	count := 0
	for _, comment := range apiResponse.Result.Comments {
		// 保存一级评论
		gamerskyComment := &GamerskyComment{
			ID:                 comment.CommentID,
			ArticleID:          articleID,
			UserID:             comment.UserID,
			Username:           comment.Nickname,
			Content:            comment.Content,
			CommentTime:        time.Unix(comment.CreateTime/1000, 0).Format("2006-01-02 15:04:05"),
			SupportCount:       comment.SupportCount,
			ReplyCount:         comment.RepliesCount,
			ParentID:           0, // 一级评论父ID为0
			UserAvatar:         comment.ImgURL,
			UserLevel:          comment.UserLevel,
			IPLocation:         comment.IPLocation,
			DeviceName:         comment.DeviceName,
			FloorNumber:        comment.FloorNumber,
			IsTuijian:          comment.IsTuijian,
			IsAuthor:           comment.IsAuthor,
			IsBest:             comment.IsBest,
			UserAuthentication: comment.UserAuthentication,
			UserGroupID:        comment.UserGroupID,
			ThirdPlatformBound: comment.ThirdPlatformBound,
			CreateTime:         time.Now().Format("2006-01-02 15:04:05"),
		}

		if err := gcc.saveCommentToDB(gamerskyComment); err != nil {
			log.Printf("保存评论失败 (ID: %d): %v", comment.CommentID, err)
		} else {
			count++
			log.Printf("保存评论: %d - %s", comment.CommentID, comment.Nickname)
		}

		// 处理回复（二级评论）
		for _, reply := range comment.Replies {
			// 跳过无效的回复数据
			if reply.CommentID == 0 || reply.Nickname == "" {
				log.Printf("跳过无效回复数据: ID=%d, Nickname=%s", reply.CommentID, reply.Nickname)
				continue
			}

			replyComment := &GamerskyComment{
				ID:                 reply.CommentID,
				ArticleID:          articleID,
				UserID:             reply.UserID,
				Username:           reply.Nickname,
				Content:            reply.Content,
				CommentTime:        time.Unix(reply.CreateTime/1000, 0).Format("2006-01-02 15:04:05"),
				SupportCount:       reply.SupportCount,
				ReplyCount:         0,                 // 二级评论通常没有回复数
				ParentID:           comment.CommentID, // 父评论ID
				UserAvatar:         reply.ImgURL,
				UserLevel:          reply.UserLevel,
				IPLocation:         reply.IPLocation,
				DeviceName:         reply.DeviceName,
				FloorNumber:        reply.FloorNumber,
				IsTuijian:          reply.IsTuijian,
				IsAuthor:           reply.IsAuthor,
				IsBest:             reply.IsBest,
				UserAuthentication: reply.UserAuthentication,
				UserGroupID:        reply.UserGroupID,
				ThirdPlatformBound: reply.ThirdPlatformBound,
				CreateTime:         time.Now().Format("2006-01-02 15:04:05"),
			}

			if err := gcc.saveCommentToDB(replyComment); err != nil {
				log.Printf("保存回复失败 (ID: %d): %v", reply.CommentID, err)
			} else {
				count++
				log.Printf("保存回复: %d - %s", reply.CommentID, reply.Nickname)
			}
		}
	}

	// 判断是否还有更多页面（简单判断：如果返回的评论数量等于请求的pageSize，则可能还有更多）
	hasMore := len(apiResponse.Result.Comments) >= 20 // pageSize为20

	log.Printf("文章 %s 第 %d 页完成，共 %d 条评论",
		articleID, pageIndex, len(apiResponse.Result.Comments))

	return count, hasMore, nil
}

// saveCommentToDB 保存评论到数据库
func (gcc *GamerskyCommentCrawler) saveCommentToDB(comment *GamerskyComment) error {
	// 使用 INSERT OR IGNORE 来实现去重
	sql := `
	INSERT OR IGNORE INTO gamersky_comments 
	(id, article_id, user_id, username, content, comment_time, support_count, reply_count, parent_id, user_avatar, user_level, ip_location, device_name, floor_number, is_tuijian, is_author, is_best, user_authentication, user_group_id, third_platform_bound, create_time)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := gcc.db.Exec(sql,
		comment.ID, comment.ArticleID, comment.UserID, comment.Username,
		comment.Content, comment.CommentTime, comment.SupportCount, comment.ReplyCount,
		comment.ParentID, comment.UserAvatar, comment.UserLevel, comment.IPLocation,
		comment.DeviceName, comment.FloorNumber, comment.IsTuijian, comment.IsAuthor,
		comment.IsBest, comment.UserAuthentication, comment.UserGroupID, comment.ThirdPlatformBound,
		comment.CreateTime)

	return err
}

// QueryComments 查询数据库中的评论
func (gcc *GamerskyCommentCrawler) QueryComments(articleID string, limit int) ([]GamerskyComment, error) {
	var query string
	var args []interface{}

	if articleID != "" {
		if limit > 0 {
			query = `
			SELECT id, article_id, user_id, username, content, comment_time, support_count, reply_count, parent_id, user_avatar, user_level, ip_location, device_name, floor_number, is_tuijian, is_author, is_best, user_authentication, user_group_id, third_platform_bound, create_time 
			FROM gamersky_comments 
			WHERE article_id = ? 
			ORDER BY id ASC 
			LIMIT ?`
			args = append(args, articleID, limit)
		} else {
			query = `
			SELECT id, article_id, user_id, username, content, comment_time, support_count, reply_count, parent_id, user_avatar, user_level, ip_location, device_name, floor_number, is_tuijian, is_author, is_best, user_authentication, user_group_id, third_platform_bound, create_time 
			FROM gamersky_comments 
			WHERE article_id = ? 
			ORDER BY id ASC`
			args = append(args, articleID)
		}
	} else {
		if limit > 0 {
			query = `
			SELECT id, article_id, user_id, username, content, comment_time, support_count, reply_count, parent_id, user_avatar, user_level, ip_location, device_name, floor_number, is_tuijian, is_author, is_best, user_authentication, user_group_id, third_platform_bound, create_time 
			FROM gamersky_comments 
			ORDER BY create_time DESC 
			LIMIT ?`
			args = append(args, limit)
		} else {
			query = `
			SELECT id, article_id, user_id, username, content, comment_time, support_count, reply_count, parent_id, user_avatar, user_level, ip_location, device_name, floor_number, is_tuijian, is_author, is_best, user_authentication, user_group_id, third_platform_bound, create_time 
			FROM gamersky_comments 
			ORDER BY create_time DESC`
		}
	}

	rows, err := gcc.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []GamerskyComment
	for rows.Next() {
		var comment GamerskyComment
		err := rows.Scan(
			&comment.ID, &comment.ArticleID, &comment.UserID, &comment.Username,
			&comment.Content, &comment.CommentTime, &comment.SupportCount, &comment.ReplyCount,
			&comment.ParentID, &comment.UserAvatar, &comment.UserLevel, &comment.IPLocation,
			&comment.DeviceName, &comment.FloorNumber, &comment.IsTuijian, &comment.IsAuthor,
			&comment.IsBest, &comment.UserAuthentication, &comment.UserGroupID, &comment.ThirdPlatformBound,
			&comment.CreateTime,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

// Close 关闭数据库连接
func (gcc *GamerskyCommentCrawler) Close() error {
	if gcc.db != nil {
		return gcc.db.Close()
	}
	return nil
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
