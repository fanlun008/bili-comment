package cmd

import (
	"fmt"
	"log"
	"strings"

	"bili-comment/crawler"

	"github.com/spf13/cobra"
)

// queryGamerskyCommentsCmd represents the query-gamersky-comments command
var queryGamerskyCommentsCmd = &cobra.Command{
	Use:   "query-gamersky-comments",
	Short: "查询Gamersky评论数据",
	Long: `查询已爬取的Gamersky评论数据。

示例：
  bili-comment query-gamersky-comments                           # 查询所有评论（默认限制20条）
  bili-comment query-gamersky-comments --article-id=2014209     # 查询指定文章的评论
  bili-comment query-gamersky-comments --limit=50               # 查询50条评论
  bili-comment query-gamersky-comments --article-id=2014209 --limit=100 # 查询指定文章的100条评论`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 从命令行参数获取配置
		articleID, _ := cmd.Flags().GetString("article-id")
		limit, _ := cmd.Flags().GetInt("limit")
		outputPath, _ := cmd.Flags().GetString("output")

		return runQueryGamerskyComments(articleID, limit, outputPath)
	},
}

func runQueryGamerskyComments(articleID string, limit int, outputPath string) error {
	log.Println("开始查询Gamersky评论数据...")

	// 创建配置
	config := &crawler.Config{
		OutputPath: outputPath,
	}

	// 创建爬虫实例
	crawlerInstance, err := crawler.NewGamerskyCommentCrawler(config)
	if err != nil {
		return fmt.Errorf("创建爬虫实例失败: %v", err)
	}
	defer crawlerInstance.Close()

	// 查询评论
	comments, err := crawlerInstance.QueryComments(articleID, limit)
	if err != nil {
		return fmt.Errorf("查询评论失败: %v", err)
	}

	// 显示结果
	if len(comments) == 0 {
		fmt.Println("没有找到评论数据")
		return nil
	}

	fmt.Printf("找到 %d 条评论：\n", len(comments))
	fmt.Println(strings.Repeat("=", 80))

	for i, comment := range comments {
		fmt.Printf("评论 %d:\n", i+1)
		fmt.Printf("  ID: %d\n", comment.ID)
		fmt.Printf("  文章ID: %s\n", comment.ArticleID)
		fmt.Printf("  用户: %s (等级: %d)\n", comment.Username, comment.UserLevel)

		// 处理评论内容显示（如果太长则截断）
		content := comment.Content
		if len(content) > 100 {
			content = content[:100] + "..."
		}
		fmt.Printf("  内容: %s\n", content)

		fmt.Printf("  时间: %s\n", comment.CommentTime)
		fmt.Printf("  点赞数: %d\n", comment.SupportCount)

		if comment.ParentID > 0 {
			fmt.Printf("  回复评论ID: %d\n", comment.ParentID)
		}

		if comment.ReplyCount > 0 {
			fmt.Printf("  回复数: %d\n", comment.ReplyCount)
		}

		fmt.Printf("  记录时间: %s\n", comment.CreateTime)
		fmt.Println(strings.Repeat("-", 80))
	}

	return nil
}

func init() {
	rootCmd.AddCommand(queryGamerskyCommentsCmd)

	// 添加命令行参数
	queryGamerskyCommentsCmd.Flags().String("article-id", "", "文章ID（可选，不指定则查询所有文章的评论）")
	queryGamerskyCommentsCmd.Flags().Int("limit", 20, "限制返回的评论数量")
	queryGamerskyCommentsCmd.Flags().String("output", "./data/gamersky.db", "数据库文件路径")
}
