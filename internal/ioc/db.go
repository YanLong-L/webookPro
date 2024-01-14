package ioc

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"webookpro/config"
	"webookpro/internal/repository/dao"
)

func InitDB() *gorm.DB {
	// 初始化db
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN), &gorm.Config{})
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
