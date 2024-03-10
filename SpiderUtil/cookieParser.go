package SpiderUtil

import (
	"errors"
	"strings"
)

func ParseCookie(rawCookies string) (string, error) {
	// 定义必须存在的cookie名称,"tgw_l7_route",
	requiredKeys := []string{"username", "hd_uid", "JSESSIONID", "cdeluid"}

	// 分割字符串为行
	lines := strings.Split(rawCookies, "\n")

	// 使用map来跟踪哪些必需的key已被找到
	foundKeys := make(map[string]bool)

	// 初始化一个字符串构建器
	var sb strings.Builder

	// 遍历每行，提取cookie名称和值
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		cookieName := parts[0]
		cookieValue := parts[1]

		// 检查cookieName是否是我们需要的key之一，如果是，标记为找到
		for _, key := range requiredKeys {
			if key == cookieName {
				foundKeys[key] = true
				break
			}
		}

		// 向字符串构建器中添加cookie名称和值
		if sb.Len() > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(cookieName)
		sb.WriteString("=")
		sb.WriteString(cookieValue)
	}
	sb.WriteString(";")

	// 检查是否所有必需的key都已找到
	for _, key := range requiredKeys {
		if !foundKeys[key] {
			// 如果任何一个key没有找到，返回错误
			return "", errors.New("missing required cookie key: " + key)
		}
	}

	// 获取最终的cookie字符串
	finalCookies := sb.String()

	return finalCookies, nil
}
