package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/huangzhilin/grpc-frame/utils"

	"github.com/fsnotify/fsnotify"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

var confMD5 string

//InitConfig  初始化config
func InitConfig(configPath, configType, configToken, configKvPath, configRemoteType string) {
	switch configType {
	case "consul":
		consulClient, err := utils.NewConsulClient(configPath, configToken)
		if err != nil {
			log.Panic(fmt.Errorf("consul连接失败:: %s \n", err))
		}
		kv, _, err := consulClient.KV().Get(configKvPath, nil)
		if err != nil {
			log.Fatalln("consul获取配置失败:", err)
		}
		viper.SetConfigType(configRemoteType)
		err = viper.ReadConfig(bytes.NewBuffer(kv.Value))
		if err != nil {
			log.Fatalln("Viper解析配置失败:", err)
		}
		// 监控配置和重新获取配置
		go watchConsulConfig(configPath, configToken, configKvPath, configRemoteType)

	case "file":
		//加载公共配置文件
		viper.SetConfigFile(configPath) // 指定配置文件路径
		err := viper.ReadInConfig()     // 查找并读取配置文件
		if err != nil {                 // 处理读取配置文件的错误
			log.Panic(fmt.Errorf("Fatal error config file: %s \n", err))
		}

		// 监控配置和重新获取配置，并解决onconfigchange多次被调用问题
		viper.WatchConfig()
		viper.OnConfigChange(func(e fsnotify.Event) {
			configByte, _ := json.Marshal(viper.AllSettings())
			tempConfMD5 := utils.Md5(configByte)
			if confMD5 != tempConfMD5 {
				fmt.Println("Config file changed:", e.Name)
				confMD5 = tempConfMD5
			}
		})
	default:
		log.Panic("请输入有效的配置属性")
	}
}

//watchConsulConfig 监听consul的key/value变化
func watchConsulConfig(configPath, configToken, configKvPath, configRemoteType string) {
	time.Sleep(time.Second * 10)

	if w, err := watch.Parse(map[string]interface{}{
		"type":  "key",
		"key":   configKvPath,
		"token": configToken,
	}); err == nil {
		w.Handler = func(u uint64, i interface{}) {
			kv := i.(*consulapi.KVPair)
			viper.SetConfigType(configRemoteType)
			err = viper.ReadConfig(bytes.NewBuffer(kv.Value))
			if err != nil {
				log.Fatalln("Viper解析配置失败:", err)
			}
		}
		err = w.RunWithConfig(configPath, &consulapi.Config{Token: configToken})
		if err != nil {
			log.Fatalln("监听consul错误:", err)
		}
	} else {
		log.Fatalln(err)
	}
}
