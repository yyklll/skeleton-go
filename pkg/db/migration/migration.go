package migration

import (
	"fmt"
	"strings"

	"github.com/yyklll/skeleton/pkg/log"
	"xorm.io/xorm"
)

const minDBVersion = 0

// Migration describes on migration from lower version to high version
type Migration interface {
	Description() string
	Migrate(*xorm.Engine) error
}

type migration struct {
	description string
	migrate     func(*xorm.Engine) error
}

// NewMigration creates a new migration
func NewMigration(desc string, fn func(*xorm.Engine) error) Migration {
	return &migration{desc, fn}
}

// Description returns the migration's description
func (m *migration) Description() string {
	return m.description
}

// Migrate executes the migration
func (m *migration) Migrate(x *xorm.Engine) error {
	return m.migrate(x)
}

// Version describes the version table. Should have only one row with id==1
type Version struct {
	ID      int64 `xorm:"pk autoincr"`
	Version int64
}

// This is a sequence of migrations. Add new migrations to the bottom of the list.
// If you want to "retire" a migration, remove it from the top of the list and
// update minDBVersion accordingly
var migrations = []Migration{
	NewMigration("init database", initDatabase),
}

// Migrate database to current version
func Migrate(x *xorm.Engine) error {
	if err := x.Sync(new(Version)); err != nil {
		return fmt.Errorf("sync: %v", err)
	}

	currentVersion := &Version{ID: 1}
	has, err := x.Get(currentVersion)
	if err != nil {
		return fmt.Errorf("get: %v", err)
	} else if !has {
		// If the version record does not exist we think
		// it is a fresh installation and we can skip all migrations.
		currentVersion.ID = 0
		currentVersion.Version = int64(minDBVersion + len(migrations))

		if _, err = x.InsertOne(currentVersion); err != nil {
			return fmt.Errorf("insert: %v", err)
		}
	}

	v := currentVersion.Version
	if minDBVersion > v {
		log.Fatalln("Auto-migration from your previously installed version no longer supported. Please try upgrading to a lower version first , then upgrade to this version.")
		return nil
	}

	if int(v-minDBVersion) > len(migrations) {
		// User downgraded Gitea.
		currentVersion.Version = int64(len(migrations) + minDBVersion)
		_, err = x.ID(1).Update(currentVersion)
		return err
	}
	for i, m := range migrations[v-minDBVersion:] {
		log.Infof("Migration[%d]: %s", v+int64(i), m.Description())
		if err = m.Migrate(x); err != nil {
			return fmt.Errorf("do migrate: %v", err)
		}
		currentVersion.Version = v + int64(i) + 1
		if _, err = x.ID(1).Update(currentVersion); err != nil {
			return err
		}
	}
	return nil
}

func dropTableColumns(sess *xorm.Session, tableName string, columnNames ...string) (err error) {
	if tableName == "" || len(columnNames) == 0 {
		return nil
	}
	// TODO: This will not work if there are foreign keys

	// Drop indexes on columns first
	// MySQL is the only driver supported
	sql := fmt.Sprintf("SHOW INDEX FROM %s WHERE column_name IN ('%s')", tableName, strings.Join(columnNames, "','"))
	res, err := sess.Query(sql)
	if err != nil {
		return err
	}
	for _, index := range res {
		indexName := index["column_name"]
		if len(indexName) > 0 {
			_, err := sess.Exec(fmt.Sprintf("DROP INDEX `%s` ON `%s`", indexName, tableName))
			if err != nil {
				return err
			}
		}
	}

	// Now drop the columns
	cols := ""
	for _, col := range columnNames {
		if cols != "" {
			cols += ", "
		}
		cols += "DROP COLUMN `" + col + "`"
	}
	if _, err := sess.Exec(fmt.Sprintf("ALTER TABLE `%s` %s", tableName, cols)); err != nil {
		return fmt.Errorf("Drop table `%s` columns %v: %v", tableName, columnNames, err)
	}

	return nil
}
