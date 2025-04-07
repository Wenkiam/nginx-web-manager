package acme

import (
	"github.com/robfig/cron"
	"log"
	"nwm/nginx"
	"time"
)

func (acme *ACME) setupCron() error {
	if acme.Cron != nil {
		return nil
	}
	if acme.Config.CronExpress == "" || !acme.Config.CronEnable {
		return nil
	}
	acme.Cron = cron.New()
	err := acme.Cron.AddJob(acme.Config.CronExpress, acme)
	if err == nil {
		acme.Cron.Start()
	}
	return err
}

func (acme *ACME) Run() {
	log.Printf("try renew certs")
	certs, err := acme.LoadCerts()
	if err != nil {
		log.Printf("load certs failed:%v", err)
		return
	}
	if len(certs) <= 0 {
		log.Printf("no certs found")
		return
	}
	reloadNginx := false
	for _, cert := range certs {
		notAfter := int(time.Until(cert.NotAfter).Hours() / 24.0)
		log.Printf("try renew cert file for %s, path:%s, file:%s", cert.Domains, cert.Path, cert.File)
		if notAfter > 3 {
			log.Printf("cert of %s expire in %d days, no renewal", cert.Domains, notAfter)
			continue
		}
		if err := acme.Renew(cert.Path); err != nil {
			log.Printf("renew cert of %s failed:%v", cert.Domains, err)
		} else {
			log.Printf("renew cert of %s success", cert.Domains)
			reloadNginx = true
		}
	}
	if reloadNginx && nginx.Reload() != nil {
		log.Printf("reload nginx success")
	}
}

func (acme *ACME) Shutdown() {
	if acme.Cron != nil {
		acme.Cron.Stop()
	}
}
