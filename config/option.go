package config

import (
	"github.com/spf13/viper"
)

func init() {
	loadConfig()
}

func loadConfig() {
	viper.AddConfigPath("./")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		panic("load config fail,please check your config file whether in config/ in the directory")
	}
	startOption()
}

func startOption() {
	//加载配置项后加载zlm的配置
	ZlmOp = InitZlmOptions()
	//加载配置项后加载sip的配置
	SipOp = InitSipOptions()
}
