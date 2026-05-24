package frame

import (
	"sync"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Table define database table interface.
type Table interface {
	TableName() string
}

// RegisterTable registers a table model for auto-migration.
// Must be called before frame.New() to take effect.
func RegisterTable(database string, table Table, initfuncs ...TableInitFunc) {
	core := getActiveCore()
	if core == nil {
		logrus.Errorf("RegisterTable: no active app, cannot register table %s", table.TableName())
		return
	}
	core.tableRegistry.Add(database, table, initfuncs...)
}

// TableInitFunc is a function called after a table is auto-migrated.
type TableInitFunc func(conn *gorm.DB) error

type tableInitTask struct {
	Model     Table
	InitFuncs []TableInitFunc
}

// databaseTableList holds registered table definitions, keyed by database name.
// Thread-safe: reads use RLock, writes use Lock.
type databaseTableList struct {
	mu sync.RWMutex
	m  map[string][]tableInitTask
}

func newDatabaseTableList() *databaseTableList {
	return &databaseTableList{m: make(map[string][]tableInitTask)}
}

// Add registers a table under the given database name.
// Duplicate table names are silently ignored.
func (tl *databaseTableList) Add(database string, table Table, initfuncs ...TableInitFunc) {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	entry := tableInitTask{Model: table, InitFuncs: initfuncs}
	tableName := table.TableName()

	if existing, ok := tl.m[database]; ok {
		for _, t := range existing {
			if t.Model.TableName() == tableName {
				return // already registered, skip
			}
		}
		tl.m[database] = append(existing, entry)
	} else {
		tl.m[database] = []tableInitTask{entry}
	}

	logrus.Infof("table %s registered to %s successfully", tableName, database)
}

// GetTables returns a defensive copy of all registered tables.
func (tl *databaseTableList) GetTables() map[string][]tableInitTask {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	result := make(map[string][]tableInitTask, len(tl.m))
	for dbName, tasks := range tl.m {
		copied := make([]tableInitTask, len(tasks))
		copy(copied, tasks)
		result[dbName] = copied
	}
	return result
}
