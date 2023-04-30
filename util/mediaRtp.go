package util

import (
	"encoding/json"
	"fmt"
	"gb28181Panda/config"
	"gb28181Panda/log"
	"gb28181Panda/model"
	"github.com/pkg/errors"
)

type C struct {
	Code int `json:"code"`
}

type Message struct {
	Msg string `json:"msg"`
}

type CodeMessage struct {
	C
	Message
}

type GetRtpInfoResp struct {
	C
	Message
	Exist     bool   `json:"exist,omitempty"`
	PeerIp    string `json:"peer_ip,omitempty"`
	PeerPort  int    `json:"peer_port,omitempty"`
	LocalIp   string `json:"local_ip,omitempty"`
	LocalPort int    `json:"local_port,omitempty"`
}

type CreateRtpServerResp struct {
	C
	Message
	Port int `json:"port"`
}

type MediaInfoResp struct {
	Code int `json:"code"`
	Data []struct {
		AliveSecond    int    `json:"aliveSecond"`
		App            string `json:"app"`
		BytesSpeed     int    `json:"bytesSpeed"`
		CreateStamp    int    `json:"createStamp"`
		IsRecordingHLS bool   `json:"isRecordingHLS"`
		IsRecordingMP4 bool   `json:"isRecordingMP4"`
		OriginSock     struct {
			Identifier string `json:"identifier"`
			LocalIP    string `json:"local_ip"`
			LocalPort  int    `json:"local_port"`
			PeerIP     string `json:"peer_ip"`
			PeerPort   int    `json:"peer_port"`
		} `json:"originSock"`
		OriginType       int    `json:"originType"`
		OriginTypeStr    string `json:"originTypeStr"`
		OriginURL        string `json:"originUrl"`
		ReaderCount      int    `json:"readerCount"`
		Schema           string `json:"schema"`
		Stream           string `json:"stream"`
		TotalReaderCount int    `json:"totalReaderCount"`
		Tracks           []struct {
			Channels      int     `json:"channels,omitempty"`
			CodecID       int     `json:"codec_id"`
			CodecIDName   string  `json:"codec_id_name"`
			CodecType     int     `json:"codec_type"`
			Frames        int     `json:"frames"`
			Ready         bool    `json:"ready"`
			SampleBit     int     `json:"sample_bit,omitempty"`
			SampleRate    int     `json:"sample_rate,omitempty"`
			Fps           float64 `json:"fps,omitempty"`
			GopIntervalMs int     `json:"gop_interval_ms,omitempty"`
			GopSize       int     `json:"gop_size,omitempty"`
			Height        int     `json:"height,omitempty"`
			KeyFrames     int     `json:"key_frames,omitempty"`
			Width         int     `json:"width,omitempty"`
		} `json:"tracks"`
		Vhost string `json:"vhost"`
	} `json:"data"`
}

// 参考https://github.com/ZLMediaKit/ZLMediaKit/wiki/MediaServer%E6%94%AF%E6%8C%81%E7%9A%84HTTP-API#24indexapiopenrtpserver
func ZlmOpenRtpServer(stream string, channel model.Channel) (rtpPort int, err error) {
	tcpMode := 0
	//0 udp 模式，1 tcp 被动模式, 2 tcp 主动模式。
	if channel.TransportType == "TCPACTIVE" {
		tcpMode = 2
	} else if channel.TransportType == "TCPPASSIVE" {
		tcpMode = 1
	}
	url := fmt.Sprintf("http://%s:%d/index/api/openRtpServer?port=0&tcp_mode=%d&stream_id=%s&secret=%s", config.ZlmOp.Ip, config.ZlmOp.HttpPort, tcpMode, stream, config.ZlmOp.Secret)
	params := map[string]interface{}{}
	body, err := SendPost(url, params)
	if err != nil {
		return 0, errors.WithMessage(err, "create rtp server fail")
	}

	resp := CreateRtpServerResp{}
	err = json.Unmarshal([]byte(body), &resp)
	if resp.Code == -300 {
		ZlmCloseRtpServer(stream)
		return ZlmOpenRtpServer(stream, channel)
	}
	if err != nil {
		return 0, errors.WithMessage(err, "unmarshal data to struct fail")
	}

	if resp.Code != 0 {
		return 0, errors.New(resp.Msg)
	}

	rtpPort = resp.Port
	return
}

// ZlmCloseRtpServer 关闭zlm中的stream数据
func ZlmCloseRtpServer(stream string) {
	url := fmt.Sprintf("http://%s:%d/index/api/closeRtpServer?stream_id=%s&secret=%s", config.ZlmOp.Ip, config.ZlmOp.HttpPort, stream, config.ZlmOp.Secret)
	params := map[string]interface{}{}
	body, err := SendPost(url, params)
	if err != nil {
		log.Error(err, "close rtp server fail")
	}

	resp := CreateRtpServerResp{}
	err = json.Unmarshal([]byte(body), &resp)
	if err != nil {
		log.Error(err, "unmarshal data to struct fail")
	}

	if resp.Code != 0 {
		log.Error("ZLM内部错误")
	}
}

func GetMediaList(stream string) (status bool, err error) {
	url := fmt.Sprintf("http://%s:%d/index/api/getMediaList?stream=%s&secret=%s&app=rtp", config.ZlmOp.Ip, config.ZlmOp.HttpPort, stream, config.ZlmOp.Secret)
	params := map[string]interface{}{}
	body, err := SendPost(url, params)
	if err != nil {
		return false, errors.WithMessage(err, "getMediaList fail")
	}

	resp := MediaInfoResp{}
	err = json.Unmarshal([]byte(body), &resp)

	if err != nil {
		log.Info("ZLM接口出错了")
		return false, errors.WithMessage(err, "unmarshal data to struct fail")
	}

	if resp.Code != 0 {
		log.Info("ZLM接口出错了")
		return false, errors.New("ZLM接口出错了")
	}

	if len(resp.Data) == 0 {
		log.Info("不存在")
		return false, errors.New("不存在")
	}
	return true, nil
}
