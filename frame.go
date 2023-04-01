package frame

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

// App frame engine
type App struct {
	*gin.Engine
	config       *Config
	dbClients    *DBMultiClient
	redisClients *RedisMultiClient
	log          *logrus.Logger
	*logrus.Entry
}

// Run engin run
// ":8080"
func (e *App) Run() error {
	go e.metricRun()
	return e.serverRun()
}

func (e *App) metricRun() {
	if e.config.EnableMetric && e.config.HTTPServer.Enable {
		// metrics
		port := e.getMetricPort()
		fmt.Printf("%s server listen %s\n", defaultMetricName, port)
		http.Handle(defaultMetricPath, promhttp.Handler())
		if err := http.ListenAndServe(port, nil); err != nil {
			fmt.Println(err.Error())
			os.Exit(-1)
		}
	}
}
func (e *App) serverRun() error {
	if e.config.HTTPServer.Enable {
		// server port
		port := e.getServerPort()
		fmt.Printf("server listen %s\n", port)
		if err := e.Engine.Run(port); err != nil {
			fmt.Println(err.Error())
			os.Exit(-1)
		}
	}
	return nil
}

// NewLogEntry new log entry
func (e *App) NewLogEntry() {
	e.Entry = e.log.WithField(TraceIDKey, generalTraceID(e.config.Project))
}

func (e *App) getServerPort() string {
	return e.config.getServerPort()
}

func (e *App) getMetricPort() string {
	return e.config.getMetricPort()
}

// New engin
func New() *App {
	// close gin log
	gin.DefaultWriter = ioutil.Discard
	e := newApp()
	return e
}

// newApp new engine
func newApp() *App {
	// default log
	SetDefaultLog()

	// step 1:  config
	conf := LoadConfig()

	// step 2:  log
	logger := NewLogger(conf)

	// step 3: mysql
	newMySQLServers(conf)
	mysqlConns := GetMySQLConn()

	// step 4: redis
	newRedisServers(conf)
	redisConns := GetRedisConn()

	e := &App{
		Engine:       defaultEngine(),
		log:          logger,
		config:       conf,
		dbClients:    mysqlConns,
		redisClients: redisConns,
	}
	// common trace id
	e.NewLogEntry()
	if e.config.HTTPServer.EnableCors {
		e.Use(CORSFunc())
	}
	e.Use(TraceFunc())
	e.Use(LoggerFunc())
	return e
}

func defaultEngine() *gin.Engine {
	r := gin.Default()
	return r
}

func (e *App) createContext(c *gin.Context) *Context {
	traceID := c.GetHeader(TraceIDKey)
	l := e.log.WithField(TraceIDKey, traceID)
	return &Context{
		Gtx:          c,
		config:       e.config,
		redisClients: e.redisClients,
		dbClients:    e.dbClients,
		Entry:        l,
	}
}

// GET get method
func (e *App) GET(relativePath string, handler func(c *Context)) {
	e.Engine.GET(relativePath, func(c *gin.Context) {
		ctx := e.createContext(c)
		handler(ctx)
	})
}

// POST post method
func (e *App) POST(relativePath string, handler func(c *Context)) {
	e.Engine.POST(relativePath, func(c *gin.Context) {
		ctx := e.createContext(c)
		handler(ctx)
	})
}

// PUT http put method
func (e *App) PUT(relativePath string, handler func(c *Context)) {
	e.Engine.PUT(relativePath, func(c *gin.Context) {
		ctx := e.createContext(c)
		handler(ctx)
	})
}

// PATCH  http patch method
func (e *App) PATCH(relativePath string, handler func(c *Context)) {
	e.Engine.PATCH(relativePath, func(c *gin.Context) {
		ctx := e.createContext(c)
		handler(ctx)
	})
}

// DELETE http delete method
func (e *App) DELETE(relativePath string, handler func(c *Context)) {
	e.Engine.DELETE(relativePath, func(c *gin.Context) {
		ctx := e.createContext(c)
		handler(ctx)
	})
}

// HEAD http head method
func (e *App) HEAD(relativePath string, handler func(c *Context)) {
	e.Engine.HEAD(relativePath, func(c *gin.Context) {
		ctx := e.createContext(c)
		handler(ctx)
	})
}

func getTraceIDFromContext(ctx context.Context) string {
	traceID := ctx.Value(TraceIDKey)
	return traceID.(string)
}
