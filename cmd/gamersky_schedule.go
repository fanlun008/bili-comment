package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bili-comment/gamersky"

	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
)

// GamerskyScheduleConfig 定时任务配置
type GamerskyScheduleConfig struct {
	Pages        int           // 每次爬取页数
	OutputPath   string        // 输出数据库路径
	RequestDelay time.Duration // 请求间隔
	CronSpec     string        // Cron表达式
}

// gamerskyScheduleCmd represents the gamersky schedule command
var gamerskyScheduleCmd = &cobra.Command{
	Use:   "gamersky-schedule",
	Short: "启动Gamersky新闻定时爬取任务",
	Long: `启动Gamersky新闻定时爬取任务，支持自定义Cron表达式。

默认每5分钟执行一次爬取任务。

示例：
  bili-comment gamersky-schedule                          # 每5分钟爬取1页新闻
  bili-comment gamersky-schedule --pages=3               # 每5分钟爬取3页新闻
  bili-comment gamersky-schedule --cron="0 */10 * * * *" # 每10分钟执行一次
  bili-comment gamersky-schedule --cron="0 0 */2 * * *"  # 每2小时执行一次

Cron表达式格式：秒 分 时 日 月 周
  * 每5分钟: "0 */5 * * * *"
  * 每小时: "0 0 * * * *"
  * 每天凌晨2点: "0 0 2 * * *"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 从命令行参数获取配置
		config := &GamerskyScheduleConfig{}

		// 获取标志值
		config.Pages, _ = cmd.Flags().GetInt("pages")
		config.OutputPath, _ = cmd.Flags().GetString("output")
		config.RequestDelay, _ = cmd.Flags().GetDuration("delay")
		config.CronSpec, _ = cmd.Flags().GetString("cron")

		// 确保延迟时间有默认值
		if config.RequestDelay == 0 {
			config.RequestDelay = 1 * time.Second
		}

		return runGamerskyScheduler(config)
	},
}

func runGamerskyScheduler(config *GamerskyScheduleConfig) error {
	log.Println("启动Gamersky新闻定时爬取服务...")
	log.Printf("定时规则: %s", config.CronSpec)
	log.Printf("每次爬取页数: %d", config.Pages)
	log.Printf("输出路径: %s", config.OutputPath)
	log.Printf("请求延迟: %v", config.RequestDelay)

	// 创建cron调度器
	c := cron.New(cron.WithSeconds()) // 支持秒级调度

	// 添加定时任务
	_, err := c.AddFunc(config.CronSpec, func() {
		log.Println("开始执行定时爬取任务...")
		startTime := time.Now()

		err := executeCrawlTask(config)
		if err != nil {
			log.Printf("定时爬取任务执行失败: %v", err)
		} else {
			duration := time.Since(startTime)
			log.Printf("定时爬取任务执行完成，耗时: %v", duration)
		}
	})

	if err != nil {
		return fmt.Errorf("添加定时任务失败: %v", err)
	}

	// 启动调度器
	c.Start()
	log.Println("定时任务调度器已启动")

	// 立即执行一次
	log.Println("立即执行一次爬取任务...")
	if err := executeCrawlTask(config); err != nil {
		log.Printf("初始爬取任务执行失败: %v", err)
	}

	// 等待信号来优雅关闭
	return waitForShutdown(c)
}

func executeCrawlTask(config *GamerskyScheduleConfig) error {
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

	log.Printf("本次爬取完成！总共爬取 %d 条新闻", totalCount)
	return nil
}

func waitForShutdown(c *cron.Cron) error {
	// 创建一个接收系统信号的通道
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("定时任务正在运行中... (按 Ctrl+C 退出)")

	// 等待信号
	sig := <-sigChan
	log.Printf("收到信号 %v，开始优雅关闭...", sig)

	// 创建一个带超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 停止调度器
	stopCtx := c.Stop()

	select {
	case <-stopCtx.Done():
		log.Println("调度器已优雅关闭")
	case <-ctx.Done():
		log.Println("关闭超时，强制退出")
	}

	log.Println("定时任务服务已关闭")
	return nil
}

func init() {
	rootCmd.AddCommand(gamerskyScheduleCmd)

	// 添加命令行参数
	gamerskyScheduleCmd.Flags().Int("pages", 1, "每次爬取的页数")
	gamerskyScheduleCmd.Flags().String("output", "./data/gamersky.db", "输出数据库文件路径")
	gamerskyScheduleCmd.Flags().Duration("delay", 1*time.Second, "请求间隔时间")
	gamerskyScheduleCmd.Flags().String("cron", "0 */5 * * * *", "Cron表达式 (秒 分 时 日 月 周)，默认每5分钟")
}
