package router

import (
	"awesomeProject/handler"
	"awesomeProject/svc"
	"github.com/gin-gonic/gin"
)

// https://codeantenna.com/a/bHJN5aSqgM
func HandleFunc(handler func(c *svc.ServiceContext)) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		handler(&svc.ServiceContext{Ctx: ctx})
	}
}

func RegisterHandlers(server *gin.Engine, serverCtx *svc.ServiceContext) {
	server.GET("/ping", handler.Ping(serverCtx))
}
