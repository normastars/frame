package frame

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var mysqlOnce sync.Once

// DBMultiClient multi db conns
type DBMultiClient struct {
	clients map[string]*gorm.DB
}

var dbMultiConn = &DBMultiClient{
	clients: map[string]*gorm.DB{},
}

// GetMySQLConn 获取 mysql 连接
func GetMySQLConn() *DBMultiClient {
	return dbMultiConn
}

// Dialect defines dialect for mysql
const Dialect = "mysql"

func newMySQLServers(conf *Config) {
	// 只会初始化一次
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
		return
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
			panic(err)
		}
		// retry connection mysql
		dbConn, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			panic(err)
		}
		return dbConn
	}
	panic(err)
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

// createDatabaseSQL 创建数据库
func createDatabaseSQL(database string) string {
	return fmt.Sprintf("CREATE DATABASE %s CHARACTER SET utf8 COLLATE utf8_general_ci", database)
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