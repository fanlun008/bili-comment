package cmd

import (
	"fmt"
	"log"
	"time"

	"bili-comment/crawler"

	"github.com/spf13/cobra"
)

// CrawlerConfig 爬虫配置
type CrawlerConfig struct {
	BV           string        // BV号
	Mode         int           // 爬取模式 (2=最新, 3=热门)
	WithReplies  bool          // 是否爬取二级评论
	MaxPages     int           // 最大页数限制
	OutputPath   string        // 输出数据库路径
	CookiePath   string        // Cookie文件路径
	RequestDelay time.Duration // 请求间隔
}

// crawlCmd represents the crawl command
var crawlCmd = &cobra.Command{
	Use:   "crawl [BV号]",
	Short: "爬取B站视频评论",
	Long: `爬取指定BV号视频的评论数据，支持一级和二级评论爬取。

示例：
  bili-comment crawl BV1HW4y1n7BF                      # 基本用法
  bili-comment crawl BV1HW4y1n7BF --mode=3             # 爬取热门评论
  bili-comment crawl BV1HW4y1n7BF --with-replies=false # 不爬取二级评论
  bili-comment crawl BV1HW4y1n7BF --delay=1s           # 设置1秒请求延迟
  bili-comment crawl BV1HW4y1n7BF --output=/tmp/comments.db # 指定输出路径`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// 从命令行参数获取配置
		config := &CrawlerConfig{
			BV: args[0],
		}

		// 获取标志值
		config.Mode, _ = cmd.Flags().GetInt("mode")
		config.WithReplies, _ = cmd.Flags().GetBool("with-replies")
		config.MaxPages, _ = cmd.Flags().GetInt("max-pages")
		config.OutputPath, _ = cmd.Flags().GetString("output")
		config.CookiePath, _ = cmd.Flags().GetString("cookie")
		config.RequestDelay, _ = cmd.Flags().GetDuration("delay")

		return runCrawler(config)
	},
}

func runCrawler(config *CrawlerConfig) error {
	log.Println("B站评论爬虫启动...")

	// 转换配置格式
	crawlerConfig := &crawler.Config{
		BV:           config.BV,
		Mode:         config.Mode,
		WithReplies:  config.WithReplies,
		MaxPages:     config.MaxPages,
		OutputPath:   config.OutputPath,
		CookiePath:   config.CookiePath,
		RequestDelay: config.RequestDelay,
	}

	// 创建爬虫实例
	crawlerInstance, err := crawler.NewBilibiliCommentCrawler(crawlerConfig)
	if err != nil {
		return fmt.Errorf("创建爬虫失败: %v", err)
	}
	defer crawlerInstance.Close()

	// 获取视频信息
	oid, title, err := crawlerInstance.GetVideoInfo(config.BV)
	if err != nil {
		return fmt.Errorf("获取视频信息失败: %v", err)
	}

	log.Printf("开始爬取视频 %s 的评论，标题：%s", config.BV, title)
	log.Printf("爬取模式：%s", map[int]string{2: "最新", 3: "热门"}[config.Mode])
	log.Printf("是否爬取二级评论：%t", config.WithReplies)
	log.Printf("请求延迟：%v", config.RequestDelay)

	// 初始化变量
	nextPageID := ""
	count := 0

	// 开始爬取
	for {
		var err error
		nextPageID, count, err = crawlerInstance.CrawlComments(config.BV, oid, nextPageID, count, title, config.WithReplies)
		if err != nil {
			log.Printf("爬取评论失败: %v", err)
			break
		}

		if nextPageID == "" || nextPageID == "0" {
			log.Printf("评论爬取完成！总共爬取 %d 条评论", count)
			break
		}

		log.Printf("当前爬取 %d 条评论", count)
		time.Sleep(config.RequestDelay) // 使用配置的延迟
	}

	log.Printf("所有评论已保存到 SQLite 数据库：%s", config.OutputPath)
	return nil
}

func init() {
	rootCmd.AddCommand(crawlCmd)

	// 添加命令行参数
	crawlCmd.Flags().Int("mode", 2, "爬取模式 (2=最新评论, 3=热门评论)")
	crawlCmd.Flags().Bool("with-replies", true, "是否爬取二级评论")
	crawlCmd.Flags().Int("max-pages", 10, "二级评论最大页数限制 (0=无限制)")
	crawlCmd.Flags().String("output", "./data/crawler.db", "输出数据库文件路径")
	crawlCmd.Flags().String("cookie", "", "Cookie文件路径 (为空时自动查找)")
	crawlCmd.Flags().Duration("delay", 500*time.Millisecond, "请求间隔时间")
}
