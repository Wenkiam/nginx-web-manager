package acme

import (
	"crypto"
	"encoding/json"
	"fmt"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"github.com/robfig/cron"
	"github.com/urfave/cli/v2"
	"log"
	"nwm/utils"
	"os"
	"path/filepath"
	"time"
)

var (
	keyType = certcrypto.RSA2048
)

type ACME struct {
	Config       *Config
	Registration *registration.Resource `json:"registration"`
	Key          crypto.PrivateKey      `json:"-"`
	client       *lego.Client
	Cron         *cron.Cron
}

type Config struct {
	FilePath    string            `json:"filePath"`
	Email       string            `json:"email"`
	CronExpress string            `json:"cron"`
	Dns         string            `json:"dns"`
	CADirURL    string            `json:"url"`
	Envs        map[string]string `json:"envs"`
	CronEnable  bool              `json:"cronEnable"`
}

func (acme *ACME) GetRegistration() *registration.Resource {
	return acme.Registration
}

func (acme *ACME) GetPrivateKey() crypto.PrivateKey {
	return acme.Key
}

func (acme *ACME) GetEmail() string {
	return acme.Config.Email
}
func FromConfig(config *Config) *ACME {
	return &ACME{
		Config: config,
	}
}
func New(ctx *cli.Context) *ACME {
	filePath := ctx.String("path")
	err := os.MkdirAll(filePath, filePerm)
	if err != nil {
		log.Fatalf("create file path failed.%v", err)
	}

	config := loadConfig(filePath)
	if email := ctx.String("email"); email != "" {
		config.Email = email
	}
	if dns := ctx.String("dns"); dns != "" {
		config.Dns = dns
	}
	if CADirURL := ctx.String("url"); CADirURL != "" {
		config.CADirURL = CADirURL
	}

	if CronExpress := ctx.String("cron"); CronExpress != "" {
		config.CronExpress = CronExpress
	}
	if config.CronExpress == "" {
		config.CronExpress = "0 0 0 * * ?"
	}
	if config.CADirURL == "" {
		config.CADirURL = "https://acme-v02.api.letsencrypt.org/directory"
	}
	return FromConfig(config)
}
func loadConfig(filePath string) *Config {
	file := filepath.Join(filePath, "config.json")
	config := &Config{
		CronEnable: true,
	}
	config.FilePath = filePath
	bytes, err := os.ReadFile(file)
	if err != nil {
		return config
	}
	err = json.Unmarshal(bytes, config)
	return config
}
func (acme *ACME) GenerateCerts(domains []string) error {
	err := acme.checkDomains(domains)
	if err != nil {
		return err
	}
	request := certificate.ObtainRequest{
		Domains: domains,
		Bundle:  true,
	}

	certificates, err := acme.client.Certificate.Obtain(request)
	if err != nil {
		return err
	}
	log.Println("request for certificates success")
	if err = acme.SaveResource(certificates); err != nil {
		return fmt.Errorf("save cert files for %s failed:%v", domains, err)
	}
	return nil
}
func (acme *ACME) checkDomains(domains []string) error {
	certInfos, err := acme.LoadCerts()
	if err != nil {
		return err
	}
	domainSet := utils.NewSet[string]()
	now := time.Now()
	for _, info := range certInfos {
		if now.After(info.NotAfter) {
			continue
		}
		domainSet.AddAll(info.Domains)
	}
	for _, d := range domains {
		if domainSet.Contains(d) {
			return fmt.Errorf("cert of domain %s is still usable", d)
		}
	}
	return nil
}
