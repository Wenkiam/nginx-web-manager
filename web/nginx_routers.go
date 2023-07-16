package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"nwm/nginx"
)

func setupNginxRouters() {
	group := engine.Group("/nginx")
	group.GET("/reload", reloadNginx)
	group.GET("/configs", allConfig)
	group.POST("/config/path", setConfigPath)
	group.GET("/config/path", getConfigPath)
	group.POST("/config/save", saveConfig)
	group.DELETE("/config/:name", delConfig)
}
func allConfig(ctx *gin.Context) {
	configs, err := nginx.AllConfigs()
	if err != nil {
		responseError(ctx, err)
	} else {
		successWithData(ctx, configs)
	}

}

func setConfigPath(ctx *gin.Context) {
	var body = map[string]string{}
	err := ctx.ShouldBindJSON(&body)
	err = nginx.SetConfigDir(body["path"])
	if err != nil {
		errorWithMsg(ctx, fmt.Sprintf("set path failed.%s", err.Error()))
	} else {
		successWithData(ctx, map[string]string{
			"path": nginx.GetPath(),
		})
	}
}

func getConfigPath(ctx *gin.Context) {
	successWithData(ctx, map[string]string{
		"path": nginx.GetPath(),
	})
}

func saveConfig(ctx *gin.Context) {
	conf := &nginx.Config{}
	err := ctx.ShouldBindJSON(conf)
	if err != nil {
		responseError(ctx, err)
		return
	}
	err = nginx.SaveConfig(conf)
	if err != nil {
		responseError(ctx, err)
	} else {
		success(ctx)
	}
}
func delConfig(ctx *gin.Context) {
	name := ctx.Param("name")
	err := nginx.DelConf(name)
	if err != nil {
		responseError(ctx, err)
	} else {
		success(ctx)
	}
}

func reloadNginx(ctx *gin.Context) {
	if err := nginx.Reload(); err != nil {
		responseError(ctx, err)
	} else {
		success(ctx)
	}
}
