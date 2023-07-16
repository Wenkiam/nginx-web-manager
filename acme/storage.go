package acme

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-acme/lego/v4/certificate"
	"golang.org/x/net/idna"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	keyExt       = ".key"
	certExt      = ".crt"
	issuerExt    = ".issuer.crt"
	jsonExt      = ".json"
	pemExt       = ".pem"
	keyPemExt    = "_key.pem"
	pemBundleExt = "_bundle" + pemExt
	filePerm     = 0644
)

func (acme *ACME) SaveResource(certRes *certificate.Resource) error {
	domain := certRes.Domain
	err := acme.WriteFile(domain, certExt, certRes.Certificate)
	if err != nil {
		log.Printf("Unable to save Certificate for domain %s\n\t%v", domain, err)
		return err
	}

	if certRes.IssuerCertificate != nil {
		err = acme.WriteFile(domain, issuerExt, certRes.IssuerCertificate)
		if err != nil {
			log.Printf("Unable to save IssuerCertificate for domain %s\n\t%v", domain, err)
			return err
		}
	}

	// if we were given a CSR, we don't know the private key
	if certRes.PrivateKey != nil {
		err = acme.WriteCertificateFiles(domain, certRes)
		if err != nil {
			log.Printf("Unable to save PrivateKey for domain %s\n\t%v", domain, err)
			return err
		}
	}

	jsonBytes, err := json.MarshalIndent(certRes, "", "\t")
	if err != nil {
		log.Printf("Unable to marshal CertResource for domain %s\n\t%v", domain, err)
		return err
	}

	err = acme.WriteFile(domain, jsonExt, jsonBytes)
	if err != nil {
		log.Printf("Unable to save CertResource for domain %s\n\t%v", domain, err)
	}
	return err
}

func (acme *ACME) WriteFile(domain, extension string, data []byte) error {
	var baseFileName string
	baseFileName = sanitizedDomain(domain)
	dir := acme.pathOf(domain)
	info, err := os.Stat(dir)
	if info != nil && !info.IsDir() {
		return fmt.Errorf("%s exists,but not a directory", dir)
	}
	if os.IsNotExist(err) {
		err = os.Mkdir(dir, filePerm)
	}
	if err != nil {
		return fmt.Errorf("can't write file to %s:%v", dir, err)
	}

	filePath := filepath.Join(dir, baseFileName+extension)
	return os.WriteFile(filePath, data, filePerm)
}
func (acme *ACME) WriteCertificateFiles(domain string, certRes *certificate.Resource) error {
	err := acme.WriteFile(domain, keyExt, certRes.PrivateKey)
	if err != nil {
		return fmt.Errorf("unable to save key file: %w", err)
	}

	err = acme.WriteFile(domain, keyPemExt, certRes.PrivateKey)
	if err != nil {
		return fmt.Errorf("unable to save PEM key file: %w", err)
	}
	err = acme.WriteFile(domain, pemExt, certRes.Certificate)
	if err != nil {
		return fmt.Errorf("unable to save PEM key file: %w", err)
	}
	err = acme.WriteFile(domain, pemBundleExt, bytes.Join([][]byte{certRes.Certificate, certRes.PrivateKey}, nil))
	if err != nil {
		return fmt.Errorf("unable to save PEM file: %w", err)
	}
	return nil
}

func sanitizedDomain(domain string) string {
	safe, err := idna.ToASCII(strings.NewReplacer(":", "-", "*", "_").Replace(domain))
	if err != nil {
		log.Printf("sanitizedDomain %s failed:%v", domain, err)
	}
	return safe
}

func (acme *ACME) saveRegistration() error {
	jsonBytes, err := json.MarshalIndent(acme.Registration, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(acme.pathOf(acme.Config.Email+".json"), jsonBytes, filePerm)
}
func (acme *ACME) pathOf(file string) string {
	return filepath.Join(acme.Config.FilePath, file)
}

func (config *Config) save() error {
	file := filepath.Join(config.FilePath, "config.json")
	jsonBytes, _ := json.MarshalIndent(config, "", "\t")
	return os.WriteFile(file, jsonBytes, filePerm)
}
func GetCertFiles(filePath string) ([]string, error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	if stat.IsDir() {
		return nil, fmt.Errorf("file path invalid")
	}
	dir := filepath.Dir(filePath)
	files, err := os.ReadDir(dir)
	certFiles := make([]string, len(files))
	index := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fileName := file.Name()
		isCert := strings.HasSuffix(fileName, certExt) || strings.HasSuffix(fileName, pemExt) || strings.HasSuffix(fileName, keyExt)
		if !isCert {
			continue
		}
		certFiles[index] = filepath.Join(dir, fileName)
		index++
	}
	if index == 0 {
		return nil, fmt.Errorf("no cert file found")
	}
	return certFiles[0:index], nil
}
func WriteInZip(fw io.Writer, files []string) (err error) {
	zw := zip.NewWriter(fw)
	for _, file := range files {
		fileName := filepath.Base(file)
		w, err := zw.Create(fileName)
		if err != nil {
			return err
		}
		fr, err := os.Open(file)
		if err != nil {
			continue
		}
		defer fr.Close()
		_, err = io.Copy(w, fr)
		if err != nil {
			return err
		}
	}
	return zw.Close()
}
