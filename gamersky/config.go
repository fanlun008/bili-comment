package gamersky

import (
	"time"
)

// Config Gamersky爬虫配置
type Config struct {
	OutputPath   string        // 输出数据库路径
	RequestDelay time.Duration // 请求间隔
}
