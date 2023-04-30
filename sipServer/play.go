package sipServer

import (
	"gb28181Panda/log"
	"gb28181Panda/model"
	"github.com/ghettovoice/gosip/sip"
	"github.com/pkg/errors"
	"net/http"
)

// Play 播放实时视频或历史回放点播
func Play(device model.Device, channel model.Channel, ssrc string, rtpPort int, playType, startT, endT string) (model.StreamInfo, error) {
	body := createVideoSdpInfo(channel, ssrc, rtpPort, playType, startT, endT)
	request, _ := createVideoMessageRequest(device, contentTypeSDP, sip.INVITE, channel.DeviceId, ssrc, body)
	transferToLog(device.Ip, device.Port)
	log.Info("Play-Request:\n", request)
	tx, err := transmitRequest(request)
	if err != nil {
		log.Error("发送视频播放请求错误")
		return model.StreamInfo{}, errors.New("发送视频播放请求错误")
	}
	resp := getResponse(tx)
	//TODO 这块的错误处理后续优化
	if resp.StatusCode() != sip.StatusCode(http.StatusOK) {
		return model.StreamInfo{}, errors.New("当前时间错误")
	}
	log.Info("收到invite响应:\n", resp)
	if resp == nil {
		log.Error("获取响应超时")
		return model.StreamInfo{}, errors.New("获取响应超时")
	}
	ackRequest := sip.NewAckRequest("", request, resp, "", nil)
	ackRequest.SetRecipient(request.Recipient())
	ackRequest.AppendHeader(&sip.ContactHeader{
		Address: request.Recipient(),
		Params:  nil,
	})
	transferToLog(device.Ip, device.Port)
	log.Info("发送ack确认:\n", ackRequest)
	err = SipServer.Send(ackRequest)
	if err != nil {
		log.Errorf("发送ack失败", err)
		return model.StreamInfo{}, errors.New("发送invite请求错误")
	}
	callId, fromTag, toTag, branch, err := getRequestTxField(request, resp)
	saveStreamSession(device.DeviceId, channel.DeviceId, ssrc, callId, fromTag, toTag, branch)
	if err != nil {
		return model.StreamInfo{}, err
	}
	info := model.MustNewStreamInfo(ssrc)
	return info, nil
}
