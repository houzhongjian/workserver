package serverd

import (
	"workserver/config"
)

func Run() {
	cf := config.ConfigMgr
	server := NewServerd(ServerdOptions{
		Port:         cf.Port,
		AccessLog:    cf.AccessLog,
		ErrorLog:     cf.AccessLog,
		FileServer:   cf.FileServer,
		ReverseProxy: cf.ReverseProxy,
		Email:        cf.Email,
		CertsDir:     cf.CertsDir,
	})
	server.Run()
}
