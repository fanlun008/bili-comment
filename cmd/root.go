package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bili-comment",
	Short: "B站评论爬虫工具",
	Long: `一个用于爬取B站视频评论和搜索视频的命令行工具。
支持爬取一级和二级评论，可配置爬取模式、输出路径等参数。
同时支持根据关键词搜索B站视频。

示例用法：
  bili-comment crawl BV1HW4y1n7BF                    # 爬取指定视频的评论
  bili-comment crawl BV1HW4y1n7BF --mode=3           # 爬取热门评论
  bili-comment crawl BV1HW4y1n7BF --with-replies=false # 不爬取二级评论
  bili-comment search 极氪001                        # 搜索关键词相关的视频
  bili-comment search 极氪001 --page=2               # 搜索第2页结果`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// 这里可以添加全局标志
}
