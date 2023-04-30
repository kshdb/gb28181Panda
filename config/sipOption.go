package config

import (
	"github.com/spf13/viper"
)

type SipOptions struct {
	Id         string
	Realm      string
	Ip         string
	Port       int
	Password   string
	UserAgent  string
	CheckRealm int
}

var SipOp *SipOptions

func InitSipOptions() *SipOptions {
	r := &SipOptions{}
	err := viper.UnmarshalKey("sip", r)
	if err != nil {
		panic(err)
	}
	return r
}
