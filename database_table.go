package frame

import (
	"fmt"
	"sync"

	"gorm.io/gorm"
)

// Table define database table interface.
type Table interface {
	TableName() string
}

// RegisterTable register database table to tablelist.
func RegisterTable(database string, table Table, initfuncs ...tableInitFunc) {
	databaseTables.Add(database, table, initfuncs...)
}

var databaseTables = newDatabaseTableList()

type databaseTableList struct {
	sync.Mutex
	m map[string][]tableInitTask
}

func newDatabaseTableList() *databaseTableList {
	return &databaseTableList{m: make(map[string][]tableInitTask, 0)}
}

func (tl *databaseTableList) Add(database string, table Table, initfuncs ...tableInitFunc) {
	tl.Mutex.Lock()
	defer tl.Mutex.Unlock()
	fmt.Printf("table %s registered to %s successfully", table.TableName(), database)
	v, ok := tl.m[database]
	ti := tableInitTask{Model: table, InitFuncs: initfuncs}
	if !ok {
		tl.m[database] = []tableInitTask{ti}
		return
	}
	find := false
	if len(v) > 0 {
		for _, ta := range v {
			if ta.Model.TableName() == table.TableName() {
				find = true
				break
			}
		}
	}
	if find {
		return
	}
	v = append(v, ti)
	tl.m[database] = v
}

func (tl *databaseTableList) List() *databaseTableList {
	tl.Mutex.Lock()
	defer tl.Mutex.Unlock()
	return databaseTables
}

type tableInitFunc func(conn *gorm.DB) error

// TableInitTask define table init task.
type tableInitTask struct {
	Model     Table
	InitFuncs []tableInitFunc
}

// TablesInit check table status, create or update tables.
func tablesInit(ctx *Context) {
	tables := databaseTables.List()
	if tables == nil || len(tables.m) <= 0 {
		return
	}
	total := 0
	for dbName, tableTasks := range tables.m {
		// config
		if !ctx.config.isEnableMySQLAutoMigrate(dbName) {
			continue
		}
		if len(tableTasks) <= 0 {
			continue
		}
		ctx.Infof("-------------AutoMigrate database: %s begin-------------", dbName)
		conn := ctx.GetDB(dbName)
		for _, v := range tableTasks {
			total = total + 1
			if err := conn.AutoMigrate(v.Model); err != nil {
				ctx.Infof("Database %s table %s auto migrate failed, %s", dbName, v.Model.TableName(), err.Error())
			} else {
				ctx.Infof("Database %s table %s auto migrate successfully", dbName, v.Model.TableName())
			}
			if len(v.InitFuncs) <= 0 {
				continue
			}
			for _, f := range v.InitFuncs {
				if err := f(conn); err != nil {
					ctx.Infof("Database %s table %s init func was executed failed, %s", dbName, v.Model.TableName(), err.Error())
				} else {
					ctx.Infof("Database %s table %s init func was executed successfully", dbName, v.Model.TableName())
				}
			}
		}
		ctx.Infof("-------------AutoMigrate database: %s end-------------", dbName)

	}
	ctx.Infof("a total of %d tables have been checked", total)
}
