package frame

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/imroc/req/v3"
	"github.com/normastars/frame/core"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func getConfig(configPath ...string) (*ConfigManager, *Config) {
	cm, cf := LoadConfig(configPath...)
	return cm, cf
}

// activeCore holds the currently active application core instance.
// Set once during New() call, used only for backward-compatible functions (GetMySQLConn, GetRedisConn, RegisterTable).
var (
	activeCore   *coreApp
	activeCoreMu sync.Mutex
)

func setActiveCore(c *coreApp) {
	activeCoreMu.Lock()
	defer activeCoreMu.Unlock()
	activeCore = c
}

func getActiveCore() *coreApp {
	activeCoreMu.Lock()
	defer activeCoreMu.Unlock()
	return activeCore
}

// App frame engine
type App struct {
	*gin.Engine
	core  *coreApp
	log   *logrus.Logger
	Entry *logrus.Entry
}

// New creates an App. Arguments can be config file paths or Options.
//
//	app := frame.New()                     // default config
//	app := frame.New("./conf/dev.json")    // custom config path
//	app := frame.New(frame.WithMockDB(db)) // mock injection for tests
func New(configs ...interface{}) *App {
	gin.DefaultWriter = io.Discard

	var configPath []string
	var opts []core.Option

	for _, arg := range configs {
		switch v := arg.(type) {
		case string:
			configPath = append(configPath, v)
		case core.Option:
			opts = append(opts, v)
		}
	}

	SetDefaultLog()
	cm, ac := getConfig(configPath...)
	logger := NewLogger(ac)

	registerMetrics()

	coreApp := newCoreApp(ac, cm, opts...)
	setActiveCore(coreApp)

	e := &App{
		Engine: coreApp.engine,
		core:   coreApp,
		log:    logger,
	}
	e.Entry = e.log.WithField(TraceIDKey, generateTraceID(ac.Project))

	if coreApp.config.HTTPServer.EnableCors {
		e.Use(CORSFunc(coreApp.config.HTTPServer.CorsOrigins...))
	}
	e.Use(TraceFunc())
	e.Use(LoggerFunc())

	// Set custom 404/405 handler that returns JSON so LoggerFunc logs it.
	e.Engine.NoRoute(func(c *gin.Context) {
		ctx := e.createContext(c)
		ctx.HTTPError2(http.StatusNotFound, "ROUTE_NOT_FOUND", "route not found", fmt.Errorf("no route: %s %s", c.Request.Method, c.Request.URL.Path))
	})
	e.Engine.NoMethod(func(c *gin.Context) {
		ctx := e.createContext(c)
		ctx.HTTPError2(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", fmt.Errorf("method not allowed: %s %s", c.Request.Method, c.Request.URL.Path))
	})

	coreApp.autoMigrateTables()

	return e
}

// NewLogEntry reinitializes the Entry field with a new trace ID.
// Deprecated: Entry is automatically initialized in New().
func (e *App) NewLogEntry() {
	e.Entry = e.log.WithField(TraceIDKey, generateTraceID(e.core.config.Project))
}

// Run starts the server with graceful shutdown support.
func (e *App) Run() error {
	if !e.core.config.HTTPServer.Enable {
		return nil
	}

	var metricServer *http.Server
	if e.core.config.EnableMetric {
		metricServer = e.startMetricServer()
	}

	srv := &http.Server{
		Addr:    e.core.config.getServerPort(),
		Handler: e.Engine,
	}

	go func() {
		logrus.Infof("server listen %s\n", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("server error: %v", err)
		}
	}()

	quit, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-quit.Done()

	logrus.Info("shutting down server...")

	// Use a shared context for shutdown; each Shutdown call is best-effort.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logrus.Errorf("server forced to shutdown: %v", err)
	}
	if metricServer != nil {
		if err := metricServer.Shutdown(ctx); err != nil {
			logrus.Errorf("metric server forced to shutdown: %v", err)
		}
	}
	logrus.Info("server exited")
	return nil
}

func (e *App) startMetricServer() *http.Server {
	port := e.core.config.getMetricPort()
	mux := http.NewServeMux()
	mux.Handle(defaultMetricPath, promhttp.Handler())
	srv := &http.Server{Addr: port, Handler: mux}
	go func() {
		logrus.Infof("%s server listen %s\n", defaultMetricName, port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("metric server error: %v", err)
		}
	}()
	return srv
}

