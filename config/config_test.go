package config

import (
	"gopkg.in/yaml.v2"
	"log"
	"testing"
)

type Conf struct {
	Type       string   `yaml:"type"`
	ServerName string   `yaml:"server_name"`
	ProxyPass  string   `yaml:"proxy_pass"`
	Module     []Module `yaml:"module"`
}

type Module struct {
	Path      string `yaml:"path"`
	ProxyPass string `yaml:"proxy_pass"`
}

var yamlStr = `type: reverse_proxy
proxy_pass: https://www.moonseer.com
server_name: ws2.houzhongjian.com
module: 
  - path: /cn/
    proxy_pass: https://ws.mengyuzhe.com

  - path: /jiancheng/
    proxy_pass: http://www.jianchengjiaoyu.com`

func TestLoad(t *testing.T) {
	obj := Conf{}

	if err := yaml.Unmarshal([]byte(yamlStr), &obj); err != nil {
		log.Println("111")
		log.Println(err)
		return
	}
	log.Println("222")
	log.Printf("result:%+v\n", obj)
}
