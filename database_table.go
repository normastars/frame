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

// globalPreRegistry holds tables registered before frame.New() is called.
// New() merges it into the coreApp's tableRegistry during startup.
var globalPreRegistry = newDatabaseTableList()

// RegisterTable registers a table model for auto-migration.
//
// It can be called either before or after frame.New():
//   - Before New(): the table is stored in a package-level pre-registry and
//     merged into the active App during New().
//   - After New(): the table is registered directly on the active App.
//
// In both cases autoMigrateTables() (called inside New()) applies the migration.
func RegisterTable(database string, table Table, initfuncs ...TableInitFunc) {
	if c := getActiveCore(); c != nil {
		c.tableRegistry.Add(database, table, initfuncs...)
		return
	}
	// No active core yet — buffer in the pre-registry so New() can pick it up.
	globalPreRegistry.Add(database, table, initfuncs...)
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

// MergeFrom copies all entries from other into tl, skipping duplicates.
// Used by New() to absorb the globalPreRegistry.
func (tl *databaseTableList) MergeFrom(other *databaseTableList) {
	tables := other.GetTables()
	for dbName, tasks := range tables {
		for _, task := range tasks {
			tl.Add(dbName, task.Model, task.InitFuncs...)
		}
	}
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
