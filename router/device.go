package router

import (
	"gb28181Panda/log"
	"gb28181Panda/model"
	"gb28181Panda/sipServer"
	"gb28181Panda/util"
	"github.com/gin-gonic/gin"
	"go/types"
	"strconv"
	"time"
)

func DeviceRouter(g *gin.RouterGroup) {
	//设备device列表
	g.GET("/list", func(c *gin.Context) {
		var device model.Device
		deviceList, err := device.DeviceList()
		if err != nil {
			util.GinFRes(c, "查询设备列表失败", nil)
			return
		}
		for key, value := range deviceList {
			formatTime, err := time.ParseInLocation(util.TIME_LAYOUT_With_SPACE, value.KeepaliveAt, time.Local)
			if err != nil {
				log.Error("Format Time Error")
				return
			}
			if time.Now().Unix()-formatTime.Unix() > 60 && deviceList[key].Status == "ON" {
				deviceList[key].Status = "OFF"
				value.Status = "OFF"
				_ = value.DeviceUpdate()
			}
		}
		util.GinSRes(c, "查询设备列表成功", deviceList)

	})
	//设备device-channel树
	g.GET("/tree", func(c *gin.Context) {
		var deviceTree model.DeviceTree
		deviceData, err := deviceTree.DeviceTreeData()
		if err != nil {
			util.GinFRes(c, "查询设备列表失败", nil)
			return
		}
		for key, value := range deviceData {
			formatTime, err := time.ParseInLocation(util.TIME_LAYOUT_With_SPACE, value.KeepaliveAt, time.Local)
			if err != nil {
				log.Error("格式化时间失败")
				return
			}
			if time.Now().Unix()-formatTime.Unix() > 60 && deviceData[key].Status == "ON" {
				deviceData[key].Status = "OFF"
			}
		}
		util.GinSRes(c, "查询设备树成功", deviceData)

	})
	//更新通道信息
	g.GET("/channel/sync/:deviceId", func(c *gin.Context) {
		var device model.Device
		device.DeviceId = c.Param("deviceId")
		device, _ = device.DeviceDetail()
		if device.DeviceId == "" {
			util.GinFRes(c, "当前设备不存在", nil)
			return
		}
		sipServer.QueryChannelSip(device)
		util.GinSRes(c, "更新通道成功", nil)
	})
	//	订阅Alarm消息
	g.GET("/subscribe/:deviceId", func(c *gin.Context) {
		var device model.Device
		device.DeviceId = c.Param("deviceId")
		device, _ = device.DeviceDetail()
		if device.DeviceId == "" {
			util.GinFRes(c, "当前设备不存在", nil)
			return
		}
		sipServer.SetGuardSip(device)
		sipServer.SubscribeAlarmSip(device)
		util.GinSRes(c, "预警订阅成功", types.Nil{})
	})

	//删除设备
	g.DELETE("/delete/:deviceId", func(c *gin.Context) {
		var device model.Device
		device.DeviceId = c.Param("deviceId")
		device, _ = device.DeviceDetail()
		if device.DeviceId == "" {
			util.GinFRes(c, "当前设备不存在", nil)
			return
		}
		if device.Status == "ON" {
			util.GinFRes(c, "当前设备在线，不可删除", nil)
			return
		}
		//先删除device
		_ = device.DeviceDelete()
		var channel model.Channel
		channel.ParentId = c.Param("deviceId")
		//再删除channel
		_ = channel.ChannelDeleteWithParentId()
		util.GinSRes(c, "删除设备成功", nil)
	})
	/*  次处无设备可联调 个人理解 ptz控制的应该是deviceId而不是channelId
	// 设备id
		DeviceId string `json:"deviceId,omitempty"`
		// 控制的命令，取值为:left、right、down、up、downright、downleft、upright、upleft、zoomin、zoomout
		Command string `json:"command,omitempty"`
		// 水平方向移动速度，取值:0-255
		HorizonSpeed int `json:"horizonSpeed,omitempty"`
		// 垂直方向移动速度，取值:0-255
		VerticalSpeed int `json:"verticalSpeed,omitempty"`
		// 变倍控制速度，取值:0-255
		ZoomSpeed int `json:"zoomSpeed,omitempty"`
	*/
	g.GET("/ptz/:deviceId/:command/:horizonSpeed/:verticalSpeed/:zoomSpeed", func(c *gin.Context) {
		var device model.Device
		device.DeviceId = c.Param("deviceId")
		device, _ = device.DeviceDetail()
		if device.DeviceId == "" {
			util.GinFRes(c, "当前设备不存在", nil)
			return
		}
		horizonSpeed, _ := strconv.Atoi(c.Param("horizonSpeed"))
		verticalSpeed, _ := strconv.Atoi(c.Param("verticalSpeed"))
		zoomSpeed, _ := strconv.Atoi(c.Param("zoomSpeed"))
		err := sipServer.ControlPTZ(device, c.Param("command"), horizonSpeed, verticalSpeed, zoomSpeed)
		if err != nil {
			util.GinFRes(c, "ptz-err", nil)
			return
		}
		util.GinSRes(c, "ptz-ok", nil)
	})
}
