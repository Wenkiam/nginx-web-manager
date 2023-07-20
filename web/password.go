package web

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/urfave/cli/v2"
	"net/http"
)

func init() {
	Flags = append(Flags, &cli.StringFlag{
		Name:    "password",
		Usage:   "password to login system",
		EnvVars: []string{"AUTH_PASSWORD"},
	})
}
func initPasswordAuth(ctx *cli.Context) {
	password := ctx.String("password")
	if password == "" {
		return
	}
	auth = &passwordAuth{
		password,
	}
	engine.POST("/login", func(context *gin.Context) {
		p := context.PostForm("password")
		if password == p {
			session := sessions.Default(context)
			session.Set("password", p)
			session.Save()
			success(context)
		} else {
			errorWithMsg(context, "invalid password")
		}
	})
	engine.GET("/login", func(context *gin.Context) {
		context.HTML(http.StatusOK, "login.html", gin.H{})
	})

}

type passwordAuth struct {
	password string
}

func (check *passwordAuth) isLogin(ctx *gin.Context) bool {
	session := sessions.Default(ctx)
	password := session.Get("password")
	p, ok := password.(string)
	return ok && p == check.password
}

func (check *passwordAuth) redirectToLogin(ctx *gin.Context) {
	redirect(ctx, "/login")
}

func (check *passwordAuth) logout(ctx *gin.Context) {
	session := sessions.Default(ctx)
	session.Delete("password")
	session.Save()
	ctx.Redirect(http.StatusFound, "/")
}
