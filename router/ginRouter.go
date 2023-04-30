package router

import (
	"github.com/gin-gonic/gin"
)

// InitGinRouter 初始化路由
func InitGinRouter(g *gin.Engine) {
	//设备路由
	DeviceRouter(g.Group("/api/device"))
	//通道路由
	ChannelRouter(g.Group("/api/channel"))
	//ZLM的hook回调接口
	HookRouter(g.Group("/index/hook"))
}
