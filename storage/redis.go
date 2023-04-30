package storage

import (
	"context"
	"fmt"
	"gb28181Panda/config"
	"gb28181Panda/log"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

type redisClient struct {
	rdb *redis.Client
}

func newRedis() *redisClient {
	redisConfig := config.NewRedisOptions()
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Addr,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		panic(fmt.Errorf("connection to redis fail,addr: %s, err: %w",
			redisConfig.Addr,
			err,
		))
	}
	rdb.IncrBy(context.Background(), "ceq", 1)
	log.Info("连接Redis成功", redisConfig.Addr)
	return &redisClient{
		rdb: rdb,
	}
}

func (r *redisClient) Get(key string) (string, error) {
	result, err := r.rdb.Get(context.Background(), key).Result()
	if err != nil {
		log.Error(err)
	}
	return result, err
}
func (r *redisClient) Set(key string, val interface{}) {
	if err := r.rdb.Set(context.Background(), key, val, redis.KeepTTL).Err(); err != nil {
		log.Error(err)
	}
}

func (r *redisClient) Del(key string) error {
	_, err := r.rdb.Del(context.Background(), key).Result()
	if err != nil {
		log.Error(err)
		return errors.New(err.Error())
	}
	return err
}

func (r *redisClient) GetCeq(key string) (int64, error) {
	return r.rdb.Incr(context.Background(), key).Result()
}
