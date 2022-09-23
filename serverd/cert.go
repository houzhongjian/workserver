package serverd

import (
	"crypto/tls"
	"fmt"
	"github.com/robfig/cron/v3"
	"log"
	"path"
	"time"
	"workserver/config"
)
func (server *Serverd) CheckCert() {
	c := cron.New()
	_, err := c.AddFunc("0 0 1 * *", server.checkCert)
	if err != nil{
		log.Printf("err:%+v\n", err)
		return
	}
}

func (server *Serverd) checkCert() {
	log.Println("checkCert")
	for _, sn := range server.GetAllServerName() {
		//检测当前证书是否过期.
		keyFile := path.Clean(fmt.Sprintf("%s/%s.key", server.CertsDir, sn))
		expireTime, err := GetCertExpireTime(keyFile)
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
	}
}

func returnCert(helloInfo *tls.ClientHelloInfo) (*tls.Certificate, error) {
	publicKey := path.Clean(fmt.Sprintf("./%s/%s.key", config.ConfigMgr.CertsDir, helloInfo.ServerName))
	privatePem := path.Clean(fmt.Sprintf("./%s/%s.pem", config.ConfigMgr.CertsDir, helloInfo.ServerName))

	cer, err := tls.LoadX509KeyPair(publicKey, privatePem)
	if err != nil {
		log.Println(err)
		return nil, nil
	}
	return &cer, nil
}