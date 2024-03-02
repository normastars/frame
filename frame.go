package frame

import (
	"context"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imroc/req/v3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

var (
	defaultLogLevel = "info"
	defaultLogMode  = "text"
	initLoadConf    = 0 // 第一次加载日志配置
)

func getLogConf() *Config {
	return &Config{
		LogLevel: defaultLogLevel,
		LogMode:  defaultLogMode,
	}
}

func getConfig(configPath ...string) *Config {
	cf := LoadConfig(configPath...)
	if initLoadConf == 0 {
		defaultLogLevel = cf.LogLevel
		defaultLogMode = cf.LogMode
		initLoadConf = 1
	}

	return cf
}

// App frame engine
type App struct {
	*gin.Engine
	config       *Config
	dbClients    *DBMultiClient
	redisClients *RedisMultiClient
	log          *logrus.Logger
	*logrus.Entry
}

// New engin
func New(configPath ...string) *App {
	// close gin log
	gin.DefaultWriter = ioutil.Discard
	e := newApp(configPath...)
	return e
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
		logrus.Infof("%s server listen %s\n", defaultMetricName, port)
		http.Handle(defaultMetricPath, promhttp.Handler())
		if err := http.ListenAndServe(port, nil); err != nil {
			logrus.Fatalln(err.Error())
		}
	}
}
func (e *App) serverRun() error {
	if e.config.HTTPServer.Enable {
		// server port
		port := e.getServerPort()
		logrus.Infof("server listen %s\n", port)
		if err := e.Engine.Run(port); err != nil {
			logrus.Fatalln(err.Error())
		}
	}
	return nil
}

// NewLogEntry new log entry
func (e *App) NewLogEntry() {
	e.Entry = e.log.WithField(TraceIDKey, generateTraceID(e.config.Project))
}

func (e *App) getServerPort() string {
	return e.config.getServerPort()
}

func (e *App) getMetricPort() string {
	return e.config.getMetricPort()
}

// newApp new engine
func newApp(configPath ...string) *App {
	// default log
	SetDefaultLog()

	// step 1:  config
	ac := getConfig(configPath...)

	// step 2:  log
	logger := NewLogger(ac)

	// step 3: mysql
	newMySQLServers(ac)
	mysqlConns := GetMySQLConn()

	// step 4: redis
	newRedisServers(ac)
	redisConns := GetRedisConn()

	e := &App{
		Engine:       defaultEngine(),
		log:          logger,
		config:       ac,
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

	// table auto migrate
	e.autoMigrateMysql(configPath...)
	return e
}

func defaultEngine() *gin.Engine {
	r := gin.Default()
	return r
}

func (e *App) createContext(c *gin.Context) *Context {
	// set http client
	return &Context{
		Gtx:          c,
		config:       e.config,
		redisClients: e.redisClients,
		dbClients:    e.dbClients,
		Entry:        e.getLogEntry(c),
		httpClient:   e.getHTTPClient(c),
	}
}

// NewContextNoGin return context but no include gin context
func NewContextNoGin(configPath ...string) *Context {
	c := getConfig(configPath...)
	traceID := generateTraceID(c.Project)
	return &Context{
		config:       c,
		redisClients: GetRedisConn(),
		dbClients:    GetMySQLConn(),
		Entry:        NewLogger(c).WithField(TraceIDKey, traceID),
		httpClient:   getHTTPClient(c, traceID),
	}
}

func (e *App) autoMigrateMysql(configPath ...string) {
	c := NewContextNoGin(configPath...)
	tablesInit(c)

}

func (e *App) getTraceID(c *gin.Context) string {
	return c.GetHeader(TraceIDKey)
}

func (e *App) getLogEntry(c *gin.Context) *logrus.Entry {
	return e.log.WithField(TraceIDKey, e.getTraceID(c))
}

func (e *App) getHTTPClient(c *gin.Context) *req.Client {
	traceID := c.GetHeader(TraceIDKey)
	return getHTTPClient(e.config, traceID)
}

func getHTTPClient(conf *Config, traceID ...string) *req.Client {
	tid := ""
	if len(traceID) <= 0 {
		tid = generateTraceID(conf.Project)
	} else {
		tid = traceID[0]
	}
	rc := req.C()
	rc = rc.SetCommonHeader(TraceIDKey, tid)
	if !conf.HTTPClient.DisableReqLog {
		rc = rc.OnAfterResponse(ReqLogMiddleware)
	}
	if conf.HTTPClient.EnableMetric {
		rc = rc.OnAfterResponse(ReqMetricMiddleware)
	}
	return rc
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
