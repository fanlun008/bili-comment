package cmd

import (
	"fmt"
	"log"
	"time"

	"bili-comment/crawler"

	"github.com/spf13/cobra"
)

// GamerskyConfig Gamersky爬虫配置
type GamerskyConfig struct {
	Pages        int           // 爬取页数
	OutputPath   string        // 输出数据库路径
	RequestDelay time.Duration // 请求间隔
}

// gamerskyCmd represents the gamersky command
var gamerskyCmd = &cobra.Command{
	Use:   "gamersky",
	Short: "爬取Gamersky游戏天空网站新闻",
	Long: `爬取Gamersky游戏天空网站的新闻列表数据。

示例：
  bili-comment gamersky                           # 爬取第1页新闻
  bili-comment gamersky --pages=5                # 爬取前5页新闻
  bili-comment gamersky --delay=1s               # 设置1秒请求延迟
  bili-comment gamersky --output=/tmp/news.db    # 指定输出路径`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 从命令行参数获取配置
		config := &GamerskyConfig{}

		// 获取标志值
		config.Pages, _ = cmd.Flags().GetInt("pages")
		config.OutputPath, _ = cmd.Flags().GetString("output")
		config.RequestDelay, _ = cmd.Flags().GetDuration("delay")

		return runGamerskyrawler(config)
	},
}

func runGamerskyrawler(config *GamerskyConfig) error {
	log.Println("Gamersky新闻爬虫启动...")

	// 转换配置格式
	crawlerConfig := &crawler.Config{
		OutputPath:   config.OutputPath,
		RequestDelay: config.RequestDelay,
	}

	// 创建爬虫实例
	crawlerInstance, err := crawler.NewGamerskyNewsCrawler(crawlerConfig)
	if err != nil {
		return fmt.Errorf("创建爬虫失败: %v", err)
	}
	defer crawlerInstance.Close()

	log.Printf("开始爬取Gamersky新闻，页数：%d", config.Pages)
	log.Printf("请求延迟：%v", config.RequestDelay)

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

	log.Printf("爬取完成！总共爬取 %d 条新闻，已保存到 SQLite 数据库：%s", totalCount, config.OutputPath)
	return nil
}

func init() {
	rootCmd.AddCommand(gamerskyCmd)

	// 添加命令行参数
	gamerskyCmd.Flags().Int("pages", 1, "爬取页数")
	gamerskyCmd.Flags().String("output", "./data/gamersky.db", "输出数据库文件路径")
	gamerskyCmd.Flags().Duration("delay", 1*time.Second, "请求间隔时间")
}
