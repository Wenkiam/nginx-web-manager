package web

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"nwm/nginx"
	"nwm/utils"
)

func setupNginxRouters() {
	group := engine.Group("/nginx", validate)
	group.GET("/reload", reloadNginx)
	group.GET("/configs", allConfig)
	group.POST("/config/path", setConfigPath)
	group.GET("/config/path", getConfigPath)
	group.POST("/config/save", saveConfig)
	group.DELETE("/config/:name", delConfig)
	group.GET("/path/history", pathHistory)
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
	path := body["path"]
	err = nginx.SetConfigDir(path)
	if err != nil {
		errorWithMsg(ctx, fmt.Sprintf("set path failed.%s", err.Error()))
	} else {
		session := sessions.Default(ctx)
		paths, ok := session.Get("nginx.path.history").([]string)
		if !ok {
			paths = make([]string, 0)
		}
		set := utils.SetOf(paths)
		set.Add(path)
		session.Set("nginx.path.history", set.ToSlice())
		session.Save()
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

func pathHistory(ctx *gin.Context) {
	session := sessions.Default(ctx)
	session.Get("nginx.path.history")
	paths, ok := session.Get("nginx.path.history").([]string)
	if !ok {
		paths = make([]string, 2)
		paths[0] = "/etc/nginx/"
		paths[1] = "/etc/nginx/conf.d/"
	}
	successWithData(ctx, paths)
}
