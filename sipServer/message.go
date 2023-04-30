package sipServer

import (
	"encoding/xml"
	"gb28181Panda/log"
	"gb28181Panda/model"
	"gb28181Panda/util"
	"github.com/ghettovoice/gosip/sip"
	"net/http"
	"strings"
)

type MessageStruct struct {
	XMLName      xml.Name
	CmdType      string           `xml:"CmdType"`
	SN           string           `xml:"SN"`
	DeviceId     string           `json:"DeviceId" xml:"DeviceID"`
	DeviceType   string           `xml:"DeviceType"`
	DeviceName   string           `xml:"DeviceName"`
	Result       string           `xml:"Result"`
	Manufacturer string           `xml:"Manufacturer"`
	Model        string           `xml:"Model"`
	Channel      string           `xml:"Channel"`
	Firmware     string           `xml:"Firmware"`
	DeviceList   []*model.Channel `xml:"DeviceList>Item"`
}

// Message Sip的Message
func Message(req sip.Request, tx sip.ServerTransaction) {
	idx := strings.Index(req.Source(), ":")
	fromIp := req.Source()[:idx]
	fromPort := req.Source()[idx+1:]
	transferFromLog(fromIp, fromPort)
	log.Info("MESSAGE-Request:\n", req)
	if l, ok := req.ContentLength(); !ok || l.Equals(0) {
		log.Info("该MESSAGE消息的消息体长度为0，返回OK")
		_ = tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, http.StatusText(http.StatusOK), ""))
	}
	body := req.Body()
	//xml GB2312 解包有问题 换成UTF-8
	body = strings.Replace(body, "GB2312", "UTF-8", 1)
	msgData := &MessageStruct{}
	if err := xml.Unmarshal([]byte(body), msgData); err != nil {
		log.Error("解析deviceInfo响应包出错", err)
		return
	}
	if msgData.XMLName.Local == "Response" {
		switch msgData.CmdType {
		case "Keepalive":
			device := model.Device{
				DeviceId:    msgData.DeviceId,
				KeepaliveAt: util.GetCurrenTimeNow(),
			}
			//更新device设备信息
			_ = device.DeviceUpdate()
		case "DeviceInfo":
			device := model.Device{
				Name:         msgData.DeviceType,
				Manufacturer: msgData.Manufacturer,
				Model:        msgData.Model,
				Firmware:     msgData.Firmware,
				DeviceId:     msgData.DeviceId,
			}
			//更新device设备信息
			_ = device.DeviceUpdate()
		case "Catalog":
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
		case "RecordInfo":
			_ = sipMessageRecordInfo(body)
		default:
			log.Infof("【cmdType】", msgData.CmdType)
		}
		_ = tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, http.StatusText(http.StatusOK), ""))
	} else if msgData.XMLName.Local == "Notify" {
		//下级设备保持keepalive是走的Message的notify 区别于Notify本身
		device := model.Device{
			DeviceId:    msgData.DeviceId,
			KeepaliveAt: util.GetCurrenTimeNow(),
		}
		//更新device设备信息
		_ = device.DeviceUpdate()
		//需要回复信息  不然设备会发送注销再走注册流程
		errBack := tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, http.StatusText(http.StatusOK), ""))
		transferToLog(fromIp, fromPort)
		log.Info("回复设备保持在线成功")
		if errBack != nil {
			log.Info("回复设备keepalive出错了")
			//只尝试一次再次回复
			_ = tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, http.StatusText(http.StatusOK), ""))
		}
	}

}
