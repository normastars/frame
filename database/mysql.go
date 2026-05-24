// Package database provides MySQL connection management.
// Instance-level Manager with lazy initialization, eliminating global state.
package database

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/normastars/frame/internal"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// validDBName enforces a strict allowlist for database names to prevent SQL injection
// in the CREATE DATABASE statement (which cannot use parameterised queries).
var validDBName = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

// Dialect defines dialect for mysql
const Dialect = "mysql"

// Manager manages MySQL connections (instance-level, concurrency-safe)
type Manager struct {
	configs   []internal.DBConfig
	clients   map[string]*gorm.DB
	mu        sync.RWMutex
	once      sync.Once
	logLevel  string
	logMode   string
	gormLog   logger.Interface
	initErr   error
}

// NewManager creates a new MySQL connection manager.
// Does not connect immediately; initializes lazily on first GetDB call.
func NewManager(configs []internal.DBConfig, logLevel, logMode string, gormLog logger.Interface) *Manager {
	return &Manager{
		configs:  configs,
		clients:  make(map[string]*gorm.DB),
		logLevel: logLevel,
		logMode:  logMode,
		gormLog:  gormLog,
	}
}

// GetDB returns a MySQL connection. Lazy initialized on first call.
func (m *Manager) GetDB(name ...string) *gorm.DB {
	m.initConnections()
	if m.initErr != nil {
		logrus.Errorf("database: init error: %v", m.initErr)
		return nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.clients) == 1 && len(name) == 0 {
		for _, v := range m.clients {
			return v
		}
	}
	if len(name) == 0 {
		logrus.Error("database: GetDB called with empty name but multiple clients exist")
		return nil
	}
	return m.clients[name[0]]
}

// initConnections lazy initializes all database connections.
// Uses sync.Once to ensure single initialization.
// The inner loop runs exclusively inside once.Do, so no additional locking is
// needed while writing to m.clients — the RWMutex is only needed for concurrent
// reads that happen after initialization completes.
func (m *Manager) initConnections() {
	m.once.Do(func() {
		if len(m.configs) == 0 {
			return
		}
		var errs []error
		for _, v := range m.configs {
			if !v.Enable {
				continue
			}
			conn, err := m.open(v)
			if err != nil {
				logrus.Errorf("database: failed to connect %s: %v", v.Name, err)
				errs = append(errs, err)
				continue
			}
			m.clients[v.Name] = conn
		}
		if len(errs) > 0 {
			m.initErr = errors.Join(errs...)
		}
	})
}

func (m *Manager) open(item internal.DBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		item.User, item.Password, item.Host, item.Database)

	slowSec := time.Duration(item.SlowThresholdSec) * time.Second
	dbLogger := logger.New(
		logrus.New(),
		logger.Config{
			SlowThreshold:             slowSec,
			IgnoreRecordNotFoundError: true,
		},
	)
	if m.gormLog != nil {
		dbLogger = m.gormLog
	}

	dbConn, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: dbLogger})
	if err == nil {
		configDBPool(dbConn)
		return dbConn, nil
	}

	if item.EnableAutoMigrate && strings.Contains(err.Error(), "Unknown database") {
		if dbErr := createDatabase(item.User, item.Password, item.Host, item.Database); dbErr != nil {
			return nil, fmt.Errorf("create database %s: %w", item.Database, dbErr)
		}
		dbConn, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("retry connect after create database: %w", err)
		}
		configDBPool(dbConn)
		return dbConn, nil
	}

	return nil, fmt.Errorf("connect mysql %s: %w", item.Host, err)
}

func createDatabase(user, password, host, database string) error {
	// Validate the database name against an allowlist before using it in raw SQL.
	// The CREATE DATABASE statement cannot use parameterised queries, so we
	// must reject any name that could carry a SQL injection payload.
	if !validDBName.MatchString(database) {
		return fmt.Errorf("database: invalid database name %q (only [A-Za-z0-9_] allowed)", database)
	}
	db, err := sql.Open(Dialect, fmt.Sprintf("%s:%s@(%s)/", user, password, host))
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(createDatabaseSQL(database))
	return err
}

func createDatabaseSQL(database string) string {
	return fmt.Sprintf("CREATE DATABASE `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;", database)
}

func configDBPool(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		logrus.Warnf("database: failed to get underlying sql.DB: %v", err)
		return
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
}
