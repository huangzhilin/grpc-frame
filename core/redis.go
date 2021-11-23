package core

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

//InitRedisClient 初始化 redis
func InitRedisClient() (*redis.Client, error) {
	RedisClient := redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis.addr"),     // 要连接的redis IP:port
		Password: viper.GetString("redis.password"), // redis 密码
		DB:       viper.GetInt("redis.db"),          // 要连接的redis 库
	})
	// 检测心跳
	_, err := RedisClient.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}
	return RedisClient, nil
}
