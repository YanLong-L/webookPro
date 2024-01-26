package startup

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
	"webookpro/config"
	"webookpro/internal/repository/dao"
)

func InitDB() *gorm.DB {
	// 初始化db
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN), &gorm.Config{})
	//db, err := gorm.Open(mysql.Open(viper.GetString("db.dsn")))
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

var mongoDB *mongo.Database
var mclient *mongo.Client

func InitMongoDB() (*mongo.Client, *mongo.Database) {
	if mongoDB == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		monitor := &event.CommandMonitor{
			Started: func(ctx context.Context,
				startedEvent *event.CommandStartedEvent) {
				fmt.Println(startedEvent.Command)
			},
		}
		opts := options.Client().
			ApplyURI("mongodb://root:example@localhost:27017/").
			SetMonitor(monitor)
		mclient, err := mongo.Connect(ctx, opts)
		if err != nil {
			panic(err)
		}
		mongoDB = mclient.Database("webookpro")
	}
	return mclient, mongoDB
}
