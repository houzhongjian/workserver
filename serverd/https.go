package serverd

import (
	"crypto/tls"
	"fmt"
	"golang.org/x/net/http2"
	"log"
	"net/http"
	"time"
)

func (server *Serverd) runHTTPSWorkServer() {
	log.Println("runHTTPSWorkServer")
	t := &tls.Config{
		GetCertificate: server.returnCert,
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