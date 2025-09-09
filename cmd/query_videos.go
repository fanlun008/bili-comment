package cmd

import (
	"fmt"
	"log"

	"bili-comment/crawler"

	"github.com/spf13/cobra"
)

// QueryConfig 查询配置
type QueryConfig struct {
	Keyword    string // 搜索关键词
	List       int    // 列出数量
	OutputPath string // 数据库路径
}

// queryVideosCmd represents the query videos command
var queryVideosCmd = &cobra.Command{
	Use:   "query-videos",
	Short: "查询数据库中的视频信息",
	Long: `查询数据库中保存的视频搜索结果。

示例：
  bili-comment query-videos                           # 查看所有视频
  bili-comment query-videos --keyword=极氪001         # 查看特定关键词的视频
  bili-comment query-videos --list=10                # 列出前10条视频
  bili-comment query-videos --keyword=特斯拉 --list=5 # 查看特斯拉相关的前5条视频`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 从命令行参数获取配置
		config := &QueryConfig{}

		// 获取标志值
		config.Keyword, _ = cmd.Flags().GetString("keyword")
		config.List, _ = cmd.Flags().GetInt("list")
		config.OutputPath, _ = cmd.Flags().GetString("output")

		return runQueryVideos(config)
	},
}

func runQueryVideos(config *QueryConfig) error {
	log.Printf("查询数据库中的视频信息...")

	// 转换配置格式
	crawlerConfig := &crawler.Config{
		OutputPath: config.OutputPath,
	}

	// 创建搜索实例
	searcher, err := crawler.NewBilibiliVideoSearcher(crawlerConfig)
	if err != nil {
		return fmt.Errorf("创建搜索器失败: %v", err)
	}
	defer searcher.Close()

	// 查询视频
	videos, err := searcher.QueryVideos(config.Keyword, config.List)
	if err != nil {
		return fmt.Errorf("查询视频失败: %v", err)
	}

	if len(videos) == 0 {
		log.Println("没有找到匹配的视频记录")
		return nil
	}

	// 输出结果
	fmt.Printf("\n找到 %d 条视频记录：\n\n", len(videos))

	for i, video := range videos {
		fmt.Printf("=== 记录 %d ===\n", i+1)
		fmt.Printf("关键词: %s\n", video.Keyword)
		fmt.Printf("BVID: %s\n", video.BVID)
		fmt.Printf("标题: %s\n", video.Title)
		fmt.Printf("作者: %s\n", video.Author)
		fmt.Printf("播放量: %d\n", video.Play)
		fmt.Printf("点赞数: %d\n", video.Like)
		fmt.Printf("时长: %s\n", video.Duration)
		fmt.Printf("收藏数: %d\n", video.Favorites)
		fmt.Printf("评论数: %d\n", video.VideoReview)
		fmt.Printf("弹幕数: %d\n", video.Danmaku)
		fmt.Printf("创建时间: %s\n", video.CreateTime)
		if video.Description != "" && video.Description != "-" {
			fmt.Printf("描述: %s\n", video.Description)
		}
		fmt.Println()
	}

	fmt.Printf("\n共显示 %d 条记录\n", len(videos))

	return nil
}

func init() {
	rootCmd.AddCommand(queryVideosCmd)

	// 添加命令行参数
	queryVideosCmd.Flags().String("keyword", "", "查询特定关键词的视频 (为空则查询所有)")
	queryVideosCmd.Flags().Int("list", 20, "列出的记录数量 (默认20条)")
	queryVideosCmd.Flags().String("output", "./data/crawler.db", "数据库文件路径")
}
