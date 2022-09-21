package serverd

import (
	"crypto/tls"
	"fmt"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/net/http2"
	"log"
	"net/http"
	"time"
	"workserver/config"
)

type ServerdOptions struct {
	Port         int
	AccessLog    string
	ErrorLog     string
	Email        string
	CertsDir     string
	FileServer   []config.FileServerConfig
	ReverseProxy []config.ReverseProxyConfig
}
type Serverd struct {
	Port         int
	AccessLog    string
	ErrorLog     string
	Email        string
	CertsDir     string
	FileServer   []config.FileServerConfig
	ReverseProxy []config.ReverseProxyConfig
	MP           map[string]ServerParams
	CertManager  autocert.Manager
}

type ServerParams struct {
	Type       config.ConfigType
	ServerName string
}

func (params ServerParams) Run(server *Serverd, w http.ResponseWriter, r *http.Request) {
	switch params.Type {
	case config.ConfigType_FileServer:
		fs, _ := getServerNameByFileServer(params.ServerName, server.FileServer)
		ServeFile(fs, w, r)
	case config.ConfigType_ReverseProxy:
		reverse, _ := getServerNameByReverseProxy(params.ServerName, server.ReverseProxy)
		StartReverseProxy(reverse, w, r)
	}
}

func NewServerd(opt ServerdOptions) *Serverd {
	return &Serverd{
		Port:         opt.Port,
		AccessLog:    opt.AccessLog,
		ErrorLog:     opt.ErrorLog,
		FileServer:   opt.FileServer,
		ReverseProxy: opt.ReverseProxy,
		MP:           make(map[string]ServerParams),
		Email:        opt.Email,
		CertsDir:     opt.CertsDir,
	}
}


func (server *Serverd) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Host, r.Method, r.RequestURI)

	serverParams, ok := server.MP[r.Host]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	serverParams.Run(server, w, r)
}

func (server *Serverd) Run() {
	server.initMap()
	server.runCerts()

	go server.runHTTPSWorkServer()
	server.runHTTPWorkServer()
}

//GetAllServerName 获取所有的server_name.
func (server *Serverd) GetAllServerName() []string {
	serverName := []string{}

	for _, fileServer := range server.FileServer {
		serverName = append(serverName, fileServer.ServerName)
	}

	for _, reverseProxy := range server.ReverseProxy {
		serverName = append(serverName, reverseProxy.ServerName)
	}

	return serverName
}
func (server *Serverd) runCerts() {
	server.CertManager = autocert.Manager{
		Prompt:      autocert.AcceptTOS,
		HostPolicy:  autocert.HostWhitelist(server.GetAllServerName()...), //your domain here
		Cache:       autocert.DirCache(server.CertsDir),                   //folder for storing certificates
		Email:       server.Email,
	}
}

func (server *Serverd) runHTTPSWorkServer() {
	t := &tls.Config{
		GetCertificate: server.CertManager.GetCertificate,
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
	panic(s.ListenAndServeTLS("",""))
}

func (server *Serverd) runHTTPWorkServer() {
	log.Println("runHTTPWorkServer")
	staticHandler := http.StripPrefix("/.well-known/acme-challenge/", http.FileServer(http.Dir(server.CertsDir)))
	http.Handle("/.well-known/acme-challenge/", staticHandler)
	err := http.ListenAndServe(":80", server.CertManager.HTTPHandler(server))
	if err != nil {
		panic(err)
	}
}

func (server *Serverd) initMap() {
	for _, fileServer := range server.FileServer {
		server.MP[fileServer.ServerName] = ServerParams{
			Type:       fileServer.Type,
			ServerName: fileServer.ServerName,
		}
	}

	for _, reverseProxy := range server.ReverseProxy {
		server.MP[reverseProxy.ServerName] = ServerParams{
			Type:       reverseProxy.Type,
			ServerName: reverseProxy.ServerName,
		}
	}
}
