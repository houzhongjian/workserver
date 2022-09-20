package serverd

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"workserver/config"
)

func ServeFile(fileServer config.FileServerConfig, w http.ResponseWriter,r *http.Request) {
	file := path.Clean(fmt.Sprintf("%s/%s", fileServer.Root, r.RequestURI))
	log.Printf("file:%+v\n", file)
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