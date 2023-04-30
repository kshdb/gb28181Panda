package sipServer

import (
	"fmt"
	"gb28181Panda/config"
	"gb28181Panda/model"
	sdp "github.com/panjjo/gosdp"
	"net"
	"time"
)

var (
	PlayNowType  = "Play"
	PlaybackType = "Playback"
)

// createVideoSdpInfo 告知设备接受流媒体信息的相关信息  目前只有udp传输方式
func createVideoSdpInfo(channel model.Channel, ssrc string, rtpPort int, playType, startT, endT string) string {
	mediaIp := fmt.Sprintf("%s", config.ZlmOp.Ip)
	//对应的是sdp的[o]参数
	origin := sdp.Origin{
		Username:       channel.DeviceId,
		SessionID:      0,
		SessionVersion: 0,
		NetworkType:    "IN",
		AddressType:    "IP4",
		Address:        mediaIp,
	}
	//UDP
	protocol := "RTP/AVP"
	//TCP被动或者TCP主动
	if channel.TransportType != "UDP" {
		protocol = "TCP/RTP/AVP"
	}
	video := sdp.Media{
		Description: sdp.MediaDescription{
			Type:     "video",
			Port:     rtpPort,
			Protocol: protocol,
			Formats:  []string{"96", "97", "98", "99"},
		},
		Connection: sdp.ConnectionData{
			NetworkType: "IN",
			AddressType: "IP4",
			IP:          net.ParseIP(mediaIp),
			TTL:         0,
		},
	}
	video.AddAttribute("recvonly")
	video.AddAttribute("rtpmap", "96", "PS/90000")
	video.AddAttribute("rtpmap", "98", "H264/90000")
	video.AddAttribute("rtpmap", "97", "MPEG4/90000")
	video.AddAttribute("rtpmap", "99", "H265/90000")
	//TCP主动
	if channel.TransportType != "TCPACTIVE" {
		video.AddAttribute("setup", "active")
		video.AddAttribute("connection", "new")
	}
	//TCP被动
	if channel.TransportType != "TCPPASSIVE" {
		video.AddAttribute("setup", "passive")
		video.AddAttribute("connection", "new")
	}
	timeing := []sdp.Timing{
		{
			Start: time.Time{},
			End:   time.Time{},
		},
	}
	if playType == "Playback" {
		sT, _ := time.ParseInLocation("2006-01-02T15:04:05", startT, time.Local)
		eT, _ := time.ParseInLocation("2006-01-02T15:04:05", endT, time.Local)
		timeing = []sdp.Timing{sdp.Timing{Start: sT, End: eT}}
	}
	msg := sdp.Message{
		Version: 0,
		Origin:  origin,
		Name:    playType,
		Medias:  sdp.Medias{video},
		Timing:  timeing,
		SSRC:    ssrc,
	}
	session := msg.Append(sdp.Session{})
	bytes := session.AppendTo([]byte{})
	return string(bytes)
}
