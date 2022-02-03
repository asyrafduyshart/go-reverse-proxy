package main

import (
	"context"

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
		panic("Error redis parse url " + err.Error())
	}
	c := redis.NewClient(opt)
	if err := c.Ping(ctx).Err(); err != nil {
		panic("Unable to connect to redis " + err.Error())
	}
	client.c = c
}

func GetKeyField(key string, field string) (string, error) {
	redisData := client.c.HGet(ctx, key, field)
	data, err := redisData.Result()
	return data, err
}
