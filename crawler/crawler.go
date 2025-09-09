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
