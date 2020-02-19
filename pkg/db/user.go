package db

import (
	"time"
)

// Action represents user operation type and other information to
// repository. It implemented interface base.Actioner so that can be
// used in template render.
type User struct {
	ID        int64 `xorm:"pk autoincr"`
	Name      string
	CreatedAt time.Time `xorm:"INDEX created"`
	UpdatedAt time.Time `xorm:"INDEX updated"`
}

func (t *User) GetName() string {
	return t.Name
}
