package util

import (
	"gb28181Panda/log"
	"github.com/beevik/etree"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"math/rand"
	"time"
)

type QueryType string
type ControlType string

type WithKeyValue func(element *etree.Element)

const (
	DeviceConfig  ControlType = "DeviceConfig"
	DeviceControl ControlType = "DeviceControl"
)

const (
	DeviceStatusCmdType   QueryType = "DeviceStatus"
	CatalogCmdType        QueryType = "Catalog"
	DeviceInfoCmdType     QueryType = "DeviceInfo"
	RecordInfoCmdType     QueryType = "RecordInfo"
	AlarmCmdType          QueryType = "Alarm"
	ConfigDownloadCmdType QueryType = "ConfigDownload"
	PresetQueryCmdType    QueryType = "PresetQuery"
	MobilePositionCmdType QueryType = "MobilePosition"
	KeepaliveCmdType      QueryType = "Keepalive"
)

// CreateQueryXML 主动查询设备相关信息xml文件
func CreateQueryXML(cmd QueryType, deviceId string, kvs ...WithKeyValue) (string, error) {
	document := etree.NewDocument()
	document.CreateProcInst("xml", "version=\"1.0\" encoding=\"UTF-8\"")
	query := document.CreateElement("Query")
	query.CreateElement("CmdType").CreateText(string(cmd))
	query.CreateElement("SN").CreateText(getSN())
	query.CreateElement("DeviceID").CreateText(deviceId)

	for _, kv := range kvs {
		kv(query)
	}

	document.Indent(2)
	body, err := document.WriteToString()
	if err != nil {
		log.Error(err)
		return "", errors.Wrap(err, "encoding catalog query request xml fail")
	}
	return body, nil
}

// CreateIFameXML 主动查询设备相关信息xml文件
func CreateIFameXML(deviceId string, kvs ...WithKeyValue) (string, error) {
	document := etree.NewDocument()
	document.CreateProcInst("xml", "version=\"1.0\" encoding=\"UTF-8\"")
	query := document.CreateElement("Control")
	query.CreateElement("CmdType").CreateText("DeviceControl")
	query.CreateElement("SN").CreateText(getSN())
	query.CreateElement("DeviceID").CreateText(deviceId)
	query.CreateElement("IFameCmd").CreateText("send")
	info := query.CreateElement("Info")
	info.CreateElement("IFameCmd").CreateText("5")
	for _, kv := range kvs {
		kv(query)
	}

	document.Indent(4)
	body, err := document.WriteToString()
	if err != nil {
		log.Error(err)
		return "", errors.Wrap(err, "encoding catalog query request xml fail")
	}
	return body, nil
}

// CreateGuardXML 布防xml文件
func CreateGuardXML(cmd ControlType, deviceId string, kvs ...WithKeyValue) (string, error) {
	document := etree.NewDocument()
	document.CreateProcInst("xml", "version=\"1.0\" encoding=\"UTF-8\"")
	query := document.CreateElement("Control")
	query.CreateElement("CmdType").CreateText(string(cmd))
	query.CreateElement("SN").CreateText(getSN())
	query.CreateElement("DeviceID").CreateText(deviceId)
	query.CreateElement("GuardCmd").CreateText("SetGuard")

	for _, kv := range kvs {
		kv(query)
	}

	document.Indent(4)
	body, err := document.WriteToString()
	if err != nil {
		log.Error(err)
		return "", errors.Wrap(err, "encoding catalog query request xml fail")
	}
	return body, nil
}

// CreateSubscribeAlarmXML 订阅预警信息xml文件
func CreateSubscribeAlarmXML(cmd QueryType, deviceId string, kvs ...WithKeyValue) (string, error) {
	document := etree.NewDocument()
	document.CreateProcInst("xml", "version=\"1.0\" encoding=\"UTF-8\"")
	query := document.CreateElement("Query")
	query.CreateElement("CmdType").CreateText(string(cmd))
	query.CreateElement("SN").CreateText(getSN())
	query.CreateElement("DeviceID").CreateText(deviceId)
	query.CreateElement("StartAlarmPriority").CreateText("0")
	query.CreateElement("EndAlarmPriority").CreateText("0")
	query.CreateElement("AlarmMethod").CreateText("0")
	query.CreateElement("StartTime").CreateText("2023-04-25T00:00:00")
	query.CreateElement("EndTime").CreateText("2023-04-27T23:59:59")

	for _, kv := range kvs {
		kv(query)
	}

	document.Indent(4)
	body, err := document.WriteToString()
	if err != nil {
		log.Error(err)
		return "", errors.Wrap(err, "encoding catalog query request xml fail")
	}
	return body, nil
}

// CreateControlXml Ptz控制指令
func CreateControlXml(cmd ControlType, deviceId string, kvs ...WithKeyValue) (string, error) {
	document := etree.NewDocument()
	document.CreateProcInst("xml", "version=\"1.0\" encoding=\"UTF-8\"")
	query := document.CreateElement("ControlPTZ")
	query.CreateElement("CmdType").CreateText(string(cmd))
	query.CreateElement("SN").CreateText(getSN())
	query.CreateElement("DeviceID").CreateText(deviceId)

	for _, kv := range kvs {
		kv(query)
	}

	document.Indent(2)
	body, err := document.WriteToString()
	if err != nil {
		log.Error(err)
		return "", errors.Wrap(err, "encoding device control request xml fail")
	}
	return body, nil

}

// CreateRecordInfoXml 查询录像数据
func CreateRecordInfoXml(cmd QueryType, channelId string, startTime string, endTime string, kvs ...WithKeyValue) (string, string, error) {
	sn := getSN()
	document := etree.NewDocument()
	document.CreateProcInst("xml", "version=\"1.0\" encoding=\"UTF-8\"")
	query := document.CreateElement("Query")
	query.CreateElement("CmdType").CreateText(string(cmd))
	query.CreateElement("SN").CreateText(sn)
	query.CreateElement("DeviceID").CreateText(channelId)
	query.CreateElement("StartTime").CreateText(startTime)
	query.CreateElement("EndTime").CreateText(endTime)
	query.CreateElement("Secrecy").CreateText("0")
	query.CreateElement("Type").CreateText("all")

	for _, kv := range kvs {
		kv(query)
	}

	document.Indent(2)
	body, err := document.WriteToString()
	if err != nil {
		log.Error(err)
		return "", "", errors.Wrap(err, "encoding device control request xml fail")
	}
	return body, sn, nil

}

// WithPTZCmd create 'PTZCmd' item of xml by value
func WithPTZCmd(ptz string) WithKeyValue {
	return func(element *etree.Element) {
		element.CreateElement("PTZCmd").CreateText(ptz)
	}
}

func getSN() string {
	rand.New(rand.NewSource(time.Now().UnixMilli()))
	return cast.ToString(rand.Intn(10) * 9876)
}
