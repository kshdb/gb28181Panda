package router

import (
	"gb28181Panda/sipServer"
)

// InitSipServer 初始化sip服务
func InitSipServer() {
	sipServer.NewServer()
}
