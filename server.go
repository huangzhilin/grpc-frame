package grpc_frame

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/huangzhilin/grpc-frame/core"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type ServerConfig struct {
	Path           string //配置类型：file->本地配置文件;consul->consul配置中心
	Type           string //配置路径: 如./config/config_develop.yaml 或 http://192.168.3.91:8500
	Token          string //consul的token，type=file时为空
	KvPath         string //consul的远程key/value路径
	RemoteType     string //consul远程配置类型，如：yaml、json、hcl
	GrpcServer     *grpc.Server
	GatewayService *http.Server
	GatewayConn    *grpc.ClientConn
	GatewayMux     *runtime.ServeMux
	Interceptors   []grpc.ServerOption //拦截器
	ServeMuxOption []runtime.ServeMuxOption
}

var errChan = make(chan error) //程序退出让服务注销掉

//Init 初始化配置,包括：初始化gorm、redis、zap日志等等初始化行为
func (t *ServerConfig) Init() *ServerConfig {
	t.GrpcServer = grpc.NewServer(t.Interceptors...)

	core.InitConfig(t.Path, t.Type, t.Token, t.KvPath, t.RemoteType)

	core.InitZap() //初始化zap日志

	return t
}

//Gorm 初始化gorm
func (t *ServerConfig) Gorm() *ServerConfig {
	DB, DBList = core.InitGorm()
	return t
}

//Redis 初始化redis
func (t *ServerConfig) Redis() *ServerConfig {
	Redis, _ = core.InitRedisClient()
	return t
}

//GrpcGateway 初始化grpc-gateway(前提需要先开启grpc服务)
func (t *ServerConfig) GrpcGateway() *ServerConfig {
	var err error
	if t.GatewayConn, err = grpc.DialContext(
		context.Background(),
		":"+viper.GetString("grpc.port"),
		grpc.WithBlock(),
		grpc.WithInsecure(),
	); err != nil {
		fmt.Println("初始化grpc-gateway失败：" + err.Error())
		return nil
	}
	//一个对外开放的mux
	t.GatewayMux = runtime.NewServeMux(
		t.ServeMuxOption...,
	)
	return t
}

//RunGrpcGateway 开启并运行grpc-gateway服务（前提需要通过ServerConfig.GrpcGateway()进行初始化）
func (t *ServerConfig) RunGrpcGateway(middleware ...func(next http.Handler) http.Handler) {
	go func() {
		fmt.Println("Listening and serving GRPC-WAGEWAY on :" + viper.GetString("grpc.gateway.port"))

		mux := http.NewServeMux()

		middlewareLen := len(middleware)
		if middlewareLen != 0 { //如果需要中间件，使用下面语句
			var handler http.Handler
			for i := middlewareLen - 1; i >= 0; i-- {
				if i == middlewareLen-1 {
					handler = middleware[i](t.GatewayMux)
				} else {
					handler = middleware[i](handler)
				}
			}
			mux.Handle("/", handler)
		} else {
			mux.Handle("/", t.GatewayMux)
		}

		t.GatewayService = &http.Server{
			Handler: mux,
			Addr:    ":" + viper.GetString("grpc.gateway.port"),
		}
		if err := t.GatewayService.ListenAndServe(); err != nil {
			errChan <- err
		}
	}()
}

//RunGrpc 开启并运行grpc服务
func (t *ServerConfig) RunGrpc() {
	go (func() {
		fmt.Println("Listening and serving GRPC on :" + viper.GetString("grpc.port"))
		lis, err := net.Listen("tcp", ":"+viper.GetString("grpc.port"))
		if err != nil {
			fmt.Println(err)
			return
		}
		if err := t.GrpcServer.Serve(lis); err != nil {
			errChan <- err
		}
	})()
}

//Wait 等待并阻塞和实现优雅关闭服务
func (t *ServerConfig) Wait() {
	go (func() {
		sigC := make(chan os.Signal)
		signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM) //监听kill  或 ctrl+c 等退出程序
		errChan <- fmt.Errorf("%s", <-sigC)
	})()

	getErr := <-errChan //如果没有错误，则一直阻塞，不会进行下面的操作
	fmt.Println(getErr)

	log.Println("准备关闭Grpc服务...")
	t.GrpcServer.GracefulStop()

	if t.GatewayService != nil {
		log.Println("准备关闭Grpc-Gateway服务...")
		// 创建一个5秒超时的context
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := t.GatewayService.Shutdown(ctx); err != nil {
			log.Fatal("关闭Grpc-Gateway服务: ", err)
		}
	}

	log.Println("服务已关闭！")
}
