package acme

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/go-acme/lego/v4/certificate"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type CertInfo struct {
	Parent    *CertInfo `json:"parent"`
	Domains   []string  `json:"domains"`
	NotBefore time.Time `json:"notBefore"`
	NotAfter  time.Time `json:"notAfter"`
	Issuer    string    `json:"issuer"`
	Subject   string    `json:"subject"`
	IsCA      bool      `json:"isCA"`
	File      string    `json:"file"`
	Path      string    `json:"path"`
}

func (info *CertInfo) String() string {
	p := ""
	if info.Parent != nil {
		p = fmt.Sprintf("\n上级机构:\n%s", info.Parent)
	}
	domains := ""
	if len(info.Domains) > 0 {
		domains = fmt.Sprintf("domains:%s\n", info.Domains)
	}
	return fmt.Sprintf("%sissuer:[%s]\nsubject:[%s]\nisCA:%v\n有效期:%v 至 %v %s",
		domains, info.Issuer, info.Subject, info.IsCA, info.NotBefore.Format(time.DateTime), info.NotAfter.Format(time.DateTime), p)
}
func InfoOf(certificate *x509.Certificate) *CertInfo {
	return &CertInfo{
		Domains:   certificate.DNSNames,
		NotBefore: certificate.NotBefore,
		NotAfter:  certificate.NotAfter,
		Issuer:    certificate.Issuer.String(),
		Subject:   certificate.Subject.String(),
		IsCA:      certificate.IsCA,
	}
}
func ParseCertBytes(certBytes []byte) (*CertInfo, error) {
	var certInfo *CertInfo = nil
	var res *CertInfo
	for {
		if len(certBytes) <= 0 {
			break
		}
		certBlock, rest := pem.Decode(certBytes)
		cert, err := x509.ParseCertificate(certBlock.Bytes)
		if err != nil {
			return nil, err
		}
		info := InfoOf(cert)
		if certInfo == nil {
			certInfo = info
			res = info
		} else {
			certInfo.Parent = info
			certInfo = info
		}
		certBytes = rest
	}
	return res, nil
}

func (acme *ACME) LoadCerts() ([]*CertInfo, error) {
	res := make([]*CertInfo, 0)
	return res, filepath.Walk(acme.Config.FilePath, func(path string, info fs.FileInfo, err error) error {
		if !strings.HasSuffix(path, certExt) {
			return nil
		}
		if strings.HasSuffix(path, issuerExt) {
			return nil
		}
		cert, err := loadFromFile(path)
		if err == nil {
			cert.File = info.Name()
			cert.Path = path
			res = append(res, cert)
		}
		return err
	})
}

func loadFromFile(file string) (*CertInfo, error) {
	bytes, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return ParseCertBytes(bytes)
}

func (acme *ACME) Renew(path string) error {
	info, err := loadFromFile(path)
	if err != nil {
		return err
	}
	notAfter := int(time.Until(info.NotAfter).Hours() / 24.0)
	if notAfter > 3 {
		return fmt.Errorf("the certificate expires in %d days, no renewal", notAfter)
	}

	privateKey, _ := loadKeyFromFile(strings.TrimSuffix(path, certExt) + keyExt)
	request := certificate.ObtainRequest{
		Domains:    info.Domains,
		PrivateKey: privateKey,
		Bundle:     true,
	}
	// 请求证书
	certificates, err := acme.client.Certificate.Obtain(request)
	if err != nil {
		return err
	}
	log.Println("renew certificates success")
	if err = acme.SaveResource(certificates); err != nil {
		return fmt.Errorf("save certs failed:%v", err)
	}
	log.Printf("certificates saved to %s", acme.pathOf(certificates.Domain))
	return nil
}
