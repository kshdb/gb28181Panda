package config

import (
	"github.com/spf13/viper"
)

type ZlmOptions struct {
	Id       string
	Ip       string
	HttpPort int
	Secret   string
	RtspPort int
	RtmpPort int
}

var ZlmOp *ZlmOptions

func InitZlmOptions() *ZlmOptions {
	r := &ZlmOptions{}
	err := viper.UnmarshalKey("media", r)
	if err != nil {
		panic(err)
	}
	return r
}

func (zlmOp *ZlmOptions) GetId() string {
	return ZlmOp.Id
}
func (zlmOp *ZlmOptions) GetIp() string {
	return ZlmOp.Ip
}
func (zlmOp *ZlmOptions) GetHttpPort() int {
	return ZlmOp.HttpPort
}
func (zlmOp *ZlmOptions) GetSecret() string {
	return ZlmOp.Secret
}
func (zlmOp *ZlmOptions) GetRtspPort() int {
	return ZlmOp.RtspPort
}
func (zlmOp *ZlmOptions) GetRtmpPort() int {
	return ZlmOp.RtmpPort
}
