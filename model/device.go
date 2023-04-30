package model

import (
	"fmt"
	"gb28181Panda/storage"
)

type Device struct {
	Id           int    `json:"id"`                                      //type:int      comment:id
	DeviceId     string `gorm:"column:device_id" json:"deviceId"`        //type:string   comment:设备deviceId号
	Domain       string `gorm:"column:domain" json:"domain"`             //type:string   comment:设备域
	Name         string `gorm:"column:name" json:"name"`                 //type:string   comment:设备名称
	Manufacturer string `gorm:"column:manufacturer" json:"manufacturer"` //type:string   comment:设备厂家
	Model        string `gorm:"column:model" json:"model"`               //type:string   comment:设备型号
	Firmware     string `gorm:"column:firmware" json:"firmware"`         //type:string   comment:设备固件版本
	Transport    string `gorm:"column:transport" json:"transport"`       //type:string   comment:传输模式
	Status       string `gorm:"column:status" json:"status"`             //type:string   comment:上下线状态on 上线  off 下线
	Ip           string `gorm:"column:ip" json:"ip"`                     //type:string   comment:设备ip地址
	Port         string `gorm:"column:port" json:"port"`                 //type:string   comment:设备端口号
	Expires      string `gorm:"column:expires" json:"expires"`           //type:string   comment:设备有效时间
	CreatedAt    string `gorm:"column:created_at" json:"createdAt"`      //type:string   comment:创建时间
	UpdatedAt    string `gorm:"column:updated_at" json:"updatedAt"`      //type:string   comment:更新时间
	RegisterAt   string `gorm:"column:register_at" json:"registerAt"`    //type:string   comment:注册时间
	KeepaliveAt  string `gorm:"column:keepalive_at" json:"keepaliveAt"`  //type:string   comment:保持状态时间
}

// DeviceTree 设备通道树状数据
type DeviceTree struct {
	Id           uint                 `json:"id"`
	DeviceId     string               `json:"deviceId" gorm:"primaryKey"`
	Name         string               `json:"name"`
	Manufacturer string               `json:"manufacturer"`
	Model        string               `json:"model"`
	Status       string               `json:"status"`
	KeepaliveAt  string               `json:"keepaliveAt"`
	Channel      []*DeviceTreeChannel `json:"channel" gorm:"foreignKey:ParentId"`
}

// DeviceTreeChannel 通道数据
type DeviceTreeChannel struct {
	DeviceId     string `json:"deviceId"`
	Name         string `json:"name"`
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	ParentId     string `json:"parentId"`
}

func (deviceTree *DeviceTree) TableName() string {
	return "t_device"
}
func (deviceTreeChannel *DeviceTreeChannel) TableName() string {
	return "t_channel"
}
func (deviceTree *DeviceTree) DeviceTreeData() ([]DeviceTree, error) {
	var dataTree []DeviceTree
	if err := storage.MysqlDb.Preload("Channel").Find(&dataTree).Error; err != nil {
		return nil, err
	} else {
		return dataTree, nil
	}
}

func (device *Device) DeviceList() ([]Device, error) {
	var list []Device
	if err := storage.MysqlDb.Table("t_device").Order("id ASC").Find(&list).Error; err != nil {
		return nil, err
	} else {
		return list, nil
	}
}

func (device *Device) DeviceDetail() (Device, error) {
	var detail Device
	if err := storage.MysqlDb.Debug().Table("t_device").Where("device_id = ?", device.DeviceId).First(&detail).Error; err != nil {
		return Device{}, err
	} else {
		return detail, nil
	}
}

func (device *Device) DeviceAdd() error {
	if err := storage.MysqlDb.Debug().Table("t_device").Create(&device).Error; err != nil {
		fmt.Println("err", err)
		return err
	} else {
		return nil
	}
}

func (device *Device) DeviceUpdate() error {
	if err := storage.MysqlDb.Debug().Table("t_device").Where("device_id = ?", device.DeviceId).Updates(device).Error; err != nil {
		return err
	} else {
		return nil
	}
}

func (device *Device) DeviceDelete() error {
	if err := storage.MysqlDb.Table("t_device").Where("device_id = ?", device.DeviceId).Delete(&device).Error; err != nil {
		return err
	} else {
		return nil
	}
}
