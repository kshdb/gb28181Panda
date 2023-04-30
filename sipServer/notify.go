package sipServer

import (
	"encoding/xml"
	"gb28181Panda/log"
	"gb28181Panda/util"
	"github.com/ghettovoice/gosip/sip"
	"net/http"
	"strings"
)

// Notify 通知
func Notify(req sip.Request, tx sip.ServerTransaction) {
	idx := strings.Index(req.Source(), ":")
	fromIp := req.Source()[:idx]
	fromPort := req.Source()[idx+1:]
	log.Info("Notify-Request:\n", req)
	transferFromLog(fromIp, fromPort)
	if l, ok := req.ContentLength(); !ok || l.Equals(0) {
		log.Debug("该MESSAGE消息的消息体长度为0，返回OK")
		_ = tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, http.StatusText(http.StatusOK), ""))
	}
	body := req.Body()
	body = strings.Replace(body, "GB2312", "UTF-8", 1)
	msgData := &MessageStruct{}
	if err := xml.Unmarshal([]byte(body), msgData); err != nil {
		log.Error("解析deviceInfo响应包出错", err)
		return
	}
	log.Info("Notify-Type:", msgData.CmdType)
	if msgData.XMLName.Local == "Notify" {
		switch msgData.CmdType {
		case "Catalog":
			log.Info("下级主动推送channel通道信息")
			//下级主动推送通道信息
			for _, item := range msgData.DeviceList {
				//先查询当前设备是否在数据库中
				channel, _ := item.ChannelDetail()
				if channel.DeviceId != "" { //更新
					item.UpdatedAt = util.GetCurrenTimeNow()
					item.ParentId = msgData.DeviceId
					_ = item.ChannelUpdate()
				} else { // 新增
					item.CreatedAt = util.GetCurrenTimeNow()
					item.UpdatedAt = util.GetCurrenTimeNow()
					item.ParentId = msgData.DeviceId
					item.TransportType = "UDP"
					item.MediaStatus = "CLOSE"
					_ = item.ChannelAdd()
				}
			}
		default:
			log.Infof("【cmdType】", msgData.CmdType)
		}
		_ = tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, http.StatusText(http.StatusOK), ""))
	}

}
