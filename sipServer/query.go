package sipServer

import (
	"fmt"
	"gb28181Panda/config"
	"gb28181Panda/log"
	"gb28181Panda/model"
	"gb28181Panda/util"
	"github.com/ghettovoice/gosip/sip"
	"github.com/pkg/errors"
	"time"
)

// QueryDeviceSip 查询设备信息SIP
func QueryDeviceSip(device model.Device) {
	xml, err := util.CreateQueryXML(util.DeviceInfoCmdType, device.DeviceId)
	if err != nil {
		return
	}
	request, _ := createMessageRequest(device, contentTypeXML, sip.MESSAGE, xml)
	transferToLog(device.Ip, device.Port)
	log.Info(request)
	tx, err := transmitRequest(request)
	if err != nil {
		log.Error("发送查询通道请求错误")
		return
	}
	resp := getResponse(tx)
	log.Info("收到invite响应:\n", resp)
}

// QueryChannelSip 查询设备通道SIP
func QueryChannelSip(device model.Device) {
	xml, err := util.CreateQueryXML(util.CatalogCmdType, device.DeviceId)
	if err != nil {
		return
	}
	request, _ := createMessageRequest(device, contentTypeXML, sip.MESSAGE, xml)
	transferToLog(device.Ip, device.Port)
	log.Info(request)
	tx, err := transmitRequest(request)
	if err != nil {
		log.Error("发送查询通道请求错误")
		return
	}
	resp := getResponse(tx)
	log.Info("收到invite响应:\n", resp)
}

// SetGuardSip 设防SIP
func SetGuardSip(device model.Device) {
	xml, err := util.CreateGuardXML(util.DeviceControl, device.DeviceId)
	if err != nil {
		return
	}
	_, requestBuilder := createMessageRequest(device, contentTypeXML, sip.MESSAGE, xml)
	requestBuilder.SetContact(newToPort(config.SipOp.Id, config.SipOp.Ip, config.SipOp.Port))
	request, _ := requestBuilder.Build()
	transferToLog(device.Ip, device.Port)
	log.Info(request)
	tx, err := transmitRequest(request)
	if err != nil {
		log.Error("发送预警订阅请求错误")
		return
	}
	resp := getResponse(tx)
	log.Info("收到预警订阅响应:\n", resp)
}

// SubscribeAlarmSip 订阅预警消息SIP
func SubscribeAlarmSip(device model.Device) {
	xml, err := util.CreateSubscribeAlarmXML(util.AlarmCmdType, device.DeviceId)
	if err != nil {
		return
	}
	_, requestBuilder := createMessageRequest(device, contentTypeXML, sip.SUBSCRIBE, xml)
	requestBuilder.SetContact(newToPort(config.SipOp.Id, config.SipOp.Ip, config.SipOp.Port))
	exe := sip.Expires(3600)
	requestBuilder.SetExpires(&exe)
	//测试添加头部Event内容
	requestBuilder.AddHeader(&sip.GenericHeader{
		HeaderName: "Event",
		Contents:   "Alarm",
	})
	request, _ := requestBuilder.Build()
	transferToLog(device.Ip, device.Port)
	log.Info(request)
	tx, err := transmitRequest(request)
	if err != nil {
		log.Error("发送预警订阅请求错误")
		return
	}
	resp := getResponse(tx)
	log.Info("收到预警订阅响应:\n", resp)
}

// 发送I帧 SIP
func IFameSip(device model.Device, channelId string) {
	xml, err := util.CreateIFameXML(channelId)
	if err != nil {
		log.Info("发送IFame请求错误")
	}
	_, requestBuilder := createMessageRequest(device, contentTypeXML, sip.MESSAGE, xml)
	requestBuilder.SetContact(newToPort(config.SipOp.Id, config.SipOp.Ip, config.SipOp.Port))
	exe := sip.Expires(3600)
	requestBuilder.SetExpires(&exe)
	request, _ := requestBuilder.Build()
	transferToLog(device.Ip, device.Port)
	log.Info(request)
	tx, err := transmitRequest(request)
	if err != nil {
		log.Info("发送IFame请求错误")
	}
	response := getResponse(tx)
	log.Info("收到IFame结果:\n", response)
}

// RecordInfoSip 获取历史视频SIP
func RecordInfoSip(device model.Device, channelId string, startTime string, endTime string) (*[]RecordInfo, error) {
	ch := make(chan int, 1)
	defer close(ch)
	xml, sn, err := util.CreateRecordInfoXml(util.RecordInfoCmdType, channelId, startTime, endTime)
	if err != nil {
		return nil, errors.New("获取数据超时")
	}
	_, requestBuilder := createMessageRequest(device, contentTypeXML, sip.MESSAGE, xml)
	requestBuilder.SetContact(newToPort(config.SipOp.Id, config.SipOp.Ip, config.SipOp.Port))
	exe := sip.Expires(3600)
	requestBuilder.SetExpires(&exe)
	request, _ := requestBuilder.Build()
	transferToLog(device.Ip, device.Port)
	log.Info(request)
	tx, err := transmitRequest(request)
	if err != nil {
		log.Error("发送请求历史视频请求错误")
		return nil, errors.New("获取数据超时")
	}
	response := getResponse(tx)
	log.Info("收到历史视频响应:\n", response)
	recordKey := fmt.Sprintf("%s%s", channelId, sn)
	//开始和结束时间
	_recordList.Delete(recordKey)
	_recordList.Store(recordKey, recordList{ch: ch, num: 0, data: []RecordInfo{}})
	tick := time.NewTicker(30 * time.Second)
	select {
	case _, chStatus := <-ch:
		if !chStatus {
			return nil, errors.New("获取数据超时-chan通道已关闭")
		}
		if list, ok := _recordList.Load(recordKey); ok {
			data := list.(recordList)
			recordData := parseRecordData(data.data)
			return &recordData, nil
		}
		return nil, errors.New("获取数据超时")
	case <-tick.C:
		// 30秒未完成返回当前获取到的数据
		if list, ok := _recordList.Load(recordKey); ok {
			data := list.(recordList)
			recordData := parseRecordData(data.data)
			return &recordData, nil
		}
		return nil, errors.New("获取数据超时")
	}
}
