package handler

import (
	"awesomeProject/logic"
	"awesomeProject/svc"
	"github.com/gin-gonic/gin"
)

func Ping(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		pingLogic := logic.NewPingLogic(c, svcCtx)
		resp, err := pingLogic.Ping()
		if err != nil {
			c.JSON(500, err)
		}
		c.JSON(200, resp)
	}
}
