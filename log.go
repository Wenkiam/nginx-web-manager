package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	lego "github.com/go-acme/lego/v4/log"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func setupLog(ctx *cli.Context) error {
	logPath := ctx.String("log")
	if strings.TrimSpace(logPath) == "" {
		return nil
	}
	err := os.MkdirAll(logPath, 0644)
	if err != nil {
		return fmt.Errorf("create log directory failed:%v", err)
	}

	ginLog, err := os.OpenFile(filepath.Join(logPath, "gin.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("setup gin log failed:%v", err)
	}
	gin.DefaultWriter = ginLog
	sysLog, err := os.OpenFile(filepath.Join(logPath, "sys.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("setup sys log failed:%v", err)
	}
	log.SetOutput(sysLog)
	lego.Logger = log.Default()
	return nil
}
