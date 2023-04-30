package sipServer

import (
	"fmt"
	"gb28181Panda/config"
	"gb28181Panda/log"
	"gb28181Panda/model"
	"gb28181Panda/storage"
	"github.com/ghettovoice/gosip/sip"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"strconv"
)

// Stop 停止播放视频
func Stop(device model.Device, ssrc string) error {
	// 从Reds读取sip相关session用户关系监控推送视频流
	txInfo, err := GetTx(ssrc)
	if err != nil {
		return errors.New("Redis中无视频流信息")
	}
	key := fmt.Sprintf("%s:%s", "Media:Stream:Transaction", ssrc)
	_ = storage.Del(key)
	requestBuilder := sip.NewRequestBuilder()
	devicePort, _ := strconv.Atoi(device.Port)
	//to := newToPort(d.DeviceId, d.Ip, devicePort)
	to := newToPort(txInfo.ChannelId, device.Ip, devicePort)
	to.Params = newParams(map[string]string{"tag": txInfo.ToTag})
	requestBuilder.SetFrom(newFromAddress(newParams(map[string]string{"tag": txInfo.FromTag})))
	requestBuilder.SetMethod(sip.BYE)
	requestBuilder.SetTo(to)
	requestBuilder.SetRecipient(to.Uri)
	via := newVia(device.Transport)
	requestBuilder.AddVia(via)
	requestBuilder.SetContact(newToPort(config.SipOp.Id, config.SipOp.Ip, config.SipOp.Port))
	callId := sip.CallID(txInfo.CallId)
	requestBuilder.SetCallID(&callId)
	userAgent := sip.UserAgentHeader(config.SipOp.UserAgent)
	requestBuilder.SetUserAgent(&userAgent)
	ceq, _ := storage.GetCeq("ceq")
	requestBuilder.SetSeqNo(cast.ToUint(ceq))
	req, _ := requestBuilder.Build()
	transferToLog(device.Ip, device.Port)
	log.Info("发送停止播放视频请求:\n", req)
	tx, err := transmitRequest(req)
	if err != nil {
		log.Error("发送停止视频播放请求错误")
		return errors.New("发送停止视频播放请求错误")
	}
	resp := getResponse(tx)
	transferFromLog(device.Ip, device.Port)
	log.Info("收到停止播放视频响应:\n", resp)
	if resp == nil {
		log.Error("获取响应超时")
		return errors.New("获取响应超时")
	}
	return nil
}
