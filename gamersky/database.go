package gamersky

import (
	"database/sql"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// GetDBConnection 获取Gamersky数据库连接
func GetDBConnection(dbPath string) (*sql.DB, error) {
	if dbPath == "" {
		dbPath = "./data/gamersky.db"
	}

	// 创建目录
	dir := strings.TrimSuffix(dbPath, "/gamersky.db")
	if dir != dbPath {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	// 连接数据库
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
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
		return nil, err
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
		answer_to_id INTEGER DEFAULT 0,
		answer_to_name TEXT DEFAULT '',
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
		return nil, err
	}

	return db, nil
}