// getTraceID extracts the trace ID from the gin request header.
func (e *App) getTraceID(c *gin.Context) string {
	return c.GetHeader(TraceIDKey)
}

func (e *App) getLogEntry(c *gin.Context) *logrus.Entry {
	return e.log.WithField(TraceIDKey, e.getTraceID(c))
}

func (e *App) getHTTPClient(c *gin.Context) *req.Client {
	traceID := e.getTraceID(c)
	return e.core.baseHTTPClient.SetCommonHeader(TraceIDKey, traceID)
}

// getHTTPClient creates an HTTP client for non-Gin contexts.
func getHTTPClient(conf *Config, traceID ...string) *req.Client {
	tid := ""
	if len(traceID) > 0 && traceID[0] != "" {
		tid = traceID[0]
	} else {
		tid = generateTraceID(conf.Project)
	}
	rc := req.C()
	if !conf.HTTPClient.DisableReqLog {
		rc = rc.OnAfterResponse(ReqLogMiddleware)
	}
	if conf.HTTPClient.EnableMetric {
		rc = rc.OnAfterResponse(ReqMetricMiddleware)
	}
	rc = rc.SetCommonHeader(TraceIDKey, tid)
	return rc
}

func defaultEngine() *gin.Engine {
	return gin.Default()
}

// registerRouteWithContext creates a gin handler that builds a frame Context
// in a closure, avoiding per-route-type boilerplate.
func (e *App) registerRouteWithContext(handler func(c *Context)) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler(e.createContext(c))
	}
}

func (e *App) createContext(c *gin.Context) *Context {
	traceID := e.getTraceID(c)
	return &Context{
		Gtx:           c,
		config:        e.core.config,
		configManager: e.core.configManager,
		coreApp:       e.core,
		Entry:         e.log.WithField(TraceIDKey, traceID),
		httpClient:    e.core.baseHTTPClient.SetCommonHeader(TraceIDKey, traceID),
		gormLogger:    e.core.gormLogger,
		redisHook:     e.core.redisHook,
	}
}

// NewContextNoGin returns a context without a gin context (for non-HTTP use cases).
func NewContextNoGin(configPath ...string) *Context {
	cm, c := getConfig(configPath...)
	traceID := generateTraceID(c.Project)
	core := newCoreApp(c, cm)
	return &Context{
		config:        c,
		configManager: cm,
		coreApp:       core,
		Entry:         NewLogger(c).WithField(TraceIDKey, traceID),
		httpClient:    getHTTPClient(c, traceID),
		gormLogger:    core.gormLogger,
		redisHook:     core.redisHook,
	}
}

// GET registers a GET route.
func (e *App) GET(relativePath string, handler func(c *Context)) {
	e.Engine.GET(relativePath, e.registerRouteWithContext(handler))
}

// POST registers a POST route.
func (e *App) POST(relativePath string, handler func(c *Context)) {
	e.Engine.POST(relativePath, e.registerRouteWithContext(handler))
}

// PUT registers a PUT route.
func (e *App) PUT(relativePath string, handler func(c *Context)) {
	e.Engine.PUT(relativePath, e.registerRouteWithContext(handler))
}

// PATCH registers a PATCH route.
func (e *App) PATCH(relativePath string, handler func(c *Context)) {
	e.Engine.PATCH(relativePath, e.registerRouteWithContext(handler))
}

// DELETE registers a DELETE route.
func (e *App) DELETE(relativePath string, handler func(c *Context)) {
	e.Engine.DELETE(relativePath, e.registerRouteWithContext(handler))
}

// HEAD registers a HEAD route.
func (e *App) HEAD(relativePath string, handler func(c *Context)) {
	e.Engine.HEAD(relativePath, e.registerRouteWithContext(handler))
}

func getTraceIDFromContext(ctx context.Context) string {
	traceID := ctx.Value(TraceIDKey)
	if traceID == nil {
		return ""
	}
	if id, ok := traceID.(string); ok {
		return id
	}
	return ""
}
