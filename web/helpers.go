package web

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/WJQSERVER-STUDIO/go-utils/logger"
	"github.com/umahmood/haversine"

	"speedtest/results"
)

var (
	serverCoord haversine.Coord
)

var (
	logw       = logger.Logw
	logInfo    = logger.LogInfo
	logWarning = logger.LogWarning
	logError   = logger.LogError
)

func getRandomData(length int) []byte {
	data := make([]byte, length)
	if _, err := rand.Read(data); err != nil {
		logError("Failed to generate random data: %s", err)
	}
	return data
}

/*
func getIPInfoURL(address string, apiKey string) string {
	//apiKey := config.LoadedConfig().IPInfoAPIKey

	ipInfoURL := `https://ipinfo.io/%s/json`
	if address != "" {
		ipInfoURL = fmt.Sprintf(ipInfoURL, address)
	} else {
		ipInfoURL = "https://ipinfo.io/json"
	}

	if apiKey != "" {
		ipInfoURL += "?token=" + apiKey
	}

	return ipInfoURL
} */

func getIPInfoURL(address string) string {
	ipInfoUrl := "https://ip.1888866.xyz/api/ip-lookup"
	if address != "" {
		ipInfoUrl = fmt.Sprintf(ipInfoUrl+"?ip=%s", address)
	} else {
		logError("No IP address provided for lookup")
		return ""
	}
	return ipInfoUrl
}

func getIPInfo(addr string) results.IPInfoResponse {
	var ret results.IPInfoResponse
	// 预检addr是否为空， 为空则返回空结果
	if addr == "" {
		logError("No IP address provided for lookup")
		return ret
	}
	resp, err := http.DefaultClient.Get(getIPInfoURL(addr))
	if err != nil {
		logError("Error getting response from ipinfo.io: %s", err)
		return ret
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		logError("Error reading response from ipinfo.io: %s", err)
		return ret
	}

	if err := json.Unmarshal(raw, &ret); err != nil {
		logError("Error parsing response from ipinfo.io: %s", err)
	}
	logInfo("Got IP info: %s", ret)
	return ret
}
