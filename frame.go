package frame

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

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
	if e.config.EnableMetric {
		// metrics
		port := e.getMetricPort()
		http.Handle("/metrics", promhttp.Handler())
		e.Infof("metric server listen %s", port)
		http.ListenAndServe(port, nil)
	}
}
func (e *App) serverRun() error {
	if e.config.HTTPServer.Enable {
		// server port
		port := e.getServerPort()
		e.Engine.Run(port)
		e.Infof("server listen %s", port)
		return http.ListenAndServe(port, nil)
	}
	return nil
}

// NewLogEntry new log entry
func (e *App) NewLogEntry() {
	e.Entry = e.log.WithField(TraceIDKey, generalTraceID(e.config.Project))
}

func (e *App) getServerPort() string {
	port := ""
	if len(e.config.HTTPServer.Configs) > 0 {
		for i := range e.config.HTTPServer.Configs {
			if e.config.HTTPServer.Configs[i].Name != "metric" && e.config.HTTPServer.Configs[i].Name != "metrics" {
				port = e.config.HTTPServer.Configs[i].Port
			}
		}
	}
	if len(port) <= 0 {
		port = "8080"
	}
	if !strings.HasPrefix(port, ":") {
		port = fmt.Sprintf(":%s", port)
	}
	return port
}

func (e *App) getMetricPort() string {
	port := ""
	if len(e.config.HTTPServer.Configs) > 0 {
		for i := range e.config.HTTPServer.Configs {
			if e.config.HTTPServer.Configs[i].Name == "metric" || e.config.HTTPServer.Configs[i].Name == "metrics" {
				port = e.config.HTTPServer.Configs[i].Port
			}
		}
	}
	if len(port) <= 0 {
		port = "9090"
	}
	if !strings.HasPrefix(port, ":") {
		port = fmt.Sprintf(":%s", port)
	}
	return port
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
