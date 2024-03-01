package frame

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var mysqlOnce sync.Once

// Dialect defines dialect for mysql
const Dialect = "mysql"

// DBMultiClient multi db conns
type DBMultiClient struct {
	clients map[string]*gorm.DB
}

var dbMultiConn = &DBMultiClient{
	clients: map[string]*gorm.DB{},
}

// GetMySQLConn return mysql client list
func GetMySQLConn() *DBMultiClient {
	return dbMultiConn
}

func newMySQLServers(conf *Config) {
	mysqlOnce.Do(func() {
		if len(conf.Mysql.Configs) > 0 && conf.Mysql.Enable {
			for _, v := range conf.Mysql.Configs {
				conn := open(conf.LogLevel, conf.LogMode, v)
				if conn != nil {
					// add connection map
					dbMultiConn.clients[v.Name] = conn
				}
			}
		}
	})
}

func open(logLevel, logMode string, item MySQLConfigItem) *gorm.DB {
	if !item.Enable {
		return nil
	}
	dsn := fmt.Sprintf("%s:%s@(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", item.User, item.Password, item.Host, item.Database)
	// Initialize a new logger instance
	l := newLoggerLevel(logLevel, logMode)
	// Set the GORM logger to the new logger instance
	slowSec := 0
	if item.SlowThresholdSec > 0 {
		slowSec = item.SlowThresholdSec
	}
	dbLogger := logger.New(
		l,
		logger.Config{
			SlowThreshold:             time.Duration(slowSec) * time.Second,
			LogLevel:                  log2gormLevel(logLevel),
			IgnoreRecordNotFoundError: true,
		},
	)
	dbConn, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: dbLogger})
	if err == nil {
		return dbConn
	}

	if item.EnableAutoMigrate && strings.Contains(err.Error(), "Unknown database") {
		// auto migrate database
		err = createDatabase(item.User, item.Password, item.Host, item.Database)
		if err != nil {
			logrus.Fatalln(err)
		}
		// retry connection mysql
		dbConn, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			logrus.Fatalln(err)
		}
		return dbConn
	}
	logrus.Fatalln(err)
	return nil
}

func createDatabase(user, password, host, database string) error {
	db, err := sql.Open(Dialect, fmt.Sprintf("%s:%s@(%s)/", user, password, host))
	if err != nil {
		return err
	}
	_, err = db.Exec(createDatabaseSQL(database))
	if err != nil {
		return err
	}
	db.Close()
	return nil
}

// createDatabaseSQL create database
func createDatabaseSQL(database string) string {
	return fmt.Sprintf("CREATE DATABASE `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;", database)
}

type gormLogger struct {
	Log     *logrus.Logger
	Disable bool
}

func newGormLogger(config *Config) logger.Interface {
	return &gormLogger{Log: NewLogger(config), Disable: config.Mysql.DisableReqLog}
}

func (l *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

func (l *gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.Disable {
		return
	}
	l.Log.WithFields(logrus.Fields{
		TraceIDKey: getTraceIDFromContext(ctx),
	}).Infof(msg, data...)
}

func (l *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.Disable {
		return
	}
	l.Log.WithFields(logrus.Fields{
		TraceIDKey: getTraceIDFromContext(ctx),
	}).Warnf(msg, data...)
}

func (l *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.Disable {
		return
	}
	l.Log.WithFields(logrus.Fields{
		TraceIDKey: getTraceIDFromContext(ctx),
	}).Errorf(msg, data...)
}

func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.Disable {
		return
	}
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			TraceIDKey: getTraceIDFromContext(ctx),
			"duration": time.Since(begin).Milliseconds(),
			"error":    err.Error(),
		}).Error(fc())
	} else {
		l.Log.WithFields(logrus.Fields{
			TraceIDKey: getTraceIDFromContext(ctx),
			"duration": time.Since(begin).Milliseconds(), //
		}).Infoln(fc())
	}
}

// func open(dialects, host, database, user, password string) *gorm.DB {
// 	// func open(dialects, user, pass, host, database string) error {
// 	var (
// 		err    error
// 		dbConn *gorm.DB
// 	)
// 	// log init
// 	// Initialize a new logger instance
// 	l := logrus.New()
// 	l.SetLevel(logrus.DebugLevel)
// 	l.SetFormatter(&logrus.JSONFormatter{})
// 	// Set the GORM logger to the new logger instance
// 	dbLogger := logger.New(
// 		l,
// 		logger.Config{
// 			SlowThreshold: time.Second,
// 			LogLevel:      logger.Info,
// 			Colorful:      false,
// 		},
// 	)
// 	// dbConn, err = gorm.Open(dialects, fmt.Sprintf("%s:%s@(%s)/%s?charset=utf8&parseTime=True&loc=Local", user, password, host, database))
// 	dsn := fmt.Sprintf("%s:%s@(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, database)
// 	dbConn, err = gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: dbLogger})
// 	if err != nil {
// 		// glog.Errorf("database open error: %s\n", err)
// 		if strings.Contains(err.Error(), "Error 1049: Unknown database") {
// 			db, _ := sql.Open(dialects, fmt.Sprintf("%s:%s@(%s)/", user, password, host))
// 			_, err = db.Exec(createDatabaseSQL(database))
// 			if err != nil {
// 				// glog.Fatalf("create database error: %s", err)
// 			}

// 		}
// 	}

// 	// set max
// 	// dbConn.DB().SetMaxIdleConns()
// 	// dbConn.DB().SetMaxOpenConns()
// 	// dbConn.DB().SetConnMaxIdleTime()
// 	// dbConn.DB().SetConnMaxIdleTime()

// 	return dbConn
// }
