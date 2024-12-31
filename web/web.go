package web

import (
	"embed"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/pires/go-proxyproto"
	log "github.com/sirupsen/logrus"

	"speedtest/config"
	"speedtest/results"
)

const (
	// chunk size is 1 MiB
	chunkSize = 1048576
)

//go:embed assets
var defaultAssets embed.FS

var (
	// generate random data for download test on start to minimize runtime overhead
	randomData = getRandomData(chunkSize)
)

// ListenAndServe 启动HTTP服务器并设置路由处理程序
func ListenAndServe(conf *config.Config) error {
	gin.SetMode(gin.DebugMode)
	r := gin.Default()
	r.UseH2C = true

	// CORS
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "OPTIONS", "HEAD"},
		AllowHeaders:    []string{"*"},
	}))

	var assetFS http.FileSystem
	if fi, err := os.Stat(conf.AssetsPath); err != nil || !fi.IsDir() {
		log.Warnf("Configured asset path %s does not exist or is not a directory, using default assets", conf.AssetsPath)
		sub, err := fs.Sub(defaultAssets, "assets")
		if err != nil {
			log.Fatalf("Failed when processing default assets: %s", err)
		}
		assetFS = http.FS(sub)
	} else {
		assetFS = justFilesFilesystem{fs: http.Dir(conf.AssetsPath)}
	}

	r.StaticFS(conf.BaseURL, assetFS)

	r.POST(conf.BaseURL+"/results/telemetry", results.Record)
	r.GET(conf.BaseURL+"/results", results.DrawPNG)
	r.GET(conf.BaseURL+"/getIP", getIP)
	r.GET(conf.BaseURL+"/garbage", garbage)
	r.GET(conf.BaseURL+"/empty", empty)

	// PHP frontend default values compatibility
	r.GET(conf.BaseURL+"/empty.php", empty)
	r.GET(conf.BaseURL+"/garbage.php", garbage)
	r.GET(conf.BaseURL+"/getIP.php", getIP)
	r.POST(conf.BaseURL+"/results/telemetry.php", results.Record)

	go listenProxyProtocol(conf, r)

	return startListener(conf, r)
}

// listenProxyProtocol 启动一个监听Proxy Protocol的HTTP服务器
func listenProxyProtocol(conf *config.Config, r *gin.Engine) {
	if conf.ProxyProtocolPort != "0" {
		addr := net.JoinHostPort(conf.BindAddress, conf.ProxyProtocolPort)
		l, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatalf("Cannot listen on proxy protocol port %s: %s", conf.ProxyProtocolPort, err)
		}

		pl := &proxyproto.Listener{Listener: l}
		defer pl.Close()

		log.Infof("Starting proxy protocol listener on %s", addr)
		log.Fatal(http.Serve(pl, r))
	}
}

// empty 处理对/empty的请求，丢弃请求体并返回成功的状态码
func empty(c *gin.Context) {
	_, err := io.Copy(io.Discard, c.Request.Body)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	c.Status(http.StatusOK)
}

// garbage 处理对/garbage的请求，返回指定数量的随机数据块
func garbage(c *gin.Context) {
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename=random.dat")
	c.Header("Content-Transfer-Encoding", "binary")

	// chunk size set to 4 by default
	chunks := 4

	ckSize := c.Query("ckSize")
	if ckSize != "" {
		i, err := strconv.ParseInt(ckSize, 10, 64)
		if err != nil {
			log.Errorf("Invalid chunk size: %s", ckSize)
			log.Warnf("Will use default value %d", chunks)
		} else {
			// limit max chunk size to 1024
			if i > 1024 {
				chunks = 1024
			} else {
				chunks = int(i)
			}
		}
	}

	for i := 0; i < chunks; i++ {
		if _, err := c.Writer.Write(randomData); err != nil {
			log.Errorf("Error writing back to client at chunk number %d: %s", i, err)
			break
		}
	}
}

// getIP 处理对/getIP的请求，返回客户端IP地址及其相关信息
func getIP(c *gin.Context) {
	var ret results.Result

	clientIP := c.ClientIP()

	isSpecialIP := true
	switch {
	case clientIP == "::1":
		ret.ProcessedString = clientIP + " - localhost IPv6 access"
	case strings.HasPrefix(clientIP, "fe80:"):
		ret.ProcessedString = clientIP + " - link-local IPv6 access"
	case strings.HasPrefix(clientIP, "127."):
		ret.ProcessedString = clientIP + " - localhost IPv4 access"
	case strings.HasPrefix(clientIP, "10."):
		ret.ProcessedString = clientIP + " - private IPv4 access"
	case regexp.MustCompile(`^172\.(1[6-9]|2\d|3[01])\.`).MatchString(clientIP):
		ret.ProcessedString = clientIP + " - private IPv4 access"
	case strings.HasPrefix(clientIP, "192.168"):
		ret.ProcessedString = clientIP + " - private IPv4 access"
	case strings.HasPrefix(clientIP, "169.254"):
		ret.ProcessedString = clientIP + " - link-local IPv4 access"
	case regexp.MustCompile(`^100\.([6-9][0-9]|1[0-2][0-7])\.`).MatchString(clientIP):
		ret.ProcessedString = clientIP + " - CGNAT IPv4 access"
	default:
		isSpecialIP = false
	}

	if isSpecialIP {
		c.JSON(200, ret)
		return
	}

	getISPInfo := c.Query("isp") == "true"
	distanceUnit := c.Query("distance")

	ret.ProcessedString = clientIP

	if getISPInfo {
		ispInfo := getIPInfo(clientIP)
		ret.RawISPInfo = ispInfo

		removeRegexp := regexp.MustCompile(`AS\d+\s`)
		isp := removeRegexp.ReplaceAllString(ispInfo.Organization, "")

		if isp == "" {
			isp = "Unknown ISP"
		}

		if ispInfo.Country != "" {
			isp += ", " + ispInfo.Country
		}

		if ispInfo.Location != "" {
			isp += " (" + calculateDistance(ispInfo.Location, distanceUnit) + ")"
		}

		ret.ProcessedString += " - " + isp
	}

	c.JSON(200, ret)
}
