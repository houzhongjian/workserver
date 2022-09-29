package serverd

import (
	"crypto/tls"
	"github.com/robfig/cron/v3"
	"io/ioutil"
	"log"
	"time"
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
	server.certMap.Lock()
	defer server.certMap.Unlock()
	for _, sn := range server.GetAllServerName() {
		//检测当前证书是否过期.
		host := server.certMap.Data[sn]
		expireTime, err := GetCertExpireTimeToByte(host.KeyData)
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

			//获取证书.
			KeyBuf, err := ioutil.ReadFile(host.KeyFile)
			if err != nil {
				log.Printf("err:%+v\n", err)
				return
			}

			pemBuf, err := ioutil.ReadFile(host.PemFile)
			if err != nil {
				log.Printf("err:%+v\n", err)
				return
			}

			host.PemData = pemBuf
			host.KeyData = KeyBuf
		}
	}
}

func (server *Serverd) returnCert(helloInfo *tls.ClientHelloInfo) (*tls.Certificate, error) {
	server.certMap.RLock()
	defer server.certMap.RUnlock()

	host, ok := server.certMap.Data[helloInfo.ServerName]
	if !ok {
		return nil, nil
	}

	cer, err := tls.X509KeyPair(host.PemData, host.KeyData)
	if err != nil {
		log.Println(err)
		return nil, nil
	}
	return &cer, nil
}