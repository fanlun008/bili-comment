package cmd

import (
	"fmt"
	"log"
	"time"

	"bili-comment/gamersky"

	"github.com/spf13/cobra"
)

// GamerskyFullConfig 完整爬取配置
type GamerskyFullConfig struct {
	NewsPages    int           // 爬取新闻页数
	CommentPages int           // 每条新闻爬取的评论页数
	OutputPath   string        // 输出数据库路径
	RequestDelay time.Duration // 请求间隔
}

// gamerskyFullCmd represents the gamersky-full command
var gamerskyFullCmd = &cobra.Command{
	Use:   "gamersky-full",
	Short: "完整爬取Gamersky新闻和评论",
	Long: `完整爬取Gamersky新闻和对应的评论数据，存储在同一个SQLite数据库中。

此命令会先爬取指定页数的新闻，然后为每条新闻爬取指定页数的评论。
适用于GitHub Actions等CI/CD环境的一次性完整数据爬取。

示例：
  bili-comment gamersky-full                                    # 爬取3页新闻，每条新闻3页评论
  bili-comment gamersky-full --news-pages=5 --comment-pages=2  # 爬取5页新闻，每条新闻2页评论
  bili-comment gamersky-full --delay=2s                        # 设置2秒请求延迟
  bili-comment gamersky-full --output=/tmp/full.db             # 指定输出路径`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 从命令行参数获取配置
		config := &GamerskyFullConfig{}

		// 获取标志值
		config.NewsPages, _ = cmd.Flags().GetInt("news-pages")
		config.CommentPages, _ = cmd.Flags().GetInt("comment-pages")
		config.OutputPath, _ = cmd.Flags().GetString("output")
		config.RequestDelay, _ = cmd.Flags().GetDuration("delay")

		// 确保延迟时间有默认值
		if config.RequestDelay == 0 {
			config.RequestDelay = 1 * time.Second
		}

		return runGamerskyFull(config)
	},
}

func runGamerskyFull(config *GamerskyFullConfig) error {
	log.Println("开始执行Gamersky完整爬取任务（新闻+评论）...")
	startTime := time.Now()

	// 转换配置格式
	crawlerConfig := &gamersky.Config{
		OutputPath:   config.OutputPath,
		RequestDelay: config.RequestDelay,
	}

	log.Printf("配置信息：")
	log.Printf("  新闻页数：%d", config.NewsPages)
	log.Printf("  每条新闻评论页数：%d", config.CommentPages)
	log.Printf("  请求延迟：%v", config.RequestDelay)
	log.Printf("  输出路径：%s", config.OutputPath)

	// 第一步：爬取新闻
	log.Println("\n=== 第一步：爬取新闻 ===")
	newsCount, newsSids, err := crawlNews(crawlerConfig, config.NewsPages)
	if err != nil {
		return fmt.Errorf("爬取新闻失败: %v", err)
	}

	log.Printf("新闻爬取完成！总共爬取 %d 条新闻", newsCount)

	// 如果没有新闻或不需要爬取评论，直接返回
	if len(newsSids) == 0 || config.CommentPages <= 0 {
		duration := time.Since(startTime)
		log.Printf("任务完成！总耗时：%v", duration)
		return nil
	}

	// 第二步：为每条新闻爬取评论
	log.Println("\n=== 第二步：爬取评论 ===")
	totalComments, err := crawlCommentsForNews(crawlerConfig, newsSids, config.CommentPages)
	if err != nil {
		return fmt.Errorf("爬取评论失败: %v", err)
	}

	duration := time.Since(startTime)
	log.Printf("\n=== 任务完成！===")
	log.Printf("总共爬取 %d 条新闻，%d 条评论", newsCount, totalComments)
	log.Printf("总耗时：%v", duration)
	log.Printf("数据已保存到：%s", config.OutputPath)

	return nil
}

// crawlNews 爬取新闻并返回新闻ID列表
func crawlNews(config *gamersky.Config, pages int) (int, []string, error) {
	// 创建新闻爬虫实例
	newsCrawler, err := gamersky.NewNewsCrawler(config)
	if err != nil {
		return 0, nil, fmt.Errorf("创建新闻爬虫失败: %v", err)
	}
	defer newsCrawler.Close()

	totalCount := 0
	var allSids []string

	for page := 1; page <= pages; page++ {
		log.Printf("正在爬取新闻第 %d 页...", page)

		count, err := newsCrawler.CrawlNews(page)
		if err != nil {
			log.Printf("爬取新闻第 %d 页失败: %v", page, err)
			continue
		}

		totalCount += count
		log.Printf("新闻第 %d 页爬取完成，新增 %d 条新闻", page, count)

		// 如果没有新闻了，停止爬取
		if count == 0 {
			log.Printf("新闻第 %d 页没有新闻，停止爬取", page)
			break
		}

		// 延迟
		if page < pages {
			time.Sleep(config.RequestDelay)
		}
	}

	// 查询所有新闻的SID
	news, err := newsCrawler.QueryNews(0) // 0表示查询所有
	if err != nil {
		return totalCount, nil, fmt.Errorf("查询新闻失败: %v", err)
	}

	for _, item := range news {
		allSids = append(allSids, item.SID)
	}

	log.Printf("获取到 %d 条新闻的SID", len(allSids))
	return totalCount, allSids, nil
}

// crawlCommentsForNews 为新闻列表爬取评论
func crawlCommentsForNews(config *gamersky.Config, newsSids []string, commentPages int) (int, error) {
	// 创建评论爬虫实例
	commentCrawler, err := gamersky.NewCommentCrawler(config)
	if err != nil {
		return 0, fmt.Errorf("创建评论爬虫失败: %v", err)
	}
	defer commentCrawler.Close()

	totalComments := 0

	for i, sid := range newsSids {
		log.Printf("正在爬取新闻 %s 的评论 (%d/%d)...", sid, i+1, len(newsSids))

		count, err := commentCrawler.CrawlComments(sid, commentPages)
		if err != nil {
			log.Printf("爬取新闻 %s 评论失败: %v", sid, err)
			continue
		}

		totalComments += count
		log.Printf("新闻 %s 评论爬取完成，新增 %d 条评论", sid, count)

		// 在每条新闻之间延迟
		if i < len(newsSids)-1 {
			time.Sleep(config.RequestDelay)
		}
	}

	return totalComments, nil
}

func init() {
	rootCmd.AddCommand(gamerskyFullCmd)

	// 添加命令行参数
	gamerskyFullCmd.Flags().Int("news-pages", 3, "爬取新闻页数")
	gamerskyFullCmd.Flags().Int("comment-pages", 3, "每条新闻爬取的评论页数")
	gamerskyFullCmd.Flags().String("output", "./data/gamersky.db", "输出数据库文件路径")
	gamerskyFullCmd.Flags().Duration("delay", 1*time.Second, "请求间隔时间")
}
