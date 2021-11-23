package utils

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

//GrpcWithConsul 通过grpc方式来调用consul中心的通过grpc协议的微服务
func GrpcWithConsul(serverName string) (*grpc.ClientConn, error) {
	//连接consul
	consulAddress := viper.GetString("consulCenter") //consul注册中心的IP地址
	client, err := NewConsulClient(consulAddress, "")
	if err != nil {
		return nil, err
	}

	//根据服务名和tag来找出所有健康的服务
	serviceHealthy, _, err := client.Health().Service(serverName, "grpc", true, nil)
	if err != nil {
		return nil, err
	}
	countService := len(serviceHealthy)

	if countService < 1 {
		return nil, errors.New("没有找到相关服务：server_name=" + serverName + ",tag=grpc")
	}

	seed := time.Now().UnixNano()
	r := rand.New(rand.NewSource(seed))
	//负载均衡   随机
	serviceOne := serviceHealthy[r.Intn(countService)]
	target := serviceOne.Service.Address + ":" + strconv.Itoa(serviceOne.Service.Port)

	//连接grpc服务    grpc.WithInsecure() 表示以安全的方式操作
	//客户端grpc拦截器，如果需要的话
	interceptor := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		fmt.Printf("该次请求总耗时：%s", time.Since(start))
		return err
	}
	opt := grpc.WithUnaryInterceptor(interceptor)

	grpcConn, err := grpc.Dial(target, grpc.WithInsecure(), opt)
	if err != nil {
		return nil, err
	}
	return grpcConn, nil
}

//newConsulClient 连接consul服务
func NewConsulClient(address, token string) (apiClient *consulapi.Client, err error) {
	config := consulapi.DefaultConfig()
	config.Address = address
	if token != "" {
		config.Token = token
	}
	if apiClient, err = consulapi.NewClient(config); err != nil {
		return nil, err
	}
	return apiClient, nil
}
