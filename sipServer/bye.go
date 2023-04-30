package sipServer

import (
	"gb28181Panda/log"
	"github.com/ghettovoice/gosip/sip"
	"strings"
)

func Bye(req sip.Request, tx sip.ServerTransaction) {
	idx := strings.Index(req.Source(), ":")
	fromIp := req.Source()[:idx]
	fromPort := req.Source()[idx+1:]
	transferFromLog(fromIp, fromPort)
	log.Info("收到Bye数据", req)
}
