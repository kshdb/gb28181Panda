package model

import (
	"fmt"
	"gb28181Panda/storage"
)

type Channel struct {
	Id            int    `json:"id"`                                              //type:int      comment:
	DeviceId      string `gorm:"column:device_id" json:"deviceId" xml:"DeviceID"` //type:string   comment:通道Id号
	Name          string `gorm:"column:name" json:"name"`                         //type:string   comment:通道名称
	Manufacturer  string `gorm:"column:manufacturer" json:"manufacturer"`         //type:string   comment:当为设备时，设备厂商
	Model         string `gorm:"column:model" json:"model"`                       //type:string   comment:当为设备时，设备型号
	Owner         string `gorm:"column:owner" json:"owner"`                       //type:string   comment:当为设备时，设备归属
	CivilCode     string `gorm:"column:civil_code" json:"civilCode"`              //type:string   comment:行政区域
	Address       string `gorm:"column:address" json:"address"`                   //type:string   comment:当为设备时，安装地址
	Parental      string `gorm:"column:parental" json:"parental"`                 //type:string   comment:当为设备时，是否有子设备，1有，0没有
	ParentId      string `gorm:"column:parent_id" json:"parentId" xml:"ParentID"` //type:string   comment:父设备/区域/系统ID
	SafetyWay     string `gorm:"column:safety_way" json:"safetyWay"`              //type:string   comment:信令安全模式，0不采用、2 S/MIME签名方式、3 S/MIME加密他签名同时采用方式、4 数字摘要方式
	RegisterWay   string `gorm:"column:register_way" json:"registerWay"`          //type:string   comment:注册方式，1 标准认证注册模式 、2 基于口令的双向认证模式、3 基于数字证书的双向认证注册模式
	Secrecy       string `gorm:"column:secrecy" json:"secrecy"`                   //type:string   comment:保密属性，0不涉密、1涉密
	Status        string `gorm:"column:status" json:"status"`                     //type:string   comment:通道状态
	TransportType string `gorm:"column:transport_type" json:"transportType"`      //type:string   comment:流媒体传输方式  udp、tcp被动 tcppassive、tcp主动  tcpactive
	MediaStatus   string `gorm:"column:media_status" json:"mediaStatus"`          //type:string   comment:流媒体当前直播状态 OPEN 有人观看  CLOSE 无人观看
	CreatedAt     string `gorm:"column:created_at" json:"createdAt"`              //type:string   comment:
	UpdatedAt     string `gorm:"column:updated_at" json:"updatedAt"`              //type:string   comment:
}

func (channel *Channel) ChannelList() ([]Channel, error) {
	var list []Channel
	if err := storage.MysqlDb.Table("t_channel").Where("parent_id = ?", channel.DeviceId).Order("id ASC").Find(&list).Error; err != nil {
		return nil, err
	} else {
		return list, nil
	}
}

func (channel *Channel) ChannelDetail() (Channel, error) {
	var detail Channel
	if err := storage.MysqlDb.Table("t_channel").Where("device_id = ?", channel.DeviceId).First(&detail).Error; err != nil {
		return Channel{}, err
	} else {
		return detail, nil
	}
}

func (channel *Channel) ChannelAdd() error {
	if err := storage.MysqlDb.Debug().Table("t_channel").Create(&channel).Error; err != nil {
		fmt.Println("err", err)
		return err
	} else {
		return nil
	}
}

func (channel *Channel) ChannelUpdate() error {
	if err := storage.MysqlDb.Debug().Table("t_channel").Where("device_id = ?", channel.DeviceId).Updates(channel).Error; err != nil {
		return err
	} else {
		return nil
	}
}

func (channel *Channel) ChannelDelete() error {
	if err := storage.MysqlDb.Table("t_channel").Where("device_id = ? AND parent_id = ?", channel.DeviceId, channel.ParentId).Delete(&channel).Error; err != nil {
		return err
	} else {
		return nil
	}
}
func (channel *Channel) ChannelDeleteWithParentId() error {
	if err := storage.MysqlDb.Table("t_channel").Where("parent_id = ?", channel.ParentId).Delete(&channel).Error; err != nil {
		return err
	} else {
		return nil
	}
}
