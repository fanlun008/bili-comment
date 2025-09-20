package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

// NewsViewer 新闻查看器
type NewsViewer struct {
	db          *DatabaseManager
	pageSize    int
	currentNews []NewsItem
}

// NewsItem 新闻项
type NewsItem struct {
	ID         string
	Title      string
	Time       string
	CommentNum int
	URL        string
}

// CommentItem 评论项
type CommentItem struct {
	ID           int64
	Username     string
	Content      string
	Time         string
	SupportCount int
	ParentID     int64
	AnswerToName string
	UserLevel    int
	IPLocation   string
}

// NewNewsViewer 创建新闻查看器
func NewNewsViewer(dbPath string) (*NewsViewer, error) {
	db, err := NewDatabaseManager(dbPath)
	if err != nil {
		return nil, err
	}

	return &NewsViewer{
		db:       db,
		pageSize: 15,
	}, nil
}

// Run 运行新闻查看器
func (nv *NewsViewer) Run() error {
	defer nv.db.Close()

	fmt.Println()
	nv.printHeader()

	page := 1
	for {
		err := nv.showNewsPage(page)
		if err != nil {
			return err
		}

		action := nv.getAction()
		switch action.Type {
		case "next":
			page++
		case "prev":
			if page > 1 {
				page--
			}
		case "view":
			if action.Value > 0 && action.Value <= len(nv.currentNews) {
				err := nv.showNewsComments(nv.currentNews[action.Value-1])
				if err != nil {
					color.Red("❌ 查看评论失败: %v\n", err)
					nv.waitForEnter()
				}
			}
		case "quit":
			nv.printGoodbye()
			return nil
		case "invalid":
			color.Yellow("⚠️  无效命令，请重新输入\n")
		}
	}
}

