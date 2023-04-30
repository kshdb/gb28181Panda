package router

import (
	"fmt"
	"gb28181Panda/log"
	"gb28181Panda/model"
	"gb28181Panda/sipServer"
	"gb28181Panda/util"
	"github.com/gin-gonic/gin"
	"strings"
)

// 无人观看
type OnStreamNoneReader struct {
	MediaServerId string `json:"mediaServerId,omitempty"`
	App           string `json:"app,omitempty"`
	Schema        string `json:"schema,omitempty"`
	Stream        string `json:"stream,omitempty"`
	Vhost         string `json:"vhost,omitempty"`
}

// HookReply hook事件默认回复
type HookReply struct {
	// 0代表允许，其他均为不允许
	Code int `json:"code,omitempty"`

	// 当code不为0时，msg字段应给出相应提示
	Msg string `json:"msg,omitempty"`
}

type OnStreamNoneReaderReply struct {
	// 固定返回0
	Code int `json:"code,omitempty"`
	// 是否关闭该流，包括推流和拉流
	Close bool `json:"close,omitempty"`
}

const (
	ParseParamFail = iota + 1
)
const (
	SuccessMsg        = "success"
	ParseParamFailMsg = "parser on_play_hook interface param fail, auth fail and not allow play"
)

// OnPlayHookParam 播放器鉴权hook事件
type OnPlayHookParam struct {
	// 流应用名
	App string `json:"app,omitempty" `

	// TCP链接唯一ID
	Id string `json:"id,omitempty"`

	// 播放器ip
	Ip string `json:"ip,omitempty"`

	// 播放url参数
	Params string `json:"params,omitempty"`

	// 播放器端口号
	Port int `json:"port,omitempty"`

	// 播放的协议，可能是rtsp、rtmp、http
	Schema string `json:"schema,omitempty"`

	// 流ID
	Stream string `json:"stream,omitempty"`

	// 流虚拟主机
	Vhost string `json:"vhost,omitempty"`

	// 服务器id,通过配置文件设置
	MediaServerId string `json:"mediaServerId,omitempty"`
}
type OnPublishHookReply struct {
	HookReply
	// 是否转换成hls协议
	EnableHls bool `json:"enable_hls,omitempty"`
	// 是否允许mp4录制
	EnableMp4 bool `json:"enable_mp4,omitempty"`
	// 是否转rtsp协议
	EnableRtsp bool `json:"enable_rtsp,omitempty"`
	// 是否转rtmp/flv协议
	EnableRtmp bool `json:"enable_rtmp,omitempty"`
	// 是否转http-ts/ws-ts协议
	EnableTs bool `json:"enable_ts,omitempty"`
	// 是否转http-fmp4/ws-fmp4协议
	EnableFmp4 bool `json:"enable_fmp4,omitempty"`
	// 转协议时是否开启音频
	EnableAudio bool `json:"enable_audio,omitempty"`
	// 转协议时，无音频是否添加静音aac音频
	AddMuteAudio bool `json:"add_mute_audio,omitempty"`
	// mp4录制文件保存根目录，置空使用默认
	Mp4SavePath string `json:"mp4_save_path,omitempty"`
	// mp4录制切片大小，单位秒
	Mp4MaxSecond int `json:"mp4_max_second,omitempty"`
	// hls文件保存保存根目录，置空使用默认
	HlsSavePath string `json:"hls_save_path,omitempty"`
	// 	断连续推延时，单位毫秒，置空使用配置文件默认值
	ContinuePushMs uint32 `json:"continue_push_ms,omitempty"`
	// MP4录制是否当作观看者参与播放人数计数
	Mp4AsPlayer bool `json:"mp4_as_player,omitempty"`
	// 该流是否开启时间戳覆盖
	ModifyStamp bool `json:"modify_stamp,omitempty"`
}

func NewOnPublishDefaultReply() OnPublishHookReply {
	return OnPublishHookReply{
		HookReply: HookReply{
			Code: 0,
			Msg:  SuccessMsg,
		},
		AddMuteAudio:   true,
		ContinuePushMs: 10000,
		EnableAudio:    true,
		EnableFmp4:     true,
		EnableHls:      true,
		EnableMp4:      false,
		EnableRtmp:     true,
		EnableRtsp:     true,
		EnableTs:       true,
		HlsSavePath:    "/hls_save_path/",
		ModifyStamp:    false,
		Mp4AsPlayer:    false,
		Mp4MaxSecond:   3600,
		Mp4SavePath:    "/mp4_save_path/",
	}
}
func HookRouter(g *gin.RouterGroup) {
	g.POST("on_stream_none_reader", func(c *gin.Context) {
		hookParam := OnStreamNoneReader{}
		if err := c.ShouldBindJSON(&hookParam); err != nil {
			log.Error(err)
			c.JSON(200, HookReply{
				Code: ParseParamFail,
				Msg:  ParseParamFailMsg,
			})
			return
		}
		log.Info("收到流无人观看事件,stream_id:", hookParam.Stream, "media_server_id:", hookParam.MediaServerId)
		deviceChannel := strings.Split(hookParam.Stream, "_")
		//关闭监控的视频流推送
		var device model.Device
		device.DeviceId = deviceChannel[0]
		device, _ = device.DeviceDetail()
		if device.DeviceId == "" {
			util.GinFRes(c, "当前设备不存在", nil)
			return
		}
		var channel model.Channel
		channel.DeviceId = deviceChannel[1]
		channel, _ = channel.ChannelDetail()
		if channel.DeviceId == "" {
			util.GinFRes(c, "当前通道不存在", nil)
			return
		}
		channel.MediaStatus = "CLOSE"
		_ = channel.ChannelUpdate()
		ssrc := hookParam.Stream
		_ = sipServer.Stop(device, ssrc)
		c.JSON(200, OnStreamNoneReaderReply{
			Code:  0,
			Close: true,
		})
	})
	g.POST("on_stream_changed", func(c *gin.Context) {
		log.Info("on_stream_changed")
		c.JSON(200, HookReply{
			Code: 0,
			Msg:  "success",
		})
	})
	g.POST("on_server_keepalive", func(c *gin.Context) {
		log.Info("on_server_keepalive")
		c.JSON(200, HookReply{
			Code: 0,
			Msg:  "success",
		})
	})
	g.POST("on_play", func(c *gin.Context) {
		log.Info("on_play")
		hookParam := OnPlayHookParam{}
		if err := c.ShouldBindJSON(&hookParam); err != nil {
			log.Error(err)
			c.JSON(200, HookReply{
				Code: ParseParamFail,
				Msg:  ParseParamFailMsg,
			})
			return
		}
		fmt.Println("hookParam", hookParam)
		c.JSON(200, HookReply{
			Code: 0,
			Msg:  "success",
		})
	})
	g.POST("on_publish", func(c *gin.Context) {
		log.Info("on_publish")
		c.JSON(200, NewOnPublishDefaultReply())
	})
}
