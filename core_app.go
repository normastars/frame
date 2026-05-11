package frame

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/imroc/req/v3"
	"github.com/normastars/frame/cache"
	"github.com/normastars/frame/core"
	"github.com/normastars/frame/database"
	"github.com/normastars/frame/logger"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	gormLog "gorm.io/gorm/logger"
)

// coreApp is the internal application core, holding all instance-level state.
// frame.App is a thin compatibility layer, delegating all methods to coreApp.
type coreApp struct {
	config        *Config
	configManager *ConfigManager

	// Sub-package managers (instance-level, no global state)
	dbManager    *database.Manager
	cacheManager *cache.Manager
	logManager   *logger.Manager

	// Shared resources
	baseHTTPClient *req.Client
	gormLogger     gormLog.Interface
	redisHook      redis.Hook

	// Table registry (instance-level, replaces global databaseTables)
	tableRegistry *databaseTableList

	// Backward-compatible connection holders (replaces global dbMultiConn/redisMultiConn)
	legacyDBConns    *DBMultiClient
	legacyRedisConns *RedisMultiClient

	// HTTP engine
	engine *gin.Engine

	// Optional mock injections (from Functional Options)
	mockDB    *core.MockDBHolder
	mockRedis *redis.Client
}

// newCoreApp creates a new internal core instance.
// Does not block on DB/Redis connections; lazy initialized on first use.
func newCoreApp(config *Config, configManager *ConfigManager, opts ...core.Option) *coreApp {
	options := &core.Options{}
	for _, opt := range opts {
		opt(options)
	}

	c := &coreApp{
		config:        config,
		configManager: configManager,
		engine:        defaultEngine(),
		gormLogger:    newGormLogger(config).LogMode(log2gormLevel(config.LogLevel)),
		redisHook:     newRedisLogHook(config),
		tableRegistry: newDatabaseTableList(),
		baseHTTPClient: func() *req.Client {
			rc := req.C()
			if !config.HTTPClient.DisableReqLog {
				rc = rc.OnAfterResponse(ReqLogMiddleware)
			}
			if config.HTTPClient.EnableMetric {
				rc = rc.OnAfterResponse(ReqMetricMiddleware)
			}
			return rc
		}(),
		logManager: logger.NewManager(config.LogLevel, config.LogMode),
		mockDB:     options.MockDB,
		mockRedis:  options.MockRedis,
	}

	// Create sub-package managers (save config only, no connections yet)
	c.dbManager = database.NewManager(
		convertDBConfigs(config),
		config.LogLevel, config.LogMode,
		c.gormLogger,
	)
	c.cacheManager = cache.NewManager(convertRedisConfigs(config))

	return c
}

// autoMigrateTables auto-migrates registered tables
func (c *coreApp) autoMigrateTables() {
	c.ensureLegacyDBConns()
	c.ensureLegacyRedisConns()

	tables := c.tableRegistry.GetTables()
	if len(tables) == 0 {
		return
	}
	total := 0
	for dbName, tableTasks := range tables {
		if !c.config.isEnableMySQLAutoMigrate(dbName) {
			continue
		}
		if len(tableTasks) == 0 {
			continue
		}
		logrus.Infof("-------------AutoMigrate database: %s begin-------------", dbName)
		conn := c.dbManager.GetDB(dbName)
		if conn == nil {
			logrus.Errorf("AutoMigrate: db %s not found, skip", dbName)
			continue
		}
		for _, v := range tableTasks {
			total++
			if err := conn.AutoMigrate(v.Model); err != nil {
				logrus.Infof("Database %s table %s auto migrate failed, %s", dbName, v.Model.TableName(), err.Error())
			} else {
				logrus.Infof("Database %s table %s auto migrate successfully", dbName, v.Model.TableName())
			}
			if len(v.InitFuncs) == 0 {
				continue
			}
			for _, f := range v.InitFuncs {
				if err := f(conn); err != nil {
					logrus.Infof("Database %s table %s init func failed, %s", dbName, v.Model.TableName(), err.Error())
				} else {
					logrus.Infof("Database %s table %s init func executed successfully", dbName, v.Model.TableName())
				}
			}
		}
		logrus.Infof("-------------AutoMigrate database: %s end-------------", dbName)
	}
	logrus.Infof("a total of %d tables have been checked", total)
}

// ensureLegacyDBConns populates backward-compatible DB connections
func (c *coreApp) ensureLegacyDBConns() {
	if c.legacyDBConns == nil {
		c.legacyDBConns = &DBMultiClient{clients: make(map[string]*gorm.DB)}
	}
	if c.config != nil && c.config.Mysql.Enable {
		for _, v := range c.config.Mysql.Configs {
			if v.Enable {
				if db := c.dbManager.GetDB(v.Name); db != nil {
					c.legacyDBConns.clients[v.Name] = db
				}
			}
		}
	}
}

// ensureLegacyRedisConns populates backward-compatible Redis connections
func (c *coreApp) ensureLegacyRedisConns() {
	if c.legacyRedisConns == nil {
		c.legacyRedisConns = &RedisMultiClient{clients: make(map[string]*redis.Client)}
	}
	if c.config != nil && c.config.Redis.Enable {
		for _, v := range c.config.Redis.Configs {
			if v.Enable {
				if rc := c.cacheManager.GetRedis(v.Name); rc != nil {
					c.legacyRedisConns.clients[v.Name] = rc
				}
			}
		}
	}
}

// convertDBConfigs converts root Config to core package data types
func convertDBConfigs(config *Config) []core.DBConfig {
	if !config.Mysql.Enable || len(config.Mysql.Configs) == 0 {
		return nil
	}
	result := make([]core.DBConfig, 0, len(config.Mysql.Configs))
	for _, v := range config.Mysql.Configs {
		result = append(result, core.DBConfig{
			Name:              v.Name,
			Enable:            v.Enable,
			EnableAutoMigrate: v.EnableAutoMigrate,
			Host:              v.Host,
			Database:          v.Database,
			User:              v.User,
			Password:          v.Password,
			SlowThresholdSec:  v.SlowThresholdSec,
			DisableReqLog:     config.Mysql.DisableReqLog,
		})
	}
	return result
}

func convertRedisConfigs(config *Config) []core.RedisConfig {
	if !config.Redis.Enable || len(config.Redis.Configs) == 0 {
		return nil
	}
	result := make([]core.RedisConfig, 0, len(config.Redis.Configs))
	for _, v := range config.Redis.Configs {
		result = append(result, core.RedisConfig{
			Name:          v.Name,
			Enable:        v.Enable,
			Host:          v.Host,
			PoolSize:      v.PoolSize,
			Password:      v.Password,
			DB:            v.DB,
			DisableReqLog: config.Redis.DisableReqLog,
		})
	}
	return result
}