// printHeader 打印标题头
func (nv *NewsViewer) printHeader() {
	color.Cyan("╔══════════════════════════════════════════════════════════════╗")
	color.Cyan("║                    🎮 Gamersky 新闻浏览器                      ║")
	color.Cyan("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

// showNewsPage 显示新闻页面
func (nv *NewsViewer) showNewsPage(page int) error {
	// 获取新闻数据
	newsList, err := nv.getNewsList(page)
	if err != nil {
		return fmt.Errorf("获取新闻列表失败: %v", err)
	}

	nv.currentNews = newsList

	if len(nv.currentNews) == 0 {
		color.Yellow("📰 暂无新闻数据")
		return nil
	}

	// 清屏
	fmt.Print("\033[2J\033[H")
	nv.printHeader()

	// 显示页面标题
	color.Blue("📄 第 %d 页 (共 %d 条新闻)\n", page, len(nv.currentNews))
	fmt.Println(strings.Repeat("─", 80))

	// 显示新闻列表
	for i, news := range nv.currentNews {
		nv.printNewsItem(i+1, news)
	}

	fmt.Println(strings.Repeat("─", 80))
	nv.printControls()
	return nil
}

// printNewsItem 打印单条新闻
func (nv *NewsViewer) printNewsItem(index int, news NewsItem) {
	// 新闻序号和标题
	color.White("🏷️  %d. ", index)
	color.Green("%s", news.Title)
	fmt.Println()

	// 新闻信息
	fmt.Print("   📅 ")
	color.Blue("%s", news.Time)

	if news.CommentNum > 0 {
		fmt.Print("   💬 ")
		color.Yellow("%d条评论", news.CommentNum)
	}

	fmt.Print("   🆔 ")
	color.Magenta("ID: %s", news.ID)
	fmt.Println()

	// URL（如果有）
	if news.URL != "" {
		fmt.Print("   🔗 ")
		color.Cyan("%s", news.URL)
		fmt.Println()
	}

	fmt.Println()
}

// printControls 打印控制说明
func (nv *NewsViewer) printControls() {
	color.Yellow("💡 操作说明:")
	fmt.Println("   🔢 输入数字 1-15: 查看对应新闻的评论")
	fmt.Println("   ⬅️  输入 'p' 或 'prev': 上一页")
	fmt.Println("   ➡️  输入 'n' 或 'next': 下一页")
	fmt.Println("   ❌ 输入 'q' 或 'quit': 退出")
	fmt.Print("\n请输入您的选择: ")
}

// Action 用户操作
type Action struct {
	Type  string // "next", "prev", "view", "quit", "invalid"
	Value int    // 仅对 "view" 有效
}

// getAction 获取用户操作
func (nv *NewsViewer) getAction() Action {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return Action{Type: "invalid"}
	}

	input = strings.TrimSpace(strings.ToLower(input))

	switch input {
	case "n", "next":
		return Action{Type: "next"}
	case "p", "prev":
		return Action{Type: "prev"}
	case "q", "quit":
		return Action{Type: "quit"}
	default:
		// 尝试解析为数字
		if num, err := strconv.Atoi(input); err == nil {
			if num >= 1 && num <= nv.pageSize {
				return Action{Type: "view", Value: num}
			}
		}
		return Action{Type: "invalid"}
	}
}

// getNewsList 获取新闻列表
func (nv *NewsViewer) getNewsList(page int) ([]NewsItem, error) {
	offset := (page - 1) * nv.pageSize
	return nv.db.GetNewsList(offset, nv.pageSize)
}

// showNewsComments 显示新闻评论
func (nv *NewsViewer) showNewsComments(news NewsItem) error {
	// 清屏
	fmt.Print("\033[2J\033[H")

	// 显示新闻标题
	color.Cyan("╔══════════════════════════════════════════════════════════════╗")
	color.Cyan("║                      📰 新闻评论                              ║")
	color.Cyan("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	color.Green("📰 %s", news.Title)
	fmt.Println()
	color.Blue("🆔 ID: %s  📅 %s  💬 %d条评论", news.ID, news.Time, news.CommentNum)
	fmt.Println()
	fmt.Println(strings.Repeat("═", 80))

	// 获取并显示评论
	comments, err := nv.getComments(news.ID)
	if err != nil {
		return fmt.Errorf("获取评论失败: %v", err)
	}

	if len(comments) == 0 {
		color.Yellow("💬 暂无评论")
	} else {
		nv.printComments(comments)
	}

	fmt.Println()
	fmt.Println(strings.Repeat("═", 80))
	color.Yellow("💡 按回车键返回新闻列表...")
	nv.waitForEnter()

	return nil
}

// getComments 获取评论列表
func (nv *NewsViewer) getComments(newsID string) ([]CommentItem, error) {
	return nv.db.GetComments(newsID)
}

// printComments 打印评论列表
func (nv *NewsViewer) printComments(comments []CommentItem) {
	for i, comment := range comments {
		nv.printComment(comment, i == len(comments)-1)
	}
}

// printComment 打印单条评论
func (nv *NewsViewer) printComment(comment CommentItem, isLast bool) {
	// 评论层级标识
	indent := ""
	if comment.ParentID != 0 {
		indent = "    "
		color.Blue("    ↳ 回复 @%s", comment.AnswerToName)
		fmt.Println()
	}

	// 用户信息行
	fmt.Print(indent)
	color.Magenta("👤 %s", comment.Username)
	fmt.Print("  ")
	color.Blue("Lv%d", comment.UserLevel)
	fmt.Print("  ")
	color.Cyan("📍 %s", comment.IPLocation)
	fmt.Print("  ")
	color.Green("📅 %s", comment.Time)

	if comment.SupportCount > 0 {
		fmt.Print("  ")
		color.Yellow("👍 %d", comment.SupportCount)
	}
	fmt.Println()

	// 评论内容
	fmt.Print(indent)
	color.White("💬 %s", comment.Content)
	fmt.Println()

	if !isLast {
		fmt.Println(indent + strings.Repeat("─", 60))
	}
}

// waitForEnter 等待用户按回车
func (nv *NewsViewer) waitForEnter() {
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
}

// printGoodbye 打印退出信息
func (nv *NewsViewer) printGoodbye() {
	fmt.Println()
	color.Green("👋 感谢使用 Gamersky 新闻浏览器！")
	color.Cyan("🎮 继续关注游戏资讯，再见！")
	fmt.Println()
}
