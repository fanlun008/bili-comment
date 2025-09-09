package cmd

import (
	"fmt"
	"log"
	"time"

	"bili-comment/crawler"

	"github.com/spf13/cobra"
)

// SearchConfig 搜索配置
type SearchConfig struct {
	Keyword      string        // 搜索关键词
	Page         int           // 页数
	PageSize     int           // 每页大小
	OutputPath   string        // 输出数据库路径
	CookiePath   string        // Cookie文件路径
	RequestDelay time.Duration // 请求间隔
}

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search [关键词]",
	Short: "搜索B站视频",
	Long: `根据关键词搜索B站视频，获取视频的基本信息。

示例：
  bili-comment search 极氪001                          # 基本用法
  bili-comment search 极氪001 --page=2                # 搜索第2页
  bili-comment search 极氪001 --page-size=20          # 设置每页20条结果
  bili-comment search 极氪001 --delay=1s              # 设置1秒请求延迟
  bili-comment search 极氪001 --output=/tmp/videos.db # 指定输出路径`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// 从命令行参数获取配置
		config := &SearchConfig{
			Keyword: args[0],
		}

		// 获取标志值
		config.Page, _ = cmd.Flags().GetInt("page")
		config.PageSize, _ = cmd.Flags().GetInt("page-size")
		config.OutputPath, _ = cmd.Flags().GetString("output")
		config.CookiePath, _ = cmd.Flags().GetString("cookie")
		config.RequestDelay, _ = cmd.Flags().GetDuration("delay")

		return runSearch(config)
	},
}

func runSearch(config *SearchConfig) error {
	log.Printf("B站视频搜索启动，关键词：%s", config.Keyword)

	// 转换配置格式
	crawlerConfig := &crawler.Config{
		OutputPath:   config.OutputPath,
		CookiePath:   config.CookiePath,
		RequestDelay: config.RequestDelay,
	}

	// 创建搜索实例
	searcher, err := crawler.NewBilibiliVideoSearcher(crawlerConfig)
	if err != nil {
		return fmt.Errorf("创建搜索器失败: %v", err)
	}
	defer searcher.Close()

	log.Printf("开始搜索关键词：%s", config.Keyword)
	log.Printf("页数：%d，每页大小：%d", config.Page, config.PageSize)
	log.Printf("请求延迟：%v", config.RequestDelay)

	// 执行搜索
	videos, err := searcher.SearchVideos(config.Keyword, config.Page, config.PageSize)
	if err != nil {
		return fmt.Errorf("搜索视频失败: %v", err)
	}

	// 保存结果到数据库
	savedCount := 0
	for _, video := range videos {
		if err := searcher.SaveVideoToDB(video); err != nil {
			log.Printf("保存视频信息失败: %v", err)
		} else {
			savedCount++
		}
	}

	log.Printf("搜索完成！共找到 %d 个视频，成功保存 %d 个", len(videos), savedCount)
	log.Printf("搜索结果已保存到 SQLite 数据库：%s", config.OutputPath)

	return nil
}

func init() {
	rootCmd.AddCommand(searchCmd)

	// 添加命令行参数
	searchCmd.Flags().Int("page", 1, "搜索页数 (默认第1页)")
	searchCmd.Flags().Int("page-size", 20, "每页结果数量 (默认20条)")
	searchCmd.Flags().String("output", "./data/crawler.db", "输出数据库文件路径")
	searchCmd.Flags().String("cookie", "", "Cookie文件路径 (为空时自动查找)")
	searchCmd.Flags().Duration("delay", 500*time.Millisecond, "请求间隔时间")
}
