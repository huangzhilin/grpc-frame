package core

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/gorm/utils"
)

var db *gorm.DB
var dbList = make(map[string]*gorm.DB)

var Enforcer *casbin.Enforcer

//InitGorm 初始化 数据库 并 产生数据库全局变量
func InitGorm() (*gorm.DB, map[string]*gorm.DB) {
	var err error
	mysqlList := viper.GetStringMap("gorm")
	if len(mysqlList) <= 0 {
		log.Panic(errors.New("mysql没有配置哦"))
	}

	for mysqlName, mysqlInfo := range mysqlList {
		mysqlInfoMap := mysqlInfo.(map[string]interface{})
		mysqlConfig := mysql.Config{
			DSN: mysqlInfoMap["dsn"].(string),
			//DefaultStringSize:         255,   // string 类型字段的默认长度
			//DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
			//DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
			//DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
			//SkipInitializeWithVersion: false, // 根据版本自动配置
		}

		gormConf := &gorm.Config{
			SkipDefaultTransaction:                   true, //为了确保数据一致性，GORM 会在事务里执行写入操作（创建、更新、删除）。如果没有这方面的要求，您可以在初始化时禁用它
			DisableForeignKeyConstraintWhenMigrating: true, //在 AutoMigrate 或 CreateTable 时，GORM 会自动创建外键约束，若要禁用该特性，可将其设置为 true
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true, // 使用单数表名
			},
		}

		//开启sql日志打印
		if mysqlInfoMap["debug"].(bool) {
			//gorm使用zap作为logging
			zapLogger := logger2zap(zap.L())
			zapLogger.SetAsDefault()
			gormConf.Logger = zapLogger.LogMode(gormlogger.Info)
			//gormConf.Logger = logger.Default.LogMode(logger.Info)
		}

		if dbList[mysqlName], err = gorm.Open(mysql.New(mysqlConfig), gormConf); err != nil {
			log.Panic(err.Error())
		}

		//连接池设置
		sqlDB, _ := dbList[mysqlName].DB()
		sqlDB.SetMaxIdleConns(10)  //设置空闲连接池中连接的最大数量
		sqlDB.SetMaxOpenConns(100) //设置打开数据库连接的最大数量

		if mysqlName == "default" {
			db = dbList[mysqlName]
		}
	}

	return db, dbList
}

//Logger gorm使用zap作为logging
type Logger struct {
	ZapLogger                 *zap.Logger
	LogLevel                  gormlogger.LogLevel
	SlowThreshold             time.Duration
	SkipCallerLookup          bool
	IgnoreRecordNotFoundError bool
}

func logger2zap(zapLogger *zap.Logger) Logger {
	return Logger{
		ZapLogger:                 zapLogger,
		LogLevel:                  gormlogger.Warn,
		SlowThreshold:             100 * time.Millisecond,
		SkipCallerLookup:          false,
		IgnoreRecordNotFoundError: false,
	}
}

func (l Logger) SetAsDefault() {
	gormlogger.Default = l
}

func (l Logger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return Logger{
		ZapLogger:                 l.ZapLogger,
		SlowThreshold:             l.SlowThreshold,
		LogLevel:                  level,
		SkipCallerLookup:          l.SkipCallerLookup,
		IgnoreRecordNotFoundError: l.IgnoreRecordNotFoundError,
	}
}

func (l Logger) Info(ctx context.Context, str string, args ...interface{}) {
	if l.LogLevel >= gormlogger.Info {
		l.ZapLogger.Sugar().Infof(str, args...)
	}
}

func (l Logger) Warn(ctx context.Context, str string, args ...interface{}) {
	if l.LogLevel >= gormlogger.Warn {
		l.ZapLogger.Sugar().Infof(str, args...)
	}
}

func (l Logger) Error(ctx context.Context, str string, args ...interface{}) {
	if l.LogLevel >= gormlogger.Error {
		l.ZapLogger.Sugar().Infof(str, args...)
	}
}

func (l Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= gormlogger.Error && (!l.IgnoreRecordNotFoundError || !errors.Is(err, gormlogger.ErrRecordNotFound)):
		sql, rows := fc()
		l.ZapLogger.Info("Error", zap.String("file", utils.FileWithLineNum()), zap.Error(err), zap.Int64("rows", rows), zap.String("sql", sql), zap.Duration("elapsed", elapsed))
	case l.SlowThreshold != 0 && elapsed > l.SlowThreshold && l.LogLevel >= gormlogger.Warn:
		sql, rows := fc()
		l.ZapLogger.Info("Warn", zap.String("file", utils.FileWithLineNum()), zap.Int64("rows", rows), zap.String("sql", sql), zap.Duration("elapsed", elapsed))
	case l.LogLevel >= gormlogger.Info:
		sql, rows := fc()
		l.ZapLogger.Info("Debug", zap.String("file", utils.FileWithLineNum()), zap.Int64("rows", rows), zap.String("sql", sql), zap.Duration("elapsed", elapsed))
	}
}
