package frame

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var mysqlOnce sync.Once

// DBMultiClient multi db conns
type DBMultiClient struct {
	clients map[string]*gorm.DB
}

var dbMultiConn = &DBMultiClient{
	clients: map[string]*gorm.DB{},
}

// // GetMySQLConns 获取mysql 连接数
// func GetMySQLConns() map[string]*gorm.DB {
// 	return mysqlConns
// }

// func GetMySQLConn(key ...string) *gorm.DB {
// 	if len(mysqlConns) == 1 && len(key) == 0 {
// 		// default mysql connection
// 		for _, v := range mysqlConns {
// 			return v
// 		}
// 	}
// 	return mysqlConns[key[0]]
// }

// Dialect defines dialect for mysql
const Dialect = "mysql"

func openMySQLServers(config MySQLConfig) {
	// 只会初始化一次
	mysqlOnce.Do(func() {
		if len(config.Configs) > 0 && config.Enable {
			for _, v := range config.Configs {
				openMySQL(v)
			}
		}
		return
	})
}

func openMySQL(mysql MySQLConfigItem) {
	conn := open(Dialect, mysql.Host, mysql.Database, mysql.User, mysql.Password)
	if conn != nil {
		// add connection map
		dbMultiConn.clients[mysql.Name] = conn
		// mysqlConns[mysql.Name] = conn
	}
	return
}

// createDatabaseSQL 创建数据库
func createDatabaseSQL(database string) string {
	return fmt.Sprintf("CREATE DATABASE %s CHARACTER SET utf8 COLLATE utf8_general_ci", database)
}

func open(dialects, host, database, user, password string) *gorm.DB {
	// func open(dialects, user, pass, host, database string) error {
	var (
		err    error
		dbConn *gorm.DB
	)
	// dbConn, err = gorm.Open(dialects, fmt.Sprintf("%s:%s@(%s)/%s?charset=utf8&parseTime=True&loc=Local", user, password, host, database))
	dsn := fmt.Sprintf("%s:%s@(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, database)
	dbConn, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		// glog.Errorf("database open error: %s\n", err)
		if strings.Contains(err.Error(), "Error 1049: Unknown database") {
			db, _ := sql.Open(dialects, fmt.Sprintf("%s:%s@(%s)/", user, password, host))
			_, err = db.Exec(createDatabaseSQL(database))
			if err != nil {
				// glog.Fatalf("create database error: %s", err)
			}
			dbConn, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
			if err != nil {
				panic(err)
			}
		}
	}

	// set max
	// dbConn.DB().SetMaxIdleConns()
	// dbConn.DB().SetMaxOpenConns()
	// dbConn.DB().SetConnMaxIdleTime()
	// dbConn.DB().SetConnMaxIdleTime()

	return dbConn
}
