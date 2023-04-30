package sipServer

import (
	"fmt"
	"gb28181Panda/config"
	"gb28181Panda/log"
	"gb28181Panda/model"
	"gb28181Panda/storage"
	"gb28181Panda/util"
	"github.com/ghettovoice/gosip/sip"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"net/http"
	"strconv"
	"time"
)

const (
	contentTypeXML = "Application/MANSCDP+xml"
	contentTypeSDP = "APPLICATION/SDP"
)

// 普通的消息组成
func createMessageRequest(device model.Device, contentType string, method sip.RequestMethod, body string) (sip.Request, *sip.RequestBuilder) {
	requestBuilder := sip.NewRequestBuilder()
	devicePort, _ := strconv.Atoi(device.Port)
	requestBuilder.SetFrom(newFromAddress(newParams(map[string]string{"tag": util.RandString(32)})))
	to := newToPort(device.DeviceId, device.Ip, devicePort)
	requestBuilder.SetTo(to)
	requestBuilder.SetRecipient(to.Uri)
	requestBuilder.AddVia(newVia(device.Transport))
	requestContentType := sip.ContentType(contentType)
	requestBuilder.SetContentType(&requestContentType)
	requestBuilder.SetMethod(method)
	userAgent := sip.UserAgentHeader(fmt.Sprintf("%s", config.SipOp.UserAgent))
	requestBuilder.SetUserAgent(&userAgent)
	requestBuilder.SetBody(body)
	ceq, _ := storage.GetCeq("ceq")
	requestBuilder.SetSeqNo(cast.ToUint(ceq))
	request, _ := requestBuilder.Build()
	return request, requestBuilder
}

// 视频播放和点播
func createVideoMessageRequest(device model.Device, contentType string, method sip.RequestMethod, channelId string, ssrc string, body string) (sip.Request, *sip.RequestBuilder) {
	requestBuilder := sip.NewRequestBuilder()
	devicePort, _ := strconv.Atoi(device.Port)
	to := newToPort(channelId, device.Ip, devicePort)
	requestBuilder.SetMethod(method)
	requestBuilder.SetFrom(newFromAddress(newParams(map[string]string{"tag": util.RandString(32)})))
	requestBuilder.SetTo(to)
	port := sip.Port(devicePort)
	sipUri := &sip.SipUri{
		FUser: sip.String{Str: channelId},
		FHost: to.Uri.Host(),
		FPort: &port,
	}
	requestBuilder.SetRecipient(sipUri)
	requestBuilder.AddVia(newVia(device.Transport))
	requestBuilder.SetContact(newToPort(config.SipOp.Id, config.SipOp.Ip, config.SipOp.Port))
	requestContentType := sip.ContentType(contentType)
	requestBuilder.SetContentType(&requestContentType)
	requestBuilder.SetBody(body)
	userAgent := sip.UserAgentHeader(config.SipOp.UserAgent)
	requestBuilder.SetUserAgent(&userAgent)
	ceq, _ := storage.GetCeq("ceq")
	requestBuilder.SetSeqNo(cast.ToUint(ceq))
	callID := sip.CallID(fmt.Sprintf("%s", util.RandString(22)))
	requestBuilder.SetCallID(&callID)
	header := sip.GenericHeader{
		HeaderName: "Subject",
		Contents:   fmt.Sprintf("%s:%s,%s:%d", channelId, ssrc, config.SipOp.Id, 0),
	}
	requestBuilder.AddHeader(&header)
	request, _ := requestBuilder.Build()
	return request, requestBuilder
}

// 向上级联注册
func createCascadeRequest(method sip.RequestMethod, cascade model.Cascade) sip.Request {
	//创建请求
	requestBuilder := sip.NewRequestBuilder()
	to := newToPort(cascade.Id, cascade.Ip, cascade.Port)
	requestBuilder.SetMethod(method)
	requestBuilder.SetFrom(newFromAddress(newParams(map[string]string{"tag": util.RandString(32)})))
	requestBuilder.SetTo(to)
	port := sip.Port(cascade.Port)
	sipUri := &sip.SipUri{
		FUser: sip.String{Str: cascade.Id},
		FHost: to.Uri.Host(),
		FPort: &port,
	}
	requestBuilder.SetRecipient(sipUri)
	requestBuilder.AddVia(newVia(cascade.Transport))
	requestBuilder.SetContact(newToPort(config.SipOp.Id, config.SipOp.Ip, config.SipOp.Port))
	userAgent := sip.UserAgentHeader(config.SipOp.UserAgent)
	requestBuilder.SetUserAgent(&userAgent)
	ceq, _ := storage.GetCeq("ceq")
	requestBuilder.SetSeqNo(cast.ToUint(ceq))
	callID := sip.CallID(fmt.Sprintf("%s", util.RandString(22)))
	requestBuilder.SetCallID(&callID)
	exe := sip.Expires(3600)
	requestBuilder.SetExpires(&exe)
	request, err := requestBuilder.Build()
	if err != nil {
		log.Error("发生错误:", err)
	}
	return request
}

// get sip tx info by sip request and response
func getRequestTxField(request sip.Request, response sip.Response) (callId, fromTag, toTag, viaBranch string, err error) {
	callID, ok := request.CallID()
	if !ok {
		return "", "", "", "", errors.New("获取CallId失败")
	}

	fromHeader, ok := request.From()
	if !ok {
		return "", "", "", "", errors.New("获取fromHeader失败")
	}
	ft, ok := fromHeader.Params.Get("tag")
	if !ok {
		return "", "", "", "", errors.New("获取fromTag失败")
	}

	toHeader, ok := response.To()
	if !ok {
		return "", "", "", "", errors.New("获取toHeader失败")
	}
	tg, okTag := toHeader.Params.Get("tag")
	if !okTag {
		return "", "", "", "", errors.New("获取toTag失败")
	}

	viaHop, ok := request.ViaHop()
	if !ok {
		return "", "", "", "", errors.New("获取viaHop失败")
	}

	branch, ok := viaHop.Params.Get("branch")
	if !ok {
		return "", "", "", "", errors.New("获取branch失败")
	}

	callId = callID.Value()
	fromTag = ft.String()

	if !okTag {
		toTag = "unkonw to tag"
	} else {
		toTag = tg.String()
	}
	viaBranch = branch.String()
	return
}

// transmitRequest 发送sip请求
func transmitRequest(req sip.Request) (sip.ClientTransaction, error) {
	transaction, err := SipServer.Request(req)
	return transaction, err
}

// getResponse 获取sip的响应结果
func getResponse(tx sip.ClientTransaction) sip.Response {
	timer := time.NewTimer(5 * time.Second)
	for {
		select {
		case resp := <-tx.Responses():
			if resp.StatusCode() == sip.StatusCode(http.StatusContinue) ||
				resp.StatusCode() == sip.StatusCode(http.StatusSwitchingProtocols) {
				continue
			}
			return resp
		case <-timer.C:
			log.Error("获取响应超时")
			return nil
		}
	}
}
func newFromAddress(params sip.Params) *sip.Address {
	portFrom := sip.Port(config.SipOp.Port)
	return &sip.Address{
		Uri: &sip.SipUri{
			FUser: sip.String{Str: config.SipOp.Id},
			FHost: config.SipOp.Realm,
			FPort: &portFrom,
		},
		Params: params,
	}
}

func newParams(m map[string]string) sip.Params {
	params := sip.NewParams()
	for k, v := range m {
		params.Add(k, sip.String{Str: v})
	}
	return params
}

func newVia(transport string) *sip.ViaHop {
	p := sip.Port(config.SipOp.Port)
	params := newParams(map[string]string{
		"branch": fmt.Sprintf("%s%d", "z9hG4bK", time.Now().UnixMilli()),
	})
	return &sip.ViaHop{
		ProtocolName:    "SIP",
		ProtocolVersion: "2.0",
		Transport:       transport,
		Host:            config.SipOp.Ip,
		Port:            &p,
		Params:          params,
	}
}

func newToPort(channelId, host string, port int) *sip.Address {
	toPort := sip.Port(port)
	return &sip.Address{
		Uri: &sip.SipUri{
			FUser: sip.String{Str: channelId},
			FHost: host,
			FPort: &toPort,
		},
	}
}
