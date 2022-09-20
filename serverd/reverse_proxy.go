package serverd

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"workserver/config"
)

func StartReverseProxy(reverseProxy config.ReverseProxyConfig, w http.ResponseWriter,r *http.Request) {
	target, err := url.Parse(reverseProxy.ProxyPass)
	if err != nil {
		log.Printf("err:%+v\n", err)
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	r.URL.Host = target.Host
	r.URL.Scheme = target.Scheme
	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	r.Host = target.Host

	proxy.ServeHTTP(w,r)
}

func getServerNameByReverseProxy(name string, reverseProxy []config.ReverseProxyConfig) (server config.ReverseProxyConfig, ok bool) {
	for _, r := range reverseProxy {
		if r.ServerName == name {
			return r,true
		}
	}
	return server, false
}