package cli

import (
	"database/sql"
	"fmt"

	"bili-comment/gamersky"
)

// DatabaseManager 数据库管理器
type DatabaseManager struct {
	db             *sql.DB
	newsCrawler    *gamersky.NewsCrawler
	commentCrawler *gamersky.CommentCrawler
}

// NewDatabaseManager 创建数据库管理器
func NewDatabaseManager(dbPath string) (*DatabaseManager, error) {
	config := &gamersky.Config{
		OutputPath: dbPath,
	}

	newsCrawler, err := gamersky.NewNewsCrawler(config)
	if err != nil {
		return nil, fmt.Errorf("创建新闻爬虫失败: %v", err)
	}

	commentCrawler, err := gamersky.NewCommentCrawler(config)
	if err != nil {
		newsCrawler.Close()
		return nil, fmt.Errorf("创建评论爬虫失败: %v", err)
	}

	return &DatabaseManager{
		newsCrawler:    newsCrawler,
		commentCrawler: commentCrawler,
	}, nil
}

// GetNewsList 获取新闻列表
func (dm *DatabaseManager) GetNewsList(offset, limit int) ([]NewsItem, error) {
	// 使用 QueryNewsWithOffset 方法支持分页
	newsInfos, err := dm.newsCrawler.QueryNewsWithOffset(offset, limit)
	if err != nil {
		return nil, err
	}

	var items []NewsItem
	for _, info := range newsInfos {
		items = append(items, NewsItem{
			ID:         info.SID,
			Title:      info.Title,
			Time:       info.Time,
			CommentNum: info.CommentNum,
			URL:        info.URL,
		})
	}

	return items, nil
}

// GetComments 获取评论列表
func (dm *DatabaseManager) GetComments(articleID string) ([]CommentItem, error) {
	// 使用现有的 QueryComments 方法
	comments, err := dm.commentCrawler.QueryComments(articleID, 0)
	if err != nil {
		return nil, err
	}

	var items []CommentItem
	for _, comment := range comments {
		items = append(items, CommentItem{
			ID:           comment.ID,
			Username:     comment.Username,
			Content:      comment.Content,
			Time:         comment.CommentTime,
			SupportCount: comment.SupportCount,
			ParentID:     comment.ParentID,
			AnswerToName: comment.AnswerToName,
			UserLevel:    comment.UserLevel,
			IPLocation:   comment.IPLocation,
		})
	}

	return items, nil
}

// Close 关闭数据库连接
func (dm *DatabaseManager) Close() error {
	if dm.newsCrawler != nil {
		dm.newsCrawler.Close()
	}
	if dm.commentCrawler != nil {
		dm.commentCrawler.Close()
	}
	return nil
}
