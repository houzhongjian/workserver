package serverd

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"strings"
	"time"

	"crypto/x509/pkix"

	"github.com/eggsampler/acme/v3"
)

type acmeAccountFile struct {
	PrivateKey string `json:"privateKey"`
	Url        string `json:"url"`
}

//IssueCertificate 签发证书.
func (server *Serverd) IssueCertificate(domain string) error {
	client, err := acme.NewClient(acme.LetsEncryptProduction)
	if err != nil {
		log.Printf("Error connecting to acme directory: %+v", err)
		return err
	}

	accountFile := path.Clean(fmt.Sprintf("%s/account.json", server.CertsDir))
	log.Println("loadAccount")
	account, err := loadAccount(accountFile, server.Email, client)
	if err != nil {
		account, err = createAccount(accountFile, server.Email, client)
		if err != nil {
			log.Printf("Error creaing new account: %+v", err)
			return err
		}
	}

	var ids = []acme.Identifier{
		{
			Type:  "dns",
			Value: domain,
		},
	}

	log.Println("NewOrder")
	order, err := client.NewOrder(account, ids)
	if err != nil {
		log.Printf("Error creating new order: %+v", err)
		return err
	}

	//server.keyAuthorization.Lock()
	//defer server.keyAuthorization.Unlock()
	for _, authUrl := range order.Authorizations {
		log.Println("FetchAuthorization")
		auth, err := client.FetchAuthorization(account, authUrl)
		if err != nil {
			log.Printf("Error fetching authorization url %+v\n", err)
			return err
		}

		chal, ok := auth.ChallengeMap[acme.ChallengeTypeHTTP01]
		if !ok {
			log.Printf("Unable to find http challenge for auth %+v", auth.Identifier.Value)
			return err
		}

		tokenFile := path.Clean(fmt.Sprintf("/.well-known/acme-challenge/%s", chal.Token))
		server.keyAuthorization.Data[tokenFile] = chal.KeyAuthorization
		log.Printf("server.keyAuthorization.Data[tokenFile]:%+v\n",server.keyAuthorization.Data)

		log.Printf("Updating challenge for authorization %s: %s", auth.Identifier.Value, chal.URL)
		chal, err = client.UpdateChallenge(account, chal)
		if err != nil {
			log.Printf("Error updating authorization %s challenge: %+v", auth.Identifier.Value, err)
			return err
		}
	}

	log.Println("GenerateKey")
	certKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Printf("Error generating certificate key: %+v", err)
		return err
	}

	b := key2pem(certKey)

	privatePem := path.Clean(fmt.Sprintf("%s/%s.pem", server.CertsDir, domain))
	if err := ioutil.WriteFile(privatePem, b, 0600); err != nil {
		log.Printf("Error writing key file %+v", err)
		return err
	}

	tpl := &x509.CertificateRequest{
		SignatureAlgorithm: x509.ECDSAWithSHA256,
		PublicKeyAlgorithm: x509.ECDSA,
		PublicKey:          certKey.Public(),
		Subject:            pkix.Name{CommonName: domain},
		DNSNames:           []string{domain},
	}
	log.Println("CreateCertificateRequest")
	csrDer, err := x509.CreateCertificateRequest(rand.Reader, tpl, certKey)
	if err != nil {
		log.Printf("Error creating certificate request: %+v", err)
		return err
	}
	csr, err := x509.ParseCertificateRequest(csrDer)
	if err != nil {
		log.Printf("Error parsing certificate request: %+v", err)
		return err
	}

	log.Println("FinalizeOrder")
	order, err = client.FinalizeOrder(account, order, csr)
	if err != nil {
		log.Printf("Error finalizing order: %+v", err)
		return err
	}

	log.Println("FinalizeOrder")
	certs, err := client.FetchCertificates(account, order.Certificate)
	if err != nil {
		log.Printf("Error fetching order certificates: %+v", err)
		return err
	}

	var pemData []string
	for _, c := range certs {
		pemData = append(pemData, strings.TrimSpace(string(pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: c.Raw,
		}))))
	}
	publicKey := path.Clean(fmt.Sprintf("%s/%s.key", server.CertsDir, domain))
	if err := ioutil.WriteFile(publicKey, []byte(strings.Join(pemData, "\n")), 0600); err != nil {
		log.Printf("Error writing certificate file %+v", err)
		return err
	}
	log.Println("Done.")
	return nil
}

func loadAccount(accountFile, email string, client acme.Client) (acme.Account, error) {
	raw, err := ioutil.ReadFile(accountFile)
	if err != nil {
		return acme.Account{}, fmt.Errorf("error reading account file %q: %v", accountFile, err)
	}
	var aaf acmeAccountFile
	if err := json.Unmarshal(raw, &aaf); err != nil {
		return acme.Account{}, fmt.Errorf("error parsing account file %q: %v", accountFile, err)
	}
	account, err := client.UpdateAccount(acme.Account{PrivateKey: pem2key([]byte(aaf.PrivateKey)), URL: aaf.Url}, getContacts(email)...)
	if err != nil {
		return acme.Account{}, fmt.Errorf("error updating existing account: %v", err)
	}
	return account, nil
}

func createAccount(accountFile, email string, client acme.Client) (acme.Account, error) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return acme.Account{}, fmt.Errorf("error creating private key: %v", err)
	}
	account, err := client.NewAccount(privKey, false, true, getContacts(email)...)
	if err != nil {
		return acme.Account{}, fmt.Errorf("error creating new account: %v", err)
	}
	raw, err := json.Marshal(acmeAccountFile{PrivateKey: string(key2pem(privKey)), Url: account.URL})
	if err != nil {
		return acme.Account{}, fmt.Errorf("error parsing new account: %v", err)
	}
	if err := ioutil.WriteFile(accountFile, raw, 0600); err != nil {
		return acme.Account{}, fmt.Errorf("error creating account file: %v", err)
	}
	return account, nil
}

func getContacts(email string) []string {
	var contacts = []string{
		"mailto:" + email,
	}
	return contacts
}

func key2pem(certKey *ecdsa.PrivateKey) []byte {
	certKeyEnc, err := x509.MarshalECPrivateKey(certKey)
	if err != nil {
		log.Fatalf("Error encoding key: %v", err)
	}

	return pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: certKeyEnc,
	})
}

func pem2key(data []byte) *ecdsa.PrivateKey {
	b, _ := pem.Decode(data)
	key, err := x509.ParseECPrivateKey(b.Bytes)
	if err != nil {
		log.Fatalf("Error decoding key: %v", err)
	}
	return key
}

//GetCertExpireTime 获取证书的过期时间.
func GetCertExpireTime(keyPath string) (t time.Time, err error) {
	certPEMBlock, err := ioutil.ReadFile(keyPath)
	if err != nil {
		log.Println(err)
		return t, err
	}

	certDERBlock, _ := pem.Decode(certPEMBlock)
	if certDERBlock == nil {
		log.Println(err)
		return t,err
	}
	x509Cert, err := x509.ParseCertificate(certDERBlock.Bytes)
	if err != nil {
		log.Println(err)
		return t,err
	}

	return x509Cert.NotAfter.In(time.Local), nil
}