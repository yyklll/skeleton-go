package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	// Needed for the MySQL driver
	"code.gitea.io/gitea/modules/setting"
	_ "github.com/go-sql-driver/mysql"
	"xorm.io/core"
	"xorm.io/xorm"

	"github.com/yyklll/skeleton/pkg/config"
)

// Engine represents a xorm engine or session.
type Engine interface {
	Table(tableNameOrBean interface{}) *xorm.Session
	Count(...interface{}) (int64, error)
	Decr(column string, arg ...interface{}) *xorm.Session
	Delete(interface{}) (int64, error)
	Exec(...interface{}) (sql.Result, error)
	Find(interface{}, ...interface{}) error
	Get(interface{}) (bool, error)
	ID(interface{}) *xorm.Session
	In(string, ...interface{}) *xorm.Session
	Incr(column string, arg ...interface{}) *xorm.Session
	Insert(...interface{}) (int64, error)
	InsertOne(interface{}) (int64, error)
	Iterate(interface{}, xorm.IterFunc) error
	Join(joinOperator string, tablename interface{}, condition string, args ...interface{}) *xorm.Session
	SQL(interface{}, ...interface{}) *xorm.Session
	Where(interface{}, ...interface{}) *xorm.Session
	Asc(colNames ...string) *xorm.Session
	Limit(limit int, start ...int) *xorm.Session
	SumInt(bean interface{}, columnName string) (res int64, err error)
}

var (
	x      *xorm.Engine
	tables []interface{}

	// HasEngine specifies if we have a xorm.Engine
	HasEngine bool
)

func init() {
	tables = append(tables, new(User))

	gonicNames := []string{"SSL", "UID"}
	for _, name := range gonicNames {
		core.LintGonicMapper[name] = true
	}
}

func getEngine(cfg *config.Database) (*xorm.Engine, error) {
	connStr := cfg.Source

	engine, err := xorm.NewEngine(cfg.Backend, connStr)
	if err != nil {
		return nil, err
	}
	// engine.SetSchema(setting.Database.Schema)
	return engine, nil
}

// InitEngine sets the xorm.Engine
func InitEngine(cfg *config.Database) (err error) {
	x, err = getEngine(cfg)
	if err != nil {
		return fmt.Errorf("Failed to connect to database: %v", err)
	}

	x.ShowExecTime(true)
	x.SetMapper(core.GonicMapper{})
	// WARNING: for serv command, MUST remove the output to os.stdout,
	// so use log file to print to stdout instead.
	x.SetLogger(NewXORMLogger(setting.Database.LogSQL))
	x.ShowSQL(setting.Database.LogSQL)
	x.SetMaxOpenConns(setting.Database.MaxOpenConns)
	x.SetMaxIdleConns(setting.Database.MaxIdleConns)
	x.SetConnMaxLifetime(setting.Database.ConnMaxLifetime)
	return nil
}

// NewEngine initializes a new xorm.Engine
func NewEngine(ctx context.Context, cfg *config.Database, migrateFunc func(*xorm.Engine) error) (err error) {
	if err = InitEngine(cfg); err != nil {
		return err
	}

	x.SetDefaultContext(ctx)

	if err = x.Ping(); err != nil {
		return err
	}

	if err = migrateFunc(x); err != nil {
		return fmt.Errorf("migrate: %v", err)
	}

	if err = x.StoreEngine("InnoDB").Sync2(tables...); err != nil {
		return fmt.Errorf("sync database struct error: %v", err)
	}

	return nil
}

// Ping tests if database is alive
func Ping() error {
	if x != nil {
		return x.Ping()
	}
	return errors.New("database not configured")
}

// DumpDatabase dumps all data from database according the special database SQL syntax to file system.
func DumpDatabase(filePath string, dbType string) error {
	var tbs []*core.Table
	for _, t := range tables {
		t := x.TableInfo(t)
		t.Table.Name = t.Name
		tbs = append(tbs, t.Table)
	}
	if len(dbType) > 0 {
		return x.DumpTablesToFile(tbs, filePath, core.DbType(dbType))
	}
	return x.DumpTablesToFile(tbs, filePath)
}

// MaxBatchInsertSize returns the table's max batch insert size
func MaxBatchInsertSize(bean interface{}) int {
	t := x.TableInfo(bean)
	return 999 / len(t.ColumnsSeq())
}

// Count returns records number according struct's fields as database query conditions
func Count(bean interface{}) (int64, error) {
	return x.Count(bean)
}

// IsTableNotEmpty returns true if table has at least one record
func IsTableNotEmpty(tableName string) (bool, error) {
	return x.Table(tableName).Exist()
}

// DeleteAllRecords will delete all the records of this table
func DeleteAllRecords(tableName string) error {
	_, err := x.Exec(fmt.Sprintf("DELETE FROM %s", tableName))
	return err
}

// GetMaxID will return max id of the table
func GetMaxID(beanOrTableName interface{}) (maxID int64, err error) {
	_, err = x.Select("MAX(id)").Table(beanOrTableName).Get(&maxID)
	return
}

// FindByMaxID filled results as the condition from database
func FindByMaxID(maxID int64, limit int, results interface{}) error {
	return x.Where("id <= ?", maxID).
		OrderBy("id DESC").
		Limit(limit).
		Find(results)
}
