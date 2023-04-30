package sipServer

import (
	"fmt"
	"gb28181Panda/log"
	"gb28181Panda/util"
	"github.com/pkg/errors"
	"sync"
	"time"
)

// MessageRecordInfoResponse 目录列表
type MessageRecordInfoResponse struct {
	CmdType  string       `xml:"CmdType"`
	SN       int          `xml:"SN"`
	DeviceID string       `xml:"DeviceID"`
	SumNum   int          `xml:"SumNum"`
	Item     []RecordItem `xml:"RecordList>Item"`
}

// RecordItem 目录详情
type RecordItem struct {
	// DeviceID 设备编号
	DeviceID string `xml:"DeviceID" bson:"DeviceID" json:"DeviceID"`
	// Name 设备名称
	Name      string `xml:"Name" bson:"Name" json:"Name"`
	FilePath  string `xml:"FilePath" bson:"FilePath" json:"FilePath"`
	Address   string `xml:"Address" bson:"Address" json:"Address"`
	StartTime string `xml:"StartTime" bson:"StartTime" json:"StartTime"`
	EndTime   string `xml:"EndTime" bson:"EndTime" json:"EndTime"`
	Secrecy   int    `xml:"Secrecy" bson:"Secrecy" json:"Secrecy"`
	Type      string `xml:"Type" bson:"Type" json:"Type"`
}

type recordList struct {
	ch   chan int
	num  int
	data []RecordInfo
}

// RecordInfo 录像开始时间段
type RecordInfo struct {
	StartTime int64 `json:"startTime"`
	EndTime   int64 `json:"endTime"`
}

// 当前获取目录文件设备集合
var _recordList *sync.Map

func sipMessageRecordInfo(body string) error {
	message := &MessageRecordInfoResponse{}
	if err := util.XMLDecode([]byte(body), message); err != nil {
		log.Info("Message Unmarshal xml err:", err, "body:", string(body))
		return err
	}
	recordKey := fmt.Sprintf("%s%d", message.DeviceID, message.SN)
	if list, ok := _recordList.Load(recordKey); ok {
		info := list.(recordList)
		info.num += len(message.Item)
		for _, item := range message.Item {
			s, _ := time.ParseInLocation("2006-01-02T15:04:05", item.StartTime, time.Local)
			e, _ := time.ParseInLocation("2006-01-02T15:04:05", item.EndTime, time.Local)
			recordInfo := RecordInfo{
				StartTime: s.Unix(),
				EndTime:   e.Unix(),
			}
			info.data = append(info.data, recordInfo)
		}
		if info.num == message.SumNum {
			// 获取到完整数据
			_recordList.Store(recordKey, info)
			info.ch <- 1
		}
		_recordList.Store(recordKey, info)
		return nil
	}
	return errors.New("未查询到历史视频")
}

func parseRecordData(data []RecordInfo) []RecordInfo {
	var tempData []RecordInfo
	for _, v := range data {
		if len(tempData) > 0 {
			recordL := len(tempData)
			if tempData[recordL-1:][0].EndTime == v.StartTime {
				tempData[recordL-1:][0].EndTime = v.EndTime
			} else {
				tempData = append(tempData, v)
			}
		} else {
			tempData = append(tempData, v)
		}
	}
	return tempData
}
