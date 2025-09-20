package cmd

import (
	"fmt"

	"bili-comment/cli"

	"github.com/spf13/cobra"
)

var (
	viewerDBPath string
)

// viewerCmd 新闻查看器命令
var viewerCmd = &cobra.Command{
	Use:   "viewer",
	Short: "启动新闻查看器",
	Long: `启动一个交互式的新闻查看器，可以：
- 按时间降序浏览新闻
- 支持分页（每页15条）
- 查看新闻评论
- 支持层级评论显示`,
	RunE: runViewer,
}

func init() {
	rootCmd.AddCommand(viewerCmd)

	viewerCmd.Flags().StringVar(&viewerDBPath, "db", "./data/gamersky.db", "数据库文件路径")
}

func runViewer(cmd *cobra.Command, args []string) error {
	fmt.Printf("正在启动新闻查看器，数据库: %s\n", viewerDBPath)

	viewer, err := cli.NewNewsViewer(viewerDBPath)
	if err != nil {
		return fmt.Errorf("创建新闻查看器失败: %v", err)
	}

	return viewer.Run()
}
