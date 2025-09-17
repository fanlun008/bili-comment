package cmd

import (
	"fmt"
	"log"

	"bili-comment/gamersky"

	"github.com/spf13/cobra"
)

// queryGamerskyCmd represents the query gamersky command
var queryGamerskyCmd = &cobra.Command{
	Use:   "query-gamersky",
	Short: "查询已爬取的Gamersky新闻",
	Long: `查询数据库中已爬取的Gamersky新闻数据。

示例：
  bili-comment query-gamersky                    # 查询所有新闻
  bili-comment query-gamersky --limit=10         # 限制查询结果数量
  bili-comment query-gamersky --output=/tmp/news.db # 指定数据库路径`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 获取参数
		limit, _ := cmd.Flags().GetInt("limit")
		outputPath, _ := cmd.Flags().GetString("output")

		return queryGamerskyNews(outputPath, limit)
	},
}

func queryGamerskyNews(dbPath string, limit int) error {
	// 创建配置
	config := &gamersky.Config{
		OutputPath: dbPath,
	}

	// 创建爬虫实例（用于查询）
	crawlerInstance, err := gamersky.NewNewsCrawler(config)
	if err != nil {
		return fmt.Errorf("创建爬虫实例失败: %v", err)
	}
	defer crawlerInstance.Close()

	// 查询新闻
	news, err := crawlerInstance.QueryNews(limit)
	if err != nil {
		return fmt.Errorf("查询新闻失败: %v", err)
	}

	if len(news) == 0 {
		log.Println("数据库中没有找到新闻数据")
		return nil
	}

	// 显示结果
	log.Printf("找到 %d 条新闻记录：", len(news))
	for i, item := range news {
		fmt.Printf("\n%d. [%s] %s\n", i+1, item.SID, item.Title)
		fmt.Printf("   时间: %s\n", item.Time)
		fmt.Printf("   评论数: %d\n", item.CommentNum)
		fmt.Printf("   链接: %s\n", item.URL)
		if item.ImageURL != "" {
			fmt.Printf("   图片: %s\n", item.ImageURL)
		}
		fmt.Printf("   置顶时间: %s\n", item.TopLineTime)
		fmt.Printf("   创建时间: %s\n", item.CreateTime)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(queryGamerskyCmd)

	// 添加命令行参数
	queryGamerskyCmd.Flags().Int("limit", 20, "限制查询结果数量 (0=无限制)")
	queryGamerskyCmd.Flags().String("output", "./data/gamersky.db", "数据库文件路径")
}
