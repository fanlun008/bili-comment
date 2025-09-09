package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

// queryCmd represents the query command
var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "查询爬取的评论数据",
	Long: `查询已爬取的评论数据，支持多种查询条件。

示例：
  bili-comment query --count           # 统计评论总数
  bili-comment query --list=10         # 显示前10条评论
  bili-comment query --bv=BV1xxx       # 查询指定视频的评论
  bili-comment query --user="用户名"    # 查询指定用户的评论`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath, _ := cmd.Flags().GetString("db")
		showCount, _ := cmd.Flags().GetBool("count")
		listLimit, _ := cmd.Flags().GetInt("list")
		bv, _ := cmd.Flags().GetString("bv")
		user, _ := cmd.Flags().GetString("user")

		return runQuery(dbPath, showCount, listLimit, bv, user)
	},
}

func runQuery(dbPath string, showCount bool, listLimit int, bv, user string) error {
	if dbPath == "" {
		dbPath = "./data/crawler.db"
	}

	// 连接数据库
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 统计评论总数
	if showCount {
		var total int
		query := "SELECT COUNT(*) FROM bilibili_comments"
		args := []interface{}{}

		if bv != "" {
			query += " WHERE 视频BV号 = ?"
			args = append(args, bv)
		} else if user != "" {
			query += " WHERE 用户名 = ?"
			args = append(args, user)
		}

		err := db.QueryRow(query, args...).Scan(&total)
		if err != nil {
			return fmt.Errorf("查询失败: %v", err)
		}

		fmt.Printf("数据库中共有 %d 条评论\n", total)
		return nil
	}

	// 列出评论
	if listLimit > 0 {
		query := `
		SELECT 序号, 用户名, 评论内容, 评论时间, 点赞数, 回复数, 视频BV号 
		FROM bilibili_comments 
		`
		args := []interface{}{}

		if bv != "" {
			query += " WHERE 视频BV号 = ?"
			args = append(args, bv)
		} else if user != "" {
			query += " WHERE 用户名 = ?"
			args = append(args, user)
		}

		query += " ORDER BY 序号 LIMIT ?"
		args = append(args, listLimit)

		rows, err := db.Query(query, args...)
		if err != nil {
			return fmt.Errorf("查询失败: %v", err)
		}
		defer rows.Close()

		fmt.Printf("%-6s %-20s %-50s %-20s %-8s %-8s %-15s\n",
			"序号", "用户名", "评论内容", "评论时间", "点赞数", "回复数", "BV号")
		fmt.Println(strings.Repeat("-", 120))

		for rows.Next() {
			var serialNumber, likeCount, replyCount int
			var username, content, commentTime, bvNum string

			err := rows.Scan(&serialNumber, &username, &content, &commentTime, &likeCount, &replyCount, &bvNum)
			if err != nil {
				log.Printf("读取行数据失败: %v", err)
				continue
			}

			// 截断过长的内容
			if len(content) > 47 {
				content = content[:47] + "..."
			}
			if len(username) > 17 {
				username = username[:17] + "..."
			}

			fmt.Printf("%-6d %-20s %-50s %-20s %-8d %-8d %-15s\n",
				serialNumber, username, content, commentTime, likeCount, replyCount, bvNum)
		}

		return nil
	}

	return fmt.Errorf("请指定查询参数: --count 或 --list")
}

func init() {
	rootCmd.AddCommand(queryCmd)

	queryCmd.Flags().String("db", "./data/crawler.db", "数据库文件路径")
	queryCmd.Flags().Bool("count", false, "统计评论总数")
	queryCmd.Flags().Int("list", 0, "显示指定数量的评论列表")
	queryCmd.Flags().String("bv", "", "查询指定BV号的评论")
	queryCmd.Flags().String("user", "", "查询指定用户的评论")
}
