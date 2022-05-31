package main

import (
	"context"

	log "github.com/asyrafduyshart/go-reverse-proxy/log"
	redis "github.com/go-redis/redis/v8"
)

type redisClient struct {
	c *redis.Client
}

var (
	client = &redisClient{}
	ctx    = context.Background()
)

func RedisInit(url string) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		log.Error("Error redis parse url %s" + err.Error())
	}
	c := redis.NewClient(opt)
	client.c = c
	errPing := Ping()
	if errPing != nil {
		log.Error("Unable to connect to redis %v", err.Error())
	} else {
		log.Info("redis %s connect!", url)
	}

}

func Ping() error {
	return client.c.Ping(ctx).Err()
}

func GetKeyField(key string, field string) (string, error) {
	redisData := client.c.HGet(ctx, key, field)
	data, err := redisData.Result()
	return data, err
}
