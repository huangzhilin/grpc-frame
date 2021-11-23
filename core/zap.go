package core

import (
	"fmt"
	"log"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var lg *zap.Logger

// InitZap 初始化Logger
func InitZap() {
	//正式环境下日志写入到文件中，测试环境下打印到控制台中
	if viper.GetString("appMode") == "release" {
		writeSyncer := getLogWriter(getLoggerFileName(),
			viper.GetInt("logger.maxSize"),
			viper.GetInt("logger.maxBackups"),
			viper.GetInt("logger.maxAge"),
			viper.GetBool("logger.compress"))
		encoder := getEncoder()
		var l = new(zapcore.Level)
		if err := l.UnmarshalText([]byte("info")); err != nil {
			log.Fatalln(err)
		}
		core := zapcore.NewCore(encoder, writeSyncer, l)
		lg = zap.New(core, zap.AddCaller())
	} else {
		lg, _ = zap.NewProduction()
	}
	zap.ReplaceGlobals(lg) // 替换zap包中全局的logger实例，后续在其他包中只需使用zap.L()调用即可
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	return zapcore.NewJSONEncoder(encoderConfig)
}

//getLogWriter 使用Lumberjack进行日志切割归档
func getLogWriter(filename string, maxSize, maxBackup, maxAge int, compress bool) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize,
		MaxBackups: maxBackup,
		MaxAge:     maxAge,
		Compress:   compress,
	}
	return zapcore.AddSync(lumberJackLogger)
}

//getLoggerFileName 获取logger日志文件名
func getLoggerFileName() string {
	now := time.Now()
	dateStr := fmt.Sprintf("./runtime/logs/%02d-%02d-%02d.log", now.Year(), int(now.Month()), now.Day())
	return dateStr
}
