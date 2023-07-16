package web

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type response struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
	Data   any    `json:"data"`
}

func successWithMsgAndData(ctx *gin.Context, msg string, data any) {
	ctx.JSON(http.StatusOK, response{
		0, msg, data,
	})
}
func successWithData(ctx *gin.Context, data any) {
	successWithMsgAndData(ctx, "success", data)
}

func successWithMessage(context *gin.Context, msg string) {
	successWithMsgAndData(context, msg, nil)
}

func success(context *gin.Context) {
	successWithData(context, nil)
}

func errorWithMsg(context *gin.Context, msg string) {
	errorWithMsgAndCode(context, msg, http.StatusInternalServerError)
}

func responseError(ctx *gin.Context, err error) {
	errorWithMsg(ctx, err.Error())
}

func errorWithMsgAndCode(context *gin.Context, msg string, code int) {
	context.JSON(http.StatusInternalServerError, response{
		code, msg, nil,
	})
}
