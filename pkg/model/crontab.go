package model

import (
	"time"
)

type Crontab struct {
	ID        int64
	Namespace string    `xorm:"varchar(32)"`
	Service   string    `xorm:"varchar(32)"`
	Command   string    `xorm:"varchar(800) notnull"`
	Schedule  string    `xorm:"varchar(20) notnull"`
	CreateAt  time.Time `xorm:"timestamp notnull created"`
	UpdateAt  time.Time `xorm:"timestamp notnull updated"`
}

func CreateCrontab(namespace, service, command, schedule string) (int64, error) {
	crontab := new(Crontab)
	crontab.Namespace = namespace
	crontab.Service = service
	crontab.Command = command
	crontab.Schedule = schedule
	if _, err := MEngine().Insert(crontab); err != nil {
		return 0, err
	}
	return crontab.ID, nil
}
