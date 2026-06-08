package frame

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os/signal"
	"sync/atomic"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/imroc/req/v3"
	"github.com/normastars/frame/internal"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

// activeCore holds the currently active application core instance.
// Set once during New() call, used only for backward-compatible functions (GetMySQLConn, GetRedisConn, RegisterTable).
// atomic.Pointer gives lock-free reads and safe single-writer updates.
var activeCore atomic.Pointer[coreApp]

func setActiveCore(c *coreApp) {
	activeCore.Store(c)
}

func getActiveCore() *coreApp {
	return activeCore.Load()
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
	var opts []internal.Option

	for _, arg := range configs {
		switch v := arg.(type) {
		case string:
			configPath = append(configPath, v)
		case internal.Option:
			opts = append(opts, v)
		}
	}

	SetDefaultLog()
	cm, ac := LoadConfig(configPath...)
	logger := NewLogger(ac)

	registerMetrics()

	coreApp := newCoreApp(ac, cm, opts...)
	setActiveCore(coreApp)

	// Absorb tables that were registered before New() was called.
	coreApp.tableRegistry.MergeFrom(globalPreRegistry)

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

	hs := e.core.config.HTTPServer
	srv := &http.Server{
		Addr:         e.core.config.getServerPort(),
		Handler:      e.Engine,
		ReadTimeout:  hs.ReadTimeout(),
		WriteTimeout: hs.WriteTimeout(),
		IdleTimeout:  hs.IdleTimeout(),
	}

	// srvErr carries startup/listen errors out of the goroutine without
	// calling os.Exit (which logrus.Fatalf would do, skipping all defers).
	srvErr := make(chan error, 1)
	go func() {
		logrus.Infof("server listen %s\n", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			srvErr <- err
		}
	}()

	quit, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-quit.Done():
		logrus.Info("shutting down server...")
	case err := <-srvErr:
		logrus.Errorf("server error: %v", err)
		return err
	}

	// Use a shared context for shutdown; each Shutdown call is best-effort.
	ctx, cancel := context.WithTimeout(context.Background(), hs.ShutdownTimeout())
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
	hs := e.core.config.HTTPServer
	mux := http.NewServeMux()
	mux.Handle(defaultMetricPath, promhttp.Handler())
	srv := &http.Server{
		Addr:        port,
		Handler:     mux,
		ReadTimeout: hs.ReadTimeout(),
		IdleTimeout: hs.IdleTimeout(),
	}
	go func() {
		logrus.Infof("%s server listen %s\n", defaultMetricName, port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Errorf("metric server error: %v", err)
		}
	}()
	return srv
}

// getTraceID extracts the trace ID from the gin request header.
func (e *App) getTraceID(c *gin.Context) string {
	return c.GetHeader(TraceIDKey)
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

// defaultEngine returns a bare gin.Engine with no middleware.
// gin.Default() bundles its own Logger + Recovery which would duplicate the
// frame-level LoggerFunc and produce noisy extra log lines.
func defaultEngine() *gin.Engine {
	return gin.New()
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

// NewContextNoGin returns a Context without a gin.Context (for jobs, CLIs, or tests).
// When called without arguments it reuses the active App's coreApp to avoid
// creating a duplicate DB/Redis connection pool on every call.
func NewContextNoGin(configPath ...string) *Context {
	if len(configPath) == 0 {
		if c := getActiveCore(); c != nil {
			traceID := generateTraceID(c.config.Project)
			return &Context{
				config:        c.config,
				configManager: c.configManager,
				coreApp:       c,
				traceID:       traceID,
				Entry:         NewLogger(c.config).WithField(TraceIDKey, traceID),
				httpClient:    getHTTPClient(c.config, traceID),
				gormLogger:    c.gormLogger,
				redisHook:     c.redisHook,
			}
		}
	}
	cm, c := LoadConfig(configPath...)
	traceID := generateTraceID(c.Project)
	ca := newCoreApp(c, cm)
	return &Context{
		config:        c,
		configManager: cm,
		coreApp:       ca,
		traceID:       traceID,
		Entry:         NewLogger(c).WithField(TraceIDKey, traceID),
		httpClient:    getHTTPClient(c, traceID),
		gormLogger:    ca.gormLogger,
		redisHook:     ca.redisHook,
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

// Group creates a route group with optional frame-compatible middleware.
// Handlers registered on the returned RouterGroup receive a full Context
// (config, DB, Redis, HTTP client, logging) — unlike the raw gin.RouterGroup
// returned by the embedded Engine.Group().
func (e *App) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	ginHandlers := make([]gin.HandlerFunc, len(handlers))
	for i, h := range handlers {
		ginHandlers[i] = e.convert2GinHandlerFunc(h)
	}
	return &RouterGroup{
		app: e,
		rg:  e.Engine.Group(relativePath, ginHandlers...),
	}
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
