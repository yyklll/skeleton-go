package migration

import (
	"fmt"
	"time"

	"xorm.io/xorm"
)

func initDatabase(x *xorm.Engine) error {
	type TestTable struct {
		ID        int64 `xorm:"pk autoincr"`
		Name      string
		CreatedAt time.Time `xorm:"INDEX created"`
		UpdatedAt time.Time `xorm:"INDEX updated"`
	}

	if err := x.Sync2(new(TestTable)); err != nil {
		return fmt.Errorf("Sync2: %v", err)
	}
	return nil
}
