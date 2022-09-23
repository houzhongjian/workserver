package serverd

import (
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net/http"
	"sync"
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
	Port             int
	AccessLog        string
	ErrorLog         string
	Email            string
	CertsDir         string
	FileServer       []config.FileServerConfig
	ReverseProxy     []config.ReverseProxyConfig
	MP               map[string]ServerParams
	CertManager      autocert.Manager
	keyAuthorization KeyAuthorization
}

type KeyAuthorization struct {
	Data map[string]string
	sync.RWMutex
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
		keyAuthorization: KeyAuthorization{
			Data:    make(map[string]string),
			RWMutex: sync.RWMutex{},
		},
	}
}

func (server *Serverd) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("---> request: ", r.Host, r.Method, r.RequestURI)

	serverParams, ok := server.MP[r.Host]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	serverParams.Run(server, w, r)
}

func (server *Serverd) Run() {
	server.initMap()
	go server.CheckCert()

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
