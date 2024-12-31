package web

import (
	"fmt"
	"speedtest/config"

	"github.com/gin-gonic/gin"
)

func GinRoute(conf *config.Config, r *gin.Engine) error {
	var (
		addr string
		port int
	)
	if conf.BindAddress == "" {
		addr = "0.0.0.0"
	} else {
		addr = conf.BindAddress
	}
	if conf.Port == 0 {
		port = 8989
	} else {
		port = conf.Port
	}
	return r.Run(fmt.Sprintf("%s:%d", addr, port))
}
