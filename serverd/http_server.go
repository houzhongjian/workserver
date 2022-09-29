package serverd

import (
	"fmt"
	"golang.org/x/crypto/acme/autocert"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sync"
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
	certMap          CertMap
}

type CertMap struct {
	sync.RWMutex
	Data map[string]CertMapData
}

type CertMapData struct {
	PemData []byte
	KeyData []byte
	KeyFile string
	PemFile string
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
		certMap: CertMap{
			RWMutex: sync.RWMutex{},
			Data:    make(map[string]CertMapData),
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
	server.initCertToMap()
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

//initCertToMap 初始化证书到map.
func (server *Serverd) initCertToMap() {
	//获取所有的证书.
	server.certMap.Lock()
	defer server.certMap.RUnlock()

	for _, sn := range server.GetAllServerName() {
		host, ok := server.certMap.Data[sn]
		if !ok {
			server.certMap.Data[sn] = CertMapData{
				PemData: nil,
				KeyData: nil,
				KeyFile: "",
				PemFile: "",
			}
		}

		//检测当前sn是否存在证书.
		keyFile := path.Clean(fmt.Sprintf("%s/%s.key", server.CertsDir, sn))
		pemFile := path.Clean(fmt.Sprintf("%s/%s.pem", server.CertsDir, sn))

		if _, err := os.Stat(keyFile); os.IsNotExist(err) {
			//不存在证书:签发证书.
			if err := server.IssueCertificate(sn); err != nil {
				log.Printf("err:%+v\n", err)
				return
			}
		}

		//检测当前证书是否过期.
		expireTime, err := GetCertExpireTimeToFile(keyFile)
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

		keyBuf, err := ioutil.ReadFile(keyFile)
		if err != nil {
			log.Printf("err:%+v\n", err)
			return
		}


		pemBuf, err := ioutil.ReadFile(pemFile)
		if err != nil {
			log.Printf("err:%+v\n", err)
			return
		}

		//将证书加载到map中.
		host.KeyFile = keyFile
		host.KeyData = keyBuf
		host.PemFile = pemFile
		host.PemData = pemBuf
	}
}
