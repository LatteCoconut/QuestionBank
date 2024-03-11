package SpiderUtil

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"regexp"
	"strings"
)

func ExecuteSpider(urls []string, cookieRaw string) ([]Question, error) {
	// 初始化结构体列表
	questions := []Question{}
	titleMap := make(map[string]bool)

	// 创建一个新的collector实例
	c := colly.NewCollector()

	// 设置请求头
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("sec-ch-ua", `"Google Chrome";v="119", "Chromium";v="119", "Not?A_Brand";v="24"`)
		r.Headers.Set("sec-ch-ua-mobile", "?0")
		r.Headers.Set("sec-ch-ua-platform", `"macOS"`)
		r.Headers.Set("upgrade-insecure-requests", "1")
		r.Headers.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36")
		r.Headers.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		r.Headers.Set("sec-fetch-site", "same-origin")
		r.Headers.Set("sec-fetch-mode", "navigate")
		r.Headers.Set("sec-fetch-user", "?1")
		r.Headers.Set("sec-fetch-dest", "document")
		r.Headers.Set("referer", "https://member.zikao365.com/qzgckh/index/indexForGckhJk.shtm?cwID=zk5095920002&cwareID=515430&sourceFlag=jdcs")
		r.Headers.Set("accept-encoding", "gzip, deflate, br, zstd")
		r.Headers.Set("accept-language", "zh-CN,zh;q=0.9")
		// 设置Cookies
		r.Headers.Set("cookie", cookieRaw)
	})

	// OnHTML注册一个回调函数，该函数会在访问页面后自动执行。
	// 这里以提取页面中所有<a>标签的href属性为例。
	c.OnHTML("a[class='ckxq']", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		// 使用strings.Fields分割字符串，自动移除所有空白字符
		fields := strings.Fields(link)
		// 使用strings.Join将分割后的字符串片段重新连接，不添加任何分隔符
		cleanLink := strings.Join(fields, "")
		// 访问错题
		e.Request.Visit(cleanLink)
	})

	c.OnHTML("div.jx_tmtit", func(e *colly.HTMLElement) {
		//title := strings.TrimSpace(e.Text)

		// 同时使用正则表达式去除序号和所有空格
		re := regexp.MustCompile(`^\s*\d+、\s*|\s+`)
		title := re.ReplaceAllString(e.Text, "")

		// 检查题目是否已经存在
		if _, exists := titleMap[title]; exists {
			// 如果题目已存在，则跳过
			return
		}

		question := Question{
			Title: title,
		}

		e.DOM.NextAll().EachWithBreak(func(_ int, s *goquery.Selection) bool {
			// 如果是<p>标签，则处理选项
			if s.Is("p") {
				optionText := strings.TrimSpace(s.Text())
				question.Options = append(question.Options, optionText)
				return true // 继续遍历
			} else {
				answer := s.Find("span").Eq(1).Text()
				re := regexp.MustCompile("[A-Za-z]+")
				matches := re.FindString(answer)
				question.Correct = matches
				return false // 遇到div，停止遍历
			}
		})

		questions = append(questions, question)
		titleMap[title] = true
	})

	//c.OnResponse(func(response *colly.Response) {
	//	// 将响应体转换为字符串
	//	htmlContent := string(response.Body)
	//	// 打印HTML内容
	//	fmt.Println(htmlContent)
	//})

	// 处理访问URL时发生的错误
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	for _, url := range urls {
		err := c.Visit(url)
		if err != nil {
			fmt.Println("Error visiting URL:", err)
			return nil, err
		}
	}

	return questions, nil

}
