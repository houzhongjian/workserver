package serverd

import (
	"fmt"
	"net/http"
	"path"
	"workserver/config"
)

func ServeFile(fileServer config.FileServerConfig, w http.ResponseWriter,r *http.Request) {
	//判断是否强制跳转到https.
	if fileServer.ForceJumpHttps && r.TLS == nil {
		host := fmt.Sprintf("https://%s%s",r.Host,r.RequestURI)
		http.Redirect(w,r,host, http.StatusMovedPermanently)
		return
	}

	file := path.Clean(fmt.Sprintf("%s/%s", fileServer.Root, r.RequestURI))
	http.ServeFile(w,r,file)
}

func getServerNameByFileServer(name string, fileServer []config.FileServerConfig) (server config.FileServerConfig, ok bool) {
	for _, f := range fileServer {
		if f.ServerName == name {
			return f,true
		}
	}
	return server, false
}