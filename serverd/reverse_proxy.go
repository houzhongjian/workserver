package serverd

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"workserver/config"
)

//StartReverseProxy 开始转发代理.
func StartReverseProxy(reverseProxy config.ReverseProxyConfig, w http.ResponseWriter,r *http.Request) {
	log.Println("StartReverseProxy")
	//判断是否强制跳转到https.
	if reverseProxy.ForceJumpHttps && r.TLS == nil {
		host := fmt.Sprintf("https://%s%s",r.Host,r.RequestURI)
		http.Redirect(w,r,host, http.StatusMovedPermanently)
		return
	}

	//判断是否配置了路由转发模块.
	log.Printf("StartReverseProxy Module:%+v\n",reverseProxy.Module)
	log.Println("requestURI 1:", r.RequestURI)
	if len(reverseProxy.Module) > 0 {
		for _, mod := range reverseProxy.Module {
			requestURI := strings.TrimPrefix(r.RequestURI, mod.Path)
			res := strings.HasPrefix(r.RequestURI, mod.Path)
			log.Println("requestURI 2:",requestURI)
			if res{
				u, _ := url.Parse(requestURI)
				startReverseProxy(mod.ProxyPass, w, r, u)
				return
			}
		}
	}

	u, _ := url.Parse(r.RequestURI)
	startReverseProxy(reverseProxy.ProxyPass, w,r, u)
}

func startReverseProxy(proxyPass string ,w http.ResponseWriter, r *http.Request,u *url.URL) {
	target, err := url.Parse(proxyPass)
	if err != nil {
		log.Printf("err:%+v\n", err)
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	r.URL.Host = target.Host
	r.URL.Scheme = target.Scheme
	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	r.Host = target.Host
	r.URL = u

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