package cmd

import (
	"fmt"
	"log"
	"time"

	"bili-comment/gamersky"

	"github.com/spf13/cobra"
)

// GamerskyOnceConfig 单次爬取配置
type GamerskyOnceConfig struct {
	Pages        int           // 爬取页数
	OutputPath   string        // 输出数据库路径
	RequestDelay time.Duration // 请求间隔
}

// gamerskyOnceCmd represents the gamersky-once command
var gamerskyOnceCmd = &cobra.Command{
	Use:   "gamersky-once",
	Short: "执行一次Gamersky新闻爬取任务",
	Long: `执行一次Gamersky新闻爬取任务，适用于GitHub Actions等CI/CD环境。

与gamersky-schedule不同，此命令执行完成后立即退出，不会持续运行。

示例：
  bili-comment gamersky-once                           # 爬取1页新闻
  bili-comment gamersky-once --pages=5                # 爬取5页新闻
  bili-comment gamersky-once --delay=2s               # 设置2秒请求延迟
  bili-comment gamersky-once --output=/tmp/news.db    # 指定输出路径`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 从命令行参数获取配置
		config := &GamerskyOnceConfig{}

		// 获取标志值
		config.Pages, _ = cmd.Flags().GetInt("pages")
		config.OutputPath, _ = cmd.Flags().GetString("output")
		config.RequestDelay, _ = cmd.Flags().GetDuration("delay")

		// 确保延迟时间有默认值
		if config.RequestDelay == 0 {
			config.RequestDelay = 1 * time.Second
		}

		return runGamerskyOnce(config)
	},
}

func runGamerskyOnce(config *GamerskyOnceConfig) error {
	log.Println("开始执行Gamersky新闻爬取任务...")
	startTime := time.Now()

	// 转换配置格式
	crawlerConfig := &gamersky.Config{
		OutputPath:   config.OutputPath,
		RequestDelay: config.RequestDelay,
	}

	// 创建爬虫实例
	crawlerInstance, err := gamersky.NewNewsCrawler(crawlerConfig)
	if err != nil {
		return fmt.Errorf("创建爬虫失败: %v", err)
	}
	defer crawlerInstance.Close()

	log.Printf("开始爬取Gamersky新闻，页数：%d", config.Pages)
	log.Printf("请求延迟：%v", config.RequestDelay)
	log.Printf("输出路径：%s", config.OutputPath)

	// 开始爬取
	totalCount := 0
	for page := 1; page <= config.Pages; page++ {
		log.Printf("正在爬取第 %d 页...", page)

		count, err := crawlerInstance.CrawlNews(page)
		if err != nil {
			log.Printf("爬取第 %d 页失败: %v", page, err)
			continue
		}

		totalCount += count
		log.Printf("第 %d 页爬取完成，新增 %d 条新闻", page, count)

		// 如果没有新闻了，停止爬取
		if count == 0 {
			log.Printf("第 %d 页没有新闻，停止爬取", page)
			break
		}

		// 延迟
		if page < config.Pages {
			time.Sleep(config.RequestDelay)
		}
	}

	duration := time.Since(startTime)
	log.Printf("爬取完成！总共爬取 %d 条新闻，耗时：%v", totalCount, duration)
	log.Printf("数据已保存到：%s", config.OutputPath)

	return nil
}

func init() {
	rootCmd.AddCommand(gamerskyOnceCmd)

	// 添加命令行参数
	gamerskyOnceCmd.Flags().Int("pages", 3, "爬取页数")
	gamerskyOnceCmd.Flags().String("output", "./data/gamersky.db", "输出数据库文件路径")
	gamerskyOnceCmd.Flags().Duration("delay", 1*time.Second, "请求间隔时间")
}