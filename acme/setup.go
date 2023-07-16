package acme

import (
	"crypto"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns"
	"github.com/go-acme/lego/v4/registration"
	"log"
	"os"
	"strings"
)

func (acme *ACME) Setup() error {
	if acme.GetEmail() == "" {
		return fmt.Errorf("email is not present")
	}
	err := acme.setupPrivateKey()
	if err != nil {
		return err
	}
	err = acme.tryLoadRegistration()
	if err != nil {
		return err
	}

	err = acme.setupClient()
	if err != nil {
		return fmt.Errorf("setup client failed.%v", err)
	}

	if acme.Registration == nil {
		err = acme.newRegistration()
	}
	if err != nil {
		return fmt.Errorf("setup new registration failed.%v", err)
	}
	err = acme.setupDnsProvider()
	if err != nil {
		return err
	}
	err = acme.setupCron()
	if err != nil {
		return fmt.Errorf("start cron failed.%v", err)
	}
	return acme.Config.save()
}
func (acme *ACME) setupClient() error {
	config := lego.NewConfig(acme)
	config.Certificate.KeyType = keyType
	config.CADirURL = acme.Config.CADirURL
	client, err := lego.NewClient(config)
	acme.client = client
	return err
}
func (acme *ACME) setupPrivateKey() error {
	keyfile := acme.pathOf(acme.Config.Email + keyExt)
	key, err := loadKeyFromFile(keyfile)
	if err != nil {
		return fmt.Errorf("load key from file %s failed:%v", keyfile, err)
	} else {
		acme.Key = key
	}
	if acme.Key == nil {
		key, err = newPrivateKey(keyfile)
	}
	if err != nil {
		return fmt.Errorf("create new private key failed.%v", err)
	}
	acme.Key = key
	log.Printf("setup private key success")
	return nil
}
func loadKeyFromFile(keyfile string) (crypto.PrivateKey, error) {
	keyBytes, err := os.ReadFile(keyfile)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	keyBlock, _ := pem.Decode(keyBytes)
	return x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
}

func newPrivateKey(file string) (crypto.PrivateKey, error) {
	privateKey, err := certcrypto.GeneratePrivateKey(keyType)
	if err != nil {
		return nil, err
	}
	certOut, err := os.Create(file)
	if err != nil {
		return nil, err
	}
	defer certOut.Close()
	pemKey := certcrypto.PEMBlock(privateKey)
	err = pem.Encode(certOut, pemKey)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}
func (acme *ACME) tryLoadRegistration() error {
	userBytes, err := os.ReadFile(acme.pathOf(acme.Config.Email + ".json"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("load registration failed.%v", err)
	}
	reg := &registration.Resource{}
	err = json.Unmarshal(userBytes, reg)
	if err != nil {
		return fmt.Errorf("could not parse file for account %s: %v", acme.Config.Email, err)
	}
	log.Printf("load registration success:%v", *reg)
	acme.Registration = reg
	return nil
}
func (acme *ACME) newRegistration() error {
	reg, err := acme.client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return err
	}
	acme.Registration = reg
	return acme.saveRegistration()
}

func (acme *ACME) setupDnsProvider() error {
	envs := acme.Config.Envs
	for k, v := range envs {
		if strings.TrimSpace(k) == "" || strings.TrimSpace(v) == "" {
			continue
		}
		err := os.Setenv(k, v)
		if err != nil {
			log.Printf("set env %s=%s failed:%v", k, v, err)
		}
	}
	dnsProvider, err := dns.NewDNSChallengeProviderByName(acme.Config.Dns)
	if err != nil {
		return fmt.Errorf("setup dns provider failed.%v", err)
	}
	err = acme.client.Challenge.SetDNS01Provider(dnsProvider)
	if err != nil {
		return fmt.Errorf("set dns provider failed.%v", err)
	}
	return nil
}
