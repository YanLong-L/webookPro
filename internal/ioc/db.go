package ioc

import (
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"webookpro/internal/repository/dao"
)

func InitDB() *gorm.DB {
	// 初始化db
	//db, err := gorm.Open(mysql.Open(config.Config.DB.DSN), &gorm.Config{})
	db, err := gorm.Open(mysql.Open(viper.GetString("db.dsn")))
	if err != nil {
		panic(err)
	}
	// 初始化table
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}
