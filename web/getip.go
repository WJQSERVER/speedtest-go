package web

import (
	"encoding/json"
	"regexp"
	"speedtest/config"
	"speedtest/results"
	"strings"

	"github.com/gin-gonic/gin"
)

// 预编译的正则表达式变量
var (
	localIPv6Regex          = regexp.MustCompile(`^::1$`)                            // 匹配本地 IPv6 地址
	linkLocalIPv6Regex      = regexp.MustCompile(`^fe80:`)                           // 匹配链路本地 IPv6 地址
	localIPv4Regex          = regexp.MustCompile(`^127\.`)                           // 匹配本地 IPv4 地址
	privateIPv4Regex10      = regexp.MustCompile(`^10\.`)                            // 匹配私有 IPv4 地址（10.0.0.0/8）
	privateIPv4Regex172     = regexp.MustCompile(`^172\.(1[6-9]|2\d|3[01])\.`)       // 匹配私有 IPv4 地址（172.16.0.0/12）
	privateIPv4Regex192     = regexp.MustCompile(`^192\.168\.`)                      // 匹配私有 IPv4 地址（192.168.0.0/16）
	linkLocalIPv4Regex      = regexp.MustCompile(`^169\.254\.`)                      // 匹配链路本地 IPv4 地址（169.254.0.0/16）
	cgnatIPv4Regex          = regexp.MustCompile(`^100\.([6-9][0-9]|1[0-2][0-7])\.`) // 匹配 CGNAT IPv4 地址（100.64.0.0/10）
	unspecifiedAddressRegex = regexp.MustCompile(`^0\.0\.0\.0$`)                     // 匹配未指定地址（0.0.0.0）
	broadcastAddressRegex   = regexp.MustCompile(`^255\.255\.255\.255$`)             // 匹配广播地址（255.255.255.255）
	removeASRegexp          = regexp.MustCompile(`AS\d+\s`)                          // 用于去除 ISP 信息中的自治系统编号
)

// getIP 处理对/getIP的请求，返回客户端IP地址及其相关信息
func getIP(c *gin.Context, cfg *config.Config) {
	var ret results.Result // 创建结果结构体实例

	clientIP := c.ClientIP() // 获取客户端 IP 地址

	// 使用正则表达式匹配不同类型的 IP 地址
	switch {
	case localIPv6Regex.MatchString(clientIP):
		ret.ProcessedString = clientIP + " - localhost IPv6 access" // 本地 IPv6 地址
	case linkLocalIPv6Regex.MatchString(clientIP):
		ret.ProcessedString = clientIP + " - link-local IPv6 access" // 链路本地 IPv6 地址
	case localIPv4Regex.MatchString(clientIP):
		ret.ProcessedString = clientIP + " - localhost IPv4 access" // 本地 IPv4 地址
	case privateIPv4Regex10.MatchString(clientIP):
		ret.ProcessedString = clientIP + " - private IPv4 access" // 私有 IPv4 地址（10.0.0.0/8）
	case privateIPv4Regex172.MatchString(clientIP):
		ret.ProcessedString = clientIP + " - private IPv4 access" // 私有 IPv4 地址（172.16.0.0/12）
	case privateIPv4Regex192.MatchString(clientIP):
		ret.ProcessedString = clientIP + " - private IPv4 access" // 私有 IPv4 地址（192.168.0.0/16）
	case linkLocalIPv4Regex.MatchString(clientIP):
		ret.ProcessedString = clientIP + " - link-local IPv4 access" // 链路本地 IPv4 地址
	case cgnatIPv4Regex.MatchString(clientIP):
		ret.ProcessedString = clientIP + " - CGNAT IPv4 access" // CGNAT IPv4 地址（100.64.0.0/10）
	case unspecifiedAddressRegex.MatchString(clientIP):
		ret.ProcessedString = clientIP + " - unspecified address" // 未指定地址（0.0.0.0）
	case broadcastAddressRegex.MatchString(clientIP):
		ret.ProcessedString = clientIP + " - broadcast address" // 广播地址（255.255.255.255）
	default:
		ret.ProcessedString = clientIP // 其他情况，返回原始 IP 地址
	}

	// 检查处理结果中是否包含特定信息
	if strings.Contains(ret.ProcessedString, " - ") {
		// 将 ret 转换为 JSON 字符串
		jsonData, err := json.Marshal(ret)
		if err != nil {
			// 如果转换失败，记录错误信息
			logInfo("Error marshaling JSON: " + err.Error())
		} else {
			// 如果转换成功，记录 JSON 字符串
			logInfo(string(jsonData))
		}
		c.JSON(200, ret) // 返回 JSON 响应
		return
	}

	// 检查是否需要获取 ISP 信息
	getISPInfo := c.Query("isp") == "true"

	if getISPInfo {
		// 调用函数获取 ISP 信息
		ispInfo := getIPInfo(clientIP)

		ret.RawISPInfo = ispInfo                                // 存储原始 ISP 信息
		isp := removeASRegexp.ReplaceAllString(ispInfo.ISP, "") // 去除 ISP 信息中的自治系统编号

		if isp == "" {
			isp = "Unknown ISP" // 如果 ISP 信息为空，设置为未知
		}

		if ispInfo.CountryName != "" {
			isp += ", " + ispInfo.CountryName // 如果有国家名称，添加到 ISP 信息中
		}

		ret.ProcessedString += " - " + isp // 更新处理后的字符串
	}

	// 将 ret 转换为 JSON 字符串
	jsonData, err := json.Marshal(ret)
	if err != nil {
		// 如果转换失败，记录错误信息
		logInfo("Error marshaling JSON: " + err.Error())
	} else {
		// 如果转换成功，记录 JSON 字符串
		logInfo(string(jsonData))
	}

	c.JSON(200, ret) // 返回 JSON 响应
}
