package serverd

import (
	"crypto/tls"
	"fmt"
	"golang.org/x/net/http2"
	"log"
	"net/http"
	"os"
	"path"
	"time"
)

func (server *Serverd) runHTTPSWorkServer() {
	for _, sn := range server.GetAllServerName() {
		//检测当前sn是否存在证书.
		keyFile := path.Clean(fmt.Sprintf("%s/%s.key", server.CertsDir, sn))
		if _, err := os.Stat(keyFile); os.IsNotExist(err) {
			//不存在证书:签发证书.
			if err := server.IssueCertificate(sn); err != nil {
				log.Printf("err:%+v\n", err)
				return
			}
			continue
		}

		//检测当前证书是否过期.
		expireTime, err := GetCertExpireTime(keyFile)
		if err != nil {
			log.Printf("err:%+v\n", err)
			return
		}
		if (expireTime.Sub(time.Now()).Hours() / 24) < 30 {
			//证书即将过期:签发证书.
			if err := server.IssueCertificate(sn); err != nil {
				log.Printf("err:%+v\n", err)
				return
			}
		}
	}

	log.Println("runHTTPSWorkServer")
	t := &tls.Config{
		GetCertificate: returnCert,
		NextProtos:     []string{http2.NextProtoTLS, "http/1.1"},
		MinVersion:     tls.VersionTLS12,
	}

	addr := fmt.Sprintf(":%d", server.Port)
	log.Println("runHTTPSWorkServer")
	s := &http.Server{
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      server,
	}
	s.Addr = addr
	s.TLSConfig = t
	panic(s.ListenAndServeTLS("", ""))
}