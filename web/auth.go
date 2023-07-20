package web

import (
	"github.com/gin-gonic/gin"
)

type Auth interface {
	isLogin(ctx *gin.Context) bool
	redirectToLogin(ctx *gin.Context)
	logout(ctx *gin.Context)
}

var auth Auth

func validate(ctx *gin.Context) {
	if auth == nil || auth.isLogin(ctx) {
		ctx.Next()
		return
	} else {
		auth.redirectToLogin(ctx)
		ctx.Abort()
	}
}
