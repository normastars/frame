package frame

import (
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// TODO LIST:
// 1. gin.Context -> frame.Context
// 2. 集成 Config
// 3. 优化 Redis,Mysql 连接,日志打印
// 4. 跑通demo
// 5. 集成 req 请求
// 6. 优化 Log 包

// Default engin
func Default() *Engine {
	// 关闭Gin的日志输出
	gin.DefaultWriter = ioutil.Discard
	e := &Engine{
		Engine: defaultEngine(),
		DB:     newGrom("TODO"),
	}
	return e

}

func defaultEngine() *gin.Engine {
	r := gin.Default()
	r.Use(TraceFunc())
	r.Use(LoggerFunc())
	return r
}

// Engine frame engine
type Engine struct {
	*gin.Engine
	DB *gorm.DB
}

func newGrom(database string) *gorm.DB {
	// 创建GORM数据库连接
	db, err := gorm.Open(database)
	if err != nil {
		panic(err)
	}
	return db
}
