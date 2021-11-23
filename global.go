package grpc_frame

import (
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

var (
	DB     *gorm.DB
	DBList = make(map[string]*gorm.DB)

	Redis *redis.Client
)
