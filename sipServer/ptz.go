package sipServer

import (
	"fmt"
	"gb28181Panda/log"
	"gb28181Panda/model"
	"gb28181Panda/util"
	"github.com/ghettovoice/gosip/sip"
	"github.com/pkg/errors"
	"strings"
)

// ControlPTZ 设备ptz控制指令
func ControlPTZ(device model.Device, command string, horizonSpeed, verticalSpeed, zoomSpeed int) error {
	cmdStr, err := createPTZCode(command, horizonSpeed, verticalSpeed, zoomSpeed)
	if err != nil {
		log.Error(err)
		return err
	}
	xml, err := util.CreateControlXml(util.DeviceControl, device.DeviceId, util.WithPTZCmd(cmdStr))
	if err != nil {
		log.Error(err)
		return err
	}
	request, _ := createMessageRequest(device, "Application/MANSCDP+xml", sip.MESSAGE, xml)
	transferToLog(device.Ip, device.Port)
	log.Info("Ptz-Request", request)
	tx, err := transmitRequest(request)
	if err != nil {
		log.Error("Ptz发送失败")
		return errors.New("Ptz发送失败")
	}
	resp := getResponse(tx)
	log.Info("收到invite响应:\n", resp)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

// 创建PTZ指令
// 根据gb28181协议的标准，前端指令中一共包含4个字节
func createPTZCode(command string, horizonSpeed, verticalSpeed, zoomSpeed int) (string, error) {
	var ptz strings.Builder
	// gb28181协议中控制指令中的前三个字节
	// 字节1是A5，字节2是组合码，高4位由版本信息组成，版本信息为0H；低四位是校验位，校验位=(字节1的高4位+字节1的低四位+字节2的高四位) % 16
	// 所以校验码 = (0xa + 0x5 + 0) % 16 = (1010 + 0101 + 0) % 16 = 15 % 16 = 15；十进制数15转十六进制= F
	// 所以字节2 = 0F
	// 字节3是地址的低8位，这里直接设置为01
	ptz.WriteString("A50F01")
	var cmd int

	// 指令码以一个字节来表示
	// 0000 0000，高位的前两个bit不做表示
	// 所以有作用的也就是后6个bit，从高到低，这些bit分别控制云台的镜头缩小、镜头放大、上、下、左、右
	// 如果有做对应的操作，就将对应的bit位置1
	switch command {
	case "right":
		// 0000 0001
		cmd = 1
	case "left":
		// 0000 0010
		cmd = 2
	case "down":
		// 0000 0100
		cmd = 4
	case "up":
		// 0000 1000
		cmd = 8
	case "downright":
		// 0000 0101
		cmd = 5
	case "downleft":
		// 0000 0110
		cmd = 6
	case "upright":
		// 0000 1001
		cmd = 9
	case "upleft":
		// 0000 1010
		cmd = 10
	case "zoomin":
		// 0001 0000
		cmd = 16
	case "zoomout":
		// 0010 0000
		cmd = 32
	case "stop":
		cmd = 0
	default:
		return "", errors.New("不合规的控制字符串")
	}

	// 根据gb标准，字节4用于表示云台的镜头缩小、镜头放大、上、下、左、右，写入指令码的16进制数
	ptz.WriteString(fmt.Sprintf("%02X", cmd))

	log.Debug("合并字节4之后:" + ptz.String())

	// 根据gb标准，字节5用于表示水平控制速度，写入水平控制方向速度的十六进制数
	ptz.WriteString(fmt.Sprintf("%02X", horizonSpeed))

	// 根据gb标准，字节6用于表示垂直控制速度，写入垂直控制方向速度的十六进制数
	ptz.WriteString(fmt.Sprintf("%02X", verticalSpeed))

	// 最后字节7的高4位用于表示变倍控制速度，后4位不关注
	// 所以这里直接与0xF0做与操作，保留前4位，后4为置0
	c := zoomSpeed & 0xF0
	ptz.WriteString(fmt.Sprintf("%02X", c))

	// 字节8用于校验位，根据gb标准，校验位=(字节1+字节2+字节3+字节4+字节5+字节6+字节7) % 256
	checkCode := (0xA5 + 0x0F + 0x01 + cmd + horizonSpeed + verticalSpeed + c) % 0x100
	ptz.WriteString(fmt.Sprintf("%02X", checkCode))
	log.Debug("最终生成的PTZCmd: " + ptz.String())
	return ptz.String(), nil
}
