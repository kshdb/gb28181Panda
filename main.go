package main

import (
	"gb28181Panda/router"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	//初始化sip服务
	router.InitSipServer()
	//设置gin的运行环境
	gin.SetMode(viper.GetString("runMode"))
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	_ = r.SetTrustedProxies([]string{viper.GetString("ip")})
	router.InitGinRouter(r)
	_ = r.Run(viper.GetString("addr")) // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
