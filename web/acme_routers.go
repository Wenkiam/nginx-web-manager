package web

import (
	"github.com/gin-gonic/gin"
	"log"
	acme2 "nwm/acme"
	"path/filepath"
)

func setupACMERouters() {
	group := engine.Group("/acme")
	group.Use(validate)
	group.POST("/generate", generateCerts)
	group.GET("/config", acmeConfig)
	group.GET("/certs", loadCertInfoFromFiles)
	group.POST("/config", saveAcmeConfig)
	group.POST("/renew", renew)
	group.POST("/download", download)
	group.GET("/providers", supportedProviders)
}

func generateCerts(ctx *gin.Context) {
	var body = map[string][]string{}
	err := ctx.ShouldBindJSON(&body)
	if err != nil {
		responseError(ctx, err)
		return
	}
	domains := body["domains"]
	err = acme.GenerateCerts(domains)
	if err != nil {
		responseError(ctx, err)
	} else {
		success(ctx)
	}
}

func acmeConfig(ctx *gin.Context) {
	successWithData(ctx, acme.Config)
}
func saveAcmeConfig(ctx *gin.Context) {
	config := &acme2.Config{}
	err := ctx.ShouldBindJSON(config)
	if err != nil {
		responseError(ctx, err)
		return
	}
	a := &acme2.ACME{Config: config}
	err = a.Setup()
	if err != nil {
		a.Shutdown()
		responseError(ctx, err)
		return
	}
	acme.Shutdown()
	acme = a
	success(ctx)
}

func loadCertInfoFromFiles(ctx *gin.Context) {
	certs, err := acme.LoadCerts()
	if err != nil {
		responseError(ctx, err)
	} else {
		kvs := make([]map[string]string, len(certs))
		for i, cert := range certs {
			kvs[i] = map[string]string{
				"file": cert.File,
				"info": cert.String(),
				"path": cert.Path,
			}
		}
		successWithData(ctx, kvs)
	}
}

func renew(ctx *gin.Context) {
	file := ctx.Request.FormValue("file")
	if file == "" {
		errorWithMsg(ctx, "param file is empty")
		return
	}
	err := acme.Renew(file)
	if err != nil {
		responseError(ctx, err)
	} else {
		success(ctx)
	}
}
func download(ctx *gin.Context) {
	path := ctx.Request.FormValue("path")
	files, err := acme2.GetCertFiles(path)
	if err != nil {
		responseError(ctx, err)
		return
	}
	ctx.Header("Content-Type", "application/zip")
	ctx.Header("Content-Disposition", "attachment; filename="+filepath.Base(path)+".zip")
	ctx.Header("Content-Transfer-Encoding", "binary")
	err = acme2.WriteInZip(ctx.Writer, files)
	if err != nil {
		log.Printf("creat zip archive failed:%v", err)
		responseError(ctx, err)
	}
}

func supportedProviders(ctx *gin.Context) {
	successWithData(ctx, []string{
		"acme-dns",
		"alidns",
		"allinkl",
		"arvancloud",
		"azure",
		"auroradns",
		"autodns",
		"bindman",
		"bluecat",
		"brandit",
		"bunny",
		"checkdomain",
		"civo",
		"clouddns",
		"cloudflare",
		"cloudns",
		"cloudxns",
		"conoha",
		"constellix",
		"derak",
		"desec",
		"designate",
		"digitalocean",
		"dnshomede",
		"dnsimple",
		"dnsmadeeasy",
		"dnspod",
		"dode",
		"domeneshop",
		"domainnameshop",
		"dreamhost",
		"duckdns",
		"dyn",
		"dynu",
		"easydns",
		"edgedns",
		"fastdns",
		"fastdns",
		"efficientip",
		"epik",
		"exec",
		"exoscale",
		"freemyip",
		"gandi",
		"gandiv5",
		"gcloud",
		"gcore",
		"glesys",
		"godaddy",
		"googledomains",
		"hetzner",
		"hostingde",
		"hosttech",
		"httpreq",
		"hurricane",
		"hyperone",
		"ibmcloud",
		"iij",
		"iijdpf",
		"infoblox",
		"infomaniak",
		"internetbs",
		"inwx",
		"ionos",
		"iwantmyname",
		"joker",
		"liara",
		"lightsail",
		"linode",
		"linodev4",
		"linodev4",
		"liquidweb",
		"luadns",
		"loopia",
		"manual",
		"mydnsjp",
		"mythicbeasts",
		"namecheap",
		"namedotcom",
		"namesilo",
		"nearlyfreespeech",
		"netcup",
		"netlify",
		"nicmanager",
		"nifcloud",
		"njalla",
		"nodion",
		"ns1",
		"oraclecloud",
		"otc",
		"ovh",
		"pdns",
		"plesk",
		"porkbun",
		"rackspace",
		"rcodezero",
		"regru",
		"rfc2136",
		"rimuhosting",
		"route53",
		"safedns",
		"sakuracloud",
		"scaleway",
		"selectel",
		"servercow",
		"simply",
		"sonic",
		"stackpath",
		"tencentcloud",
		"transip",
		"ultradns",
		"variomedia",
		"vegadns",
		"vercel",
		"versio",
		"vinyldns",
		"vkcloud",
		"vscale",
		"vultr",
		"websupport",
		"wedos",
		"yandex",
		"yandexcloud",
		"zoneee",
		"zonomi",
	})
}
