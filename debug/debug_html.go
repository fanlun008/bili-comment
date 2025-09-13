package debug
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"
)

func main() {
	url := "https://wap.gamersky.com/"
	
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)
		return
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		return
	}

	content := string(body)
	
	// 保存HTML到文件进行分析
	err = os.WriteFile("debug_output.html", body, 0644)
	if err != nil {
		fmt.Printf("保存文件失败: %v\n", err)
	} else {
		fmt.Println("HTML内容已保存到 debug_output.html")
	}
	
	fmt.Printf("HTML内容长度: %d 字符\n", len(content))
	
	// 查找data-id模式
	dataIdPattern := `data-id="([^"]*)"`
	dataIdRegex := regexp.MustCompile(dataIdPattern)
	dataIdMatches := dataIdRegex.FindAllStringSubmatch(content, -1)
	
	fmt.Printf("找到 %d 个 data-id 匹配项:\n", len(dataIdMatches))
	for i, match := range dataIdMatches {
		if i < 10 { // 只显示前10个
			fmt.Printf("  %d: %s\n", i+1, match[1])
		}
	}
	
	// 查找titleAndTime模式  
	titleTimePattern := `<div class="titleAndTime">(.*?)</div>`
	titleTimeRegex := regexp.MustCompile(titleTimePattern)
	titleTimeMatches := titleTimeRegex.FindAllStringSubmatch(content, 5) // 只取前5个
	
	fmt.Printf("\n找到 %d 个 titleAndTime 匹配项:\n", len(titleTimeMatches))
	for i, match := range titleTimeMatches {
		fmt.Printf("  匹配项 %d:\n%s\n\n", i+1, match[1])
	}
}