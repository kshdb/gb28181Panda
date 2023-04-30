package model

import (
	"fmt"
	"gb28181Panda/config"
	"strings"
)

// StreamInfo 流信息
type StreamInfo struct {
	MediaServerId string `json:"mediaServerId"`
	App           string `json:"app"`
	Ip            string `json:"ip"`
	DeviceID      string `json:"deviceID"`
	ChannelId     string `json:"channelId"`
	Stream        string `json:"stream"`
	Rtmp          string `json:"rtmp"`
	Rtsp          string `json:"rtsp"`
	Flv           string `json:"flv"`
	HttpsFlv      string `json:"httpsFlv"`
	WsFlv         string `json:"wsFlv"`
	//WSFly         string `json:"WSFly"`
	//WSSFly        string `json:"WSSFly"`
	Fmp4 string `json:"fmp4"`
	//HttpsFmp4     string `json:"httpsFmp4"`
	//WSFmpt4       string `json:"WSFmpt4"`
	Hls string `json:"hls"`
	//HttpsHls      string `json:"httpsHls"`
	//WsHls         string `json:"wsHls"`
	Ts string `json:"ts"`
	//HttpsTs       string `json:"httpsTs"`
	//WebsocketTs   string `json:"websocketTs"`
	Ssrc string `json:"ssrc"`
}

const (
	rtsp  = "rtsp://%s:%d/rtp/%s"
	rtmp  = "rtmp://%s:%d/rtp/%s"
	wsFlv = "ws://%s:%d/rtp/%s.live.flv"
	http  = "http://%s:%d/rtp/%s/hls.m3u8"
	flv   = "http://%s:%d/rtp/%s.live.flv"
	fmp4  = "http://%s:%d/rtp/%s.llive.mp4"
	ts    = "http://%s:%d/rtp/%s.llive.ts"
)

func MustNewStreamInfo(ssrc string) StreamInfo {
	index := strings.Index(ssrc, "_")
	deviceId := ssrc[0:index]
	channelID := ssrc[index+1:]
	mediaIp := config.ZlmOp.Ip
	httpPort := config.ZlmOp.HttpPort
	rtmpPort := config.ZlmOp.RtmpPort
	rtspPort := config.ZlmOp.RtspPort
	return StreamInfo{
		App:       "rtp",
		DeviceID:  deviceId,
		ChannelId: channelID,
		Stream:    ssrc,
		Rtmp:      fmt.Sprintf(rtmp, mediaIp, rtmpPort, ssrc),
		Rtsp:      fmt.Sprintf(rtsp, mediaIp, rtspPort, ssrc),
		Hls:       fmt.Sprintf(http, mediaIp, httpPort, ssrc),
		Flv:       fmt.Sprintf(flv, mediaIp, httpPort, ssrc),
		Fmp4:      fmt.Sprintf(fmp4, mediaIp, httpPort, ssrc),
		Ts:        fmt.Sprintf(ts, mediaIp, httpPort, ssrc),
		WsFlv:     fmt.Sprintf(wsFlv, mediaIp, httpPort, ssrc),
	}
}
