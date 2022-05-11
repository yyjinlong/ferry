package model

import (
	"time"
)

type Namespace struct {
	ID       int64
	Name     string    `xorm:"varchar(50) notnull"`
	token    string    `xorm:"text notnull"`
	Creator  string    `xorm:"varchar(50) notnull"`
	CreateAt time.Time `xorm:"timestamp notnull created"`
}
