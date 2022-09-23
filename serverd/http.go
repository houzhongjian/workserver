package serverd

import (
	"log"
	"net/http"
)

func (server *Serverd) runHTTPWorkServer() {
	log.Println("runHTTPWorkServer")
	http.HandleFunc("/.well-known/acme-challenge/", func(w http.ResponseWriter, r *http.Request) {
		server.keyAuthorization.RLock()
		defer server.keyAuthorization.RUnlock()
		keyAuthorization, ok := server.keyAuthorization.Data[r.RequestURI]
		if !ok {
			log.Println("keyAuthorization=获取失败")
			return
		}
		log.Println("keyAuthorization=", keyAuthorization)
		w.Write([]byte(keyAuthorization))
	})
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		panic(err)
	}
}