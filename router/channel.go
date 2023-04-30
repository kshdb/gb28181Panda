package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gb28181Panda/config"
	"gb28181Panda/log"
	"gb28181Panda/model"
	"gb28181Panda/sipServer"
	"gb28181Panda/util"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"time"
)

type WebrtcSdp struct {
	Code int    `json:"code"`
	ID   string `json:"id"`
	Sdp  string `json:"sdp"`
	Type string `json:"type"`
}

func ChannelRouter(g *gin.RouterGroup) {
	//设备channel列表
	g.GET("/list/:deviceId", func(c *gin.Context) {
		channel := model.Channel{DeviceId: c.Param("deviceId")}
		channelList, err := channel.ChannelList()
		if err != nil {
			util.GinFRes(c, "查询通道列表失败", nil)
			return
		}
		util.GinSRes(c, "查询通道列表成功", channelList)
	})

	//播放设备+通道的视频
	g.GET("/play/:channelId", func(c *gin.Context) {
		var channel model.Channel
		channel.DeviceId = c.Param("channelId")
		channel, _ = channel.ChannelDetail()
		if channel.DeviceId == "" {
			util.GinFRes(c, "当前通道不存在", nil)
			return
		}
		//channel.ParentId 为device的id
		ssrc := channel.ParentId + "_" + c.Param("channelId")

		var device model.Device
		device.DeviceId = channel.ParentId
		device, _ = device.DeviceDetail()
		if device.DeviceId == "" {
			util.GinFRes(c, "当前设备不存在", nil)
			return
		}
		//先去zlm流媒体查询ssrc是否存在，存在就直接返回播放地址
		mediaStatus, _ := util.GetMediaList(ssrc)
		if mediaStatus {
			streamInfo := model.MustNewStreamInfo(ssrc)
			util.GinSRes(c, "成功", streamInfo)
			//设置sql中当前通道是处于播放状态的
			channel.MediaStatus = "OPEN"
			_ = channel.ChannelUpdate()
			return
		}

		rtpPort, err := util.ZlmOpenRtpServer(ssrc, channel)
		if err != nil {
			log.Info("初始化ZLM接受rtp信息失败")
			util.GinFRes(c, "ZLM初始化失败", nil)
			return
		}
		log.Info("初始化ZLM接受rtp信息成功", rtpPort, ssrc)
		streamInfo, err := sipServer.Play(device, channel, ssrc, rtpPort, sipServer.PlayNowType, "0", "0")
		if err != nil {
			log.Errorf("%+v", err)
			util.GinFRes(c, err.Error(), nil)
			return
		}
		//设置sql中当前通道是处于播放状态的
		channel.MediaStatus = "OPEN"
		_ = channel.ChannelUpdate()
		util.GinSRes(c, "成功", streamInfo)
	})

	//设置通道的流媒体传输方式
	g.GET("/tcp-mode/:channelId", func(c *gin.Context) {
		tcpMode := c.Query("tcpMode")
		if tcpMode == "" {
			util.GinFRes(c, "参数错误", nil)
			return
		}
		var channel model.Channel
		channel.DeviceId = c.Param("channelId")
		channel, _ = channel.ChannelDetail()
		if channel.DeviceId == "" {
			util.GinFRes(c, "当前通道不存在", nil)
			return
		}
		//设置通道流媒体传输方式
		channel.TransportType = tcpMode
		_ = channel.ChannelUpdate()
		util.GinSRes(c, "设置流媒体传输方式成功", nil)
	})

	//播放历史视频
	g.GET("/playback/:channelId", func(c *gin.Context) {
		startTime := c.Query("startTime")
		endTime := c.Query("endTime")
		userId := c.Query("userId")
		if startTime == "" {
			util.GinFRes(c, "开始时间错误", nil)
			return
		}
		if endTime == "" {
			util.GinFRes(c, "结束时间错误", nil)
			return
		}
		var channel model.Channel
		channel.DeviceId = c.Param("channelId")
		channel, _ = channel.ChannelDetail()
		if channel.DeviceId == "" {
			util.GinFRes(c, "当前通道不存在", nil)
			return
		}
		//channel.ParentId 为device的id 区分不同的用户查看历史回放 加上userId，不然公用一个ssrc的话会导致不同用户查看相同通道历史回放时间一样
		ssrc := channel.ParentId + "_" + c.Param("channelId") + "_playback_" + userId

		var device model.Device
		device.DeviceId = channel.ParentId
		device, _ = device.DeviceDetail()
		if device.DeviceId == "" {
			util.GinFRes(c, "当前设备不存在", nil)
			return
		}
		rtpPort, err := util.ZlmOpenRtpServer(ssrc, channel)
		if err != nil {
			log.Info("初始化ZLM接受rtp信息失败")
			util.GinFRes(c, "ZLM初始化失败", nil)
			return
		}
		log.Info("初始化ZLM接受rtp信息成功", rtpPort, ssrc)
		startT, _ := time.ParseInLocation("2006-01-02 15:04:05", startTime, time.Local)
		endT, _ := time.ParseInLocation("2006-01-02 15:04:05", endTime, time.Local)
		//先停止历史视频播放
		_ = sipServer.Stop(device, ssrc)
		//强制I帧
		//sipServer.IFameSip(device, c.Param("channelId"))
		streamInfo, err := sipServer.Play(device, channel, ssrc, rtpPort, sipServer.PlaybackType, startT.Format("2006-01-02T15:04:05"), endT.Format("2006-01-02T15:04:05"))
		if err != nil {
			log.Info(err)
			util.GinFRes(c, "SIP错误", nil)
			return
		}
		util.GinSRes(c, "成功", streamInfo)
	})
	//	历史视频序列
	g.GET("/record/:channelId", func(c *gin.Context) {
		startTime := c.Query("startTime")
		endTime := c.Query("endTime")
		if startTime == "" {
			util.GinFRes(c, "开始时间错误", nil)
			return
		}
		if endTime == "" {
			util.GinFRes(c, "结束时间错误", nil)
			return
		}
		//TODO 比较开始时间和结束时间
		var channel model.Channel
		channel.DeviceId = c.Param("channelId")
		channel, _ = channel.ChannelDetail()
		if channel.DeviceId == "" {
			util.GinFRes(c, "当前通道不存在", nil)
			return
		}
		var device model.Device
		device.DeviceId = channel.ParentId
		device, _ = device.DeviceDetail()
		if device.DeviceId == "" {
			util.GinFRes(c, "当前设备不存在", nil)
			return
		}
		if device.Status == "off" {
			util.GinFRes(c, "当前设备离线", nil)
			return
		}
		startT, _ := time.ParseInLocation("2006-01-02 15:04:05", startTime, time.Local)
		endT, _ := time.ParseInLocation("2006-01-02 15:04:05", endTime, time.Local)

		res, err := sipServer.RecordInfoSip(device, c.Param("channelId"), startT.Format("2006-01-02T15:04:05"), endT.Format("2006-01-02T15:04:05"))
		if err != nil {
			util.GinFRes(c, "获取历史视频失败", nil)
			return
		}
		util.GinSRes(c, "获取历史视频成功", res)
	})
	g.POST("/webrtc/:channelId", func(c *gin.Context) {
		var channel model.Channel
		channel.DeviceId = c.Param("channelId")
		channel, _ = channel.ChannelDetail()
		if channel.DeviceId == "" {
			util.GinFRes(c, "当前通道不存在", nil)
			return
		}
		//channel.ParentId 为device的id
		ssrc := channel.ParentId + "_" + c.Param("channelId")
		var device model.Device
		device.DeviceId = channel.ParentId
		device, _ = device.DeviceDetail()
		if device.DeviceId == "" {
			util.GinFRes(c, "当前设备不存在", nil)
			return
		}
		body := c.Request.Body
		x, _ := io.ReadAll(body)
		formBytesReader := bytes.NewReader(x)
		client := &http.Client{}
		url := fmt.Sprintf("http://%s:%d/index/api/webrtc?app=rtp&stream=%s&type=play&secret=%s", config.ZlmOp.Ip, config.ZlmOp.HttpPort, ssrc, config.ZlmOp.Secret)
		req, err := http.NewRequest("POST", url, formBytesReader)
		if err != nil {
			log.Fatal("生成请求失败！", err)
			util.GinFRes(c, "webrtc失败", nil)
			return
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		response, err := client.Do(req)
		defer response.Body.Close()
		webrtcSdp := WebrtcSdp{}
		responseBody, readErr := io.ReadAll(response.Body)
		if readErr != nil {
			log.Fatal(readErr)
			util.GinFRes(c, "webrtc失败", nil)
			return
		}
		jsonErr := json.Unmarshal(responseBody, &webrtcSdp)
		if jsonErr != nil {
			log.Fatal(jsonErr)
			util.GinFRes(c, "webrtc失败", nil)
			return
		}
		if response.StatusCode == http.StatusOK {
			if webrtcSdp.Code == 0 {
				util.GinSRes(c, "成功", gin.H{
					"sdp":  webrtcSdp.Sdp,
					"type": "answer",
				})
			} else {
				util.GinFRes(c, "webrtc失败，清尝试重新打开视频后再播放webrtc", nil)
			}
		} else {
			util.GinFRes(c, "webrtc失败", nil)
		}
	})

	//停止视频
	g.GET("/stop/:channelId", func(c *gin.Context) {
		var channel model.Channel
		channel.DeviceId = c.Param("channelId")
		channel, err := channel.ChannelDetail()
		if channel.DeviceId == "" {
			util.GinFRes(c, "当前通道不存在", nil)
			return
		}
		if err != nil {
			util.GinFRes(c, "sql错误", nil)
			return
		}
		var device model.Device
		device.DeviceId = channel.ParentId
		device, err = device.DeviceDetail()
		if device.DeviceId == "" {
			util.GinFRes(c, "当前设备不存在", nil)
			return
		}
		if err != nil {
			util.GinFRes(c, "sql错误", nil)
			return
		}
		//channel.ParentId 为device的id
		ssrc := channel.ParentId + "_" + c.Param("channelId")
		//更新sql中当前直播状态字段
		channel.MediaStatus = "CLOSE"
		_ = channel.ChannelUpdate()

		err = sipServer.Stop(device, ssrc)
		if err != nil {
			util.GinFRes(c, "sip停止视频错误", nil)
			return
		}
		util.ZlmCloseRtpServer(ssrc)
		util.GinSRes(c, "关闭视频成功", nil)
	})
}
