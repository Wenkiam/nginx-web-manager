package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/urfave/cli/v2"
	"html/template"
	"log"
	"net/http"
	acme2 "nwm/acme"
	"nwm/html"
)

var (
	port   = 8080
	engine *gin.Engine
	acme   *acme2.ACME
)

func Setup(ctx *cli.Context) {
	engine = gin.Default()
	port = ctx.Int("port")
	acme = acme2.New(ctx)
	err := acme.Setup()
	if err != nil {
		log.Printf("ACME setup failed: %v. You can setup ACME from web console later", err)
	} else {
		log.Printf("ACME setup success")
	}
	setupRouters()
}
func StartServer() error {
	//	engine.LoadHTMLGlob("html/template/**")
	err := engine.Run(fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("start http server failed.%v", err)
	}
	return nil
}
func setupRouters() {
	engine.StaticFS("assets", http.FS(html.Static))
	engine.SetHTMLTemplate(template.Must(template.New("").ParseFS(html.Template, "template/**")))
	engine.GET("/", index)

	setupNginxRouters()
	setupACMERouters()

}
func index(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "index.html", gin.H{})
}
