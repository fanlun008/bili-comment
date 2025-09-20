package gamersky

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

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
				RootID                   int64  `json:"rootId"`
				ReplyID                  int64  `json:"replyId"`
				CreateTime               int64  `json:"createTime"`
				ReplyContent             string `json:"replyContent"`
				PraisesCount             int    `json:"praisesCount"`
				UserID                   int    `json:"userId"`
				UserName                 string `json:"userName"`
				UserHeadImageURL         string `json:"userHeadImageURL"`
				DeviceName               string `json:"deviceName"`
				UserLevel                int    `json:"userLevel"`
				UserAuthentication       string `json:"userAuthentication"`
				UserGroupID              int    `json:"userGroupId"`
				ThirdPlatformBound       string `json:"thirdPlatformBound"`
				ObjectCommentID          int64  `json:"objectCommentId"`
				ObjectUserID             int    `json:"objectUserId"`
				ObjectUserName           string `json:"objectUserName"`
				ObjectUserHeadImageURL   string `json:"objectUserHeadImageURL"`
				ObjectUserGroupID        int    `json:"objectUserGroupId"`
				ObjectUserAuthentication string `json:"objectUserAuthentication"`
				ObjectUserLevel          int    `json:"objectUserLevel"`
				IsAuthor                 bool   `json:"is_author"`
				IPLocation               string `json:"ip_location"`
				BeAuthorPraise           bool   `json:"beAuthorPraise"`
			} `json:"replies"`
			RepliesCount int `json:"repliesCount"`
		} `json:"comments"`
	} `json:"result"`
}

// CommentCrawler Gamersky评论爬虫结构体
type CommentCrawler struct {
	db     *sql.DB
	config *Config
}

// NewCommentCrawler 创建新的Gamersky评论爬虫实例
func NewCommentCrawler(config *Config) (*CommentCrawler, error) {
	// 初始化数据库
	db, err := GetDBConnection(config.OutputPath)
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %v", err)
	}

	return &CommentCrawler{
		db:     db,
		config: config,
	}, nil
}

// CrawlComments 爬取指定文章的评论
func (gcc *CommentCrawler) CrawlComments(articleID string, maxPages int) (int, error) {
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
func (gcc *CommentCrawler) crawlCommentsPage(articleID string, pageIndex int) (int, bool, error) {
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
		gamerskyComment := &Comment{
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
			if reply.ReplyID == 0 || reply.UserName == "" {
				log.Printf("跳过无效回复数据: ID=%d, UserName=%s", reply.ReplyID, reply.UserName)
				continue
			}

			replyComment := &Comment{
				ID:                 reply.ReplyID,
				ArticleID:          articleID,
				UserID:             reply.UserID,
				Username:           reply.UserName,
				Content:            reply.ReplyContent,
				CommentTime:        time.Unix(reply.CreateTime/1000, 0).Format("2006-01-02 15:04:05"),
				SupportCount:       reply.PraisesCount,
				ReplyCount:         0,                 // 二级评论通常没有回复数
				ParentID:           comment.CommentID, // 父评论ID
				UserAvatar:         reply.UserHeadImageURL,
				UserLevel:          reply.UserLevel,
				IPLocation:         reply.IPLocation,
				DeviceName:         reply.DeviceName,
				FloorNumber:        0,     // 回复通常没有楼层号
				IsTuijian:          false, // 回复结构中没有此字段
				IsAuthor:           reply.IsAuthor,
				IsBest:             false, // 回复结构中没有此字段
				UserAuthentication: reply.UserAuthentication,
				UserGroupID:        reply.UserGroupID,
				ThirdPlatformBound: reply.ThirdPlatformBound,
				CreateTime:         time.Now().Format("2006-01-02 15:04:05"),
			}

			if err := gcc.saveCommentToDB(replyComment); err != nil {
				log.Printf("保存回复失败 (ID: %d): %v", reply.ReplyID, err)
			} else {
				count++
				log.Printf("保存回复: %d - %s", reply.ReplyID, reply.UserName)
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
func (gcc *CommentCrawler) saveCommentToDB(comment *Comment) error {
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
func (gcc *CommentCrawler) QueryComments(articleID string, limit int) ([]Comment, error) {
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

	var comments []Comment
	for rows.Next() {
		var comment Comment
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
func (gcc *CommentCrawler) Close() error {
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
