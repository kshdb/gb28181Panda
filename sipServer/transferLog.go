package sipServer

import (
	"fmt"
	"gb28181Panda/config"
	"gb28181Panda/log"
)

func transferFromLog(fromIp, fromPort string) {
	log.Info(fmt.Sprintf("[%s:%d]<<<<<<[%s:%s]", config.SipOp.Ip, config.SipOp.Port, fromIp, fromPort))
}
func transferToLog(toIp, toPort string) {
	log.Info(fmt.Sprintf("[%s:%d]>>>>[%s:%s]", config.SipOp.Ip, config.SipOp.Port, toIp, toPort))
}
