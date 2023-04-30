package config

import (
	"github.com/spf13/viper"
)

type RedisOptions struct {
	Addr     string
	Password string
	DB       int
}

func NewRedisOptions() *RedisOptions {
	r := &RedisOptions{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       3,
	}
	err := viper.UnmarshalKey("redis", r)
	if err != nil {
		panic(err)
	}
	return r
}
