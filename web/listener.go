//go:build !linux
// +build !linux

package web

import (
	"crypto/tls"
	"net"
	"net/http"

	"speedtest/config"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func startListener(conf *config.Config, r *gin.Engine) error {
	var s error

	addr := net.JoinHostPort(conf.BindAddress, conf.Port)
	log.Infof("Starting backend server on %s", addr)

	// TLS and HTTP/2.
	if conf.EnableTLS {
		log.Info("Use TLS connection.")
		if !conf.EnableHTTP2 {
			srv := &http.Server{
				Addr:         addr,
				Handler:      r,
				TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
			}
			s = srv.ListenAndServeTLS(conf.TLSCertFile, conf.TLSKeyFile)
		} else {
			s = http.ListenAndServeTLS(addr, conf.TLSCertFile, conf.TLSKeyFile, r)
		}
	} else {
		if conf.EnableHTTP2 {
			log.Errorf("TLS is mandatory for HTTP/2. Ignore settings that enable HTTP/2.")
		}
		s = http.ListenAndServe(addr, r)
	}

	return s
}
