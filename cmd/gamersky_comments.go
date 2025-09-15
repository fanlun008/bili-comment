package cmd

import (
	"fmt"
	"log"
	"time"

	"bili-comment/crawler"

	"github.com/spf13/cobra"
)

// GamerskyCommentsConfig Gamersky评论爬虫配置
type GamerskyCommentsConfig struct {
	ArticleID    string        // 文章ID
	Pages        int           // 爬取页数
	OutputPath   string        // 输出数据库路径
	RequestDelay time.Duration // 请求间隔
}

// gamerskyCommentsCmd represents the gamersky-comments command
var gamerskyCommentsCmd = &cobra.Command{
	Use:   "gamersky-comments",
	Short: "爬取Gamersky游戏天空网站文章评论",
	Long: `爬取Gamersky游戏天空网站的文章评论数据。

示例：
  bili-comment gamersky-comments --article-id=2014209          # 爬取指定文章的评论
  bili-comment gamersky-comments --article-id=2014209 --pages=5  # 爬取前5页评论
  bili-comment gamersky-comments --article-id=2014209 --delay=1s # 设置1秒请求延迟
  bili-comment gamersky-comments --article-id=2014209 --output=/tmp/comments.db # 指定输出路径`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 从命令行参数获取配置
		config := &GamerskyCommentsConfig{}

		// 获取标志值
		config.ArticleID, _ = cmd.Flags().GetString("article-id")
		config.Pages, _ = cmd.Flags().GetInt("pages")
		config.OutputPath, _ = cmd.Flags().GetString("output")
		config.RequestDelay, _ = cmd.Flags().GetDuration("delay")

		// 验证必要参数
		if config.ArticleID == "" {
			return fmt.Errorf("必须指定 --article-id 参数")
		}

		// 确保延迟时间有默认值
		if config.RequestDelay == 0 {
			config.RequestDelay = 1 * time.Second
		}

		return runGamerskyCommentsCrawler(config)
	},
}

func runGamerskyCommentsCrawler(config *GamerskyCommentsConfig) error {
	log.Println("Gamersky评论爬虫启动...")

	// 转换配置格式
	crawlerConfig := &crawler.Config{
		OutputPath:   config.OutputPath,
		RequestDelay: config.RequestDelay,
	}

	// 创建爬虫实例
	crawlerInstance, err := crawler.NewGamerskyCommentCrawler(crawlerConfig)
	if err != nil {
		return fmt.Errorf("创建评论爬虫失败: %v", err)
	}
	defer crawlerInstance.Close()

	log.Printf("开始爬取文章 %s 的评论，页数：%d", config.ArticleID, config.Pages)
	log.Printf("请求延迟：%v", config.RequestDelay)

	// 开始爬取评论
	totalCount, err := crawlerInstance.CrawlComments(config.ArticleID, config.Pages)
	if err != nil {
		return fmt.Errorf("爬取评论失败: %v", err)
	}

	log.Printf("评论爬取完成！总共爬取 %d 条评论，已保存到 SQLite 数据库：%s", totalCount, config.OutputPath)
	return nil
}

func init() {
	rootCmd.AddCommand(gamerskyCommentsCmd)

	// 添加命令行参数
	gamerskyCommentsCmd.Flags().String("article-id", "", "文章ID（必需）")
	gamerskyCommentsCmd.Flags().Int("pages", 10, "爬取页数")
	gamerskyCommentsCmd.Flags().String("output", "./data/gamersky.db", "输出数据库文件路径")
	gamerskyCommentsCmd.Flags().Duration("delay", 1*time.Second, "请求间隔时间")

	// 标记必需的参数
	gamerskyCommentsCmd.MarkFlagRequired("article-id")
}
