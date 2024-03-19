package ioc

import (
	promsdk "github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
	dao2 "webookpro/interactive/repository/dao"
	"webookpro/pkg/gormx/connpool"
	"webookpro/pkg/logger"
)

type SrcDB *gorm.DB
type DstDB *gorm.DB

// InitSRC 初始化源库
func InitSRC(l logger.Logger) SrcDB {
	return InitDB(l, "src")
}

// InitDST 初始化目标库
func InitDST(l logger.Logger) DstDB {
	return InitDB(l, "dst")
}

// InitDoubleWritePool 初始化gorm的connpool，用于在初始化gorm时配置
func InitDoubleWritePool(src SrcDB, dst DstDB) *connpool.DoubleWritePool {
	pattern := viper.GetString("migrator.pattern")
	return connpool.NewDoubleWritePool(src.ConnPool, dst.ConnPool, pattern)
}

// InitBizDB 这个是进行数据迁移时，业务用的，支持双写的 DB
func InitBizDB(pool *connpool.DoubleWritePool) *gorm.DB {
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn: pool,
	}))
	if err != nil {
		panic(err)
	}
	return db
}

// InitDB 初始化一个gormDB 通过一个key,区分加载配置文件中的哪个库的数据库dsn
func InitDB(l logger.Logger, key string) *gorm.DB {
	//type Config struct {
	//	DSN string `yaml:"dsn"`
	//}
	//var cfg = Config{
	//	DSN: "root:root@tcp(localhost:13316)/webook_default",
	//}
	//err := viper.UnmarshalKey("db."+key, &cfg)
	dsn := viper.GetString("db." + key + ".dsn")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	// 接入prometheus
	//000cb := newCallbacks(key)
	//err = db.Use(cb)
	//if err != nil {
	//	panic(err)
	//}

	// 初始化表
	err = dao2.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

type Callbacks struct {
	vector *promsdk.SummaryVec
}

func (pcb *Callbacks) Name() string {
	return "prometheus-query"
}

func (pcb *Callbacks) Initialize(db *gorm.DB) error {
	pcb.registerAll(db)
	return nil
}

func newCallbacks(key string) *Callbacks {
	vector := promsdk.NewSummaryVec(promsdk.SummaryOpts{
		// 在这边，你要考虑设置各种 Namespace
		Namespace: "geekbang_daming",
		Subsystem: "webook",
		Name:      "gorm_query_time_" + key,
		Help:      "统计 GORM 的执行时间",
		ConstLabels: map[string]string{
			"db": "webook",
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.9:   0.01,
			0.99:  0.005,
			0.999: 0.0001,
		},
	},
		// 如果是 JOIN 查询，table 就是 JOIN 在一起的
		// 或者 table 就是主表，A JOIN B，记录的是 A
		[]string{"type", "table"})

	pcb := &Callbacks{
		vector: vector,
	}
	promsdk.MustRegister(vector)
	return pcb
}

func (pcb *Callbacks) registerAll(db *gorm.DB) {
	// 作用于 INSERT 语句
	err := db.Callback().Create().Before("*").
		Register("prometheus_create_before", pcb.before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Create().After("*").
		Register("prometheus_create_after", pcb.after("create"))
	if err != nil {
		panic(err)
	}

	err = db.Callback().Update().Before("*").
		Register("prometheus_update_before", pcb.before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Update().After("*").
		Register("prometheus_update_after", pcb.after("update"))
	if err != nil {
		panic(err)
	}

	err = db.Callback().Delete().Before("*").
		Register("prometheus_delete_before", pcb.before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Delete().After("*").
		Register("prometheus_delete_after", pcb.after("delete"))
	if err != nil {
		panic(err)
	}

	err = db.Callback().Raw().Before("*").
		Register("prometheus_raw_before", pcb.before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Raw().After("*").
		Register("prometheus_raw_after", pcb.after("raw"))
	if err != nil {
		panic(err)
	}

	err = db.Callback().Row().Before("*").
		Register("prometheus_row_before", pcb.before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Row().After("*").
		Register("prometheus_row_after", pcb.after("row"))
	if err != nil {
		panic(err)
	}
}

func (c *Callbacks) before() func(db *gorm.DB) {
	return func(db *gorm.DB) {
		startTime := time.Now()
		db.Set("start_time", startTime)
	}
}

func (c *Callbacks) after(typ string) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		val, _ := db.Get("start_time")
		startTime, ok := val.(time.Time)
		if !ok {
			// 你啥都干不了
			return
		}
		table := db.Statement.Table
		if table == "" {
			table = "unknown"
		}
		c.vector.WithLabelValues(typ, table).
			Observe(float64(time.Since(startTime).Milliseconds()))
	}
}

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(msg string, args ...interface{}) {
	g(msg, logger.Field{Key: "args", Value: args})
}

type DoSomething interface {
	DoABC() string
}

type DoSomethingFunc func() string

func (d DoSomethingFunc) DoABC() string {
	return d()
}
