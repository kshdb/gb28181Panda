package storage

import (
	"fmt"
	"gb28181Panda/log"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var MysqlDb *gorm.DB

func init() {
	//创建一个数据库的连接
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		viper.GetString("mysql.username"),
		viper.GetString("mysql.password"),
		viper.GetString("mysql.host"),
		viper.GetString("mysql.port"),
		viper.GetString("mysql.name"),
	)
	//newLogger := logger.New(
	//	log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
	//	logger.Config{
	//		SlowThreshold:             time.Second, // Slow SQL threshold
	//		LogLevel:                  logger.Info, // Log level
	//		IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
	//		ParameterizedQueries:      true,        // Don't include params in the SQL log
	//		Colorful:                  false,       // Disable color
	//	},
	//)
	MysqlDb, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		//Logger: newLogger,
	})
	//显示sql语句
	MysqlDb.Debug()
	log.Info("连接Mysql成功", fmt.Sprintf("%s:%s", viper.GetString("mysql.host"), viper.GetString("mysql.port")))
	if err != nil {
		panic("连接数据库失败")
	}
}
