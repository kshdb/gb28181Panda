package sipServer

import (
	"fmt"
	"gb28181Panda/config"
	"gb28181Panda/log"
	"github.com/ghettovoice/gosip"
	l "github.com/ghettovoice/gosip/log"
	"github.com/ghettovoice/gosip/sip"
	"sync"
)

var (
	SipServer  gosip.Server
	serverOnce sync.Once
)

func NewServer() {
	serverOnce.Do(func() {
		SipServer = gosip.NewServer(
			gosip.ServerConfig{
				UserAgent: config.SipOp.UserAgent,
			},
			nil,
			nil,
			l.NewDefaultLogrusLogger(),
		)
	})
	if err := ListenTCP(); err != nil {
		log.Error("初始化SIP-TCP服务错误", fmt.Sprintf("%s:%d", config.SipOp.Ip, config.SipOp.Port))
		return
	}
	log.Info("初始化SIP-TCP服务成功", fmt.Sprintf("%s:%d", config.SipOp.Ip, config.SipOp.Port))
	if err := ListenUDP(); err != nil {
		log.Error("初始化SIP-UDP服务错误", fmt.Sprintf("%s:%d", config.SipOp.Ip, config.SipOp.Port))
		return
	}
	log.Info("初始化SIP-UDP服务成功", fmt.Sprintf("%s:%d", config.SipOp.Ip, config.SipOp.Port))
	registerHandler()
	_recordList = &sync.Map{}
}

func ListenTCP() error {
	return SipServer.Listen("tcp", fmt.Sprintf("%s:%d", config.SipOp.Ip, config.SipOp.Port), nil)
}
func ListenUDP() error {
	return SipServer.Listen("udp", fmt.Sprintf("%s:%d", config.SipOp.Ip, config.SipOp.Port), nil)
}

func registerHandler() {
	_ = SipServer.OnRequest(sip.REGISTER, Register)
	_ = SipServer.OnRequest(sip.MESSAGE, Message)
	_ = SipServer.OnRequest(sip.NOTIFY, Notify)
	_ = SipServer.OnRequest(sip.INVITE, Invite)
	_ = SipServer.OnRequest(sip.ACK, Ack)
	_ = SipServer.OnRequest(sip.BYE, Bye)
}
