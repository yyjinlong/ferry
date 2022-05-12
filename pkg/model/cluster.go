package model

import (
	"time"
)

type Cluster struct {
	ID       int64
	Name     string    `xorm:"varchar(50) notnull"`
	Token    string    `xorm:"text notnull"`
	Creator  string    `xorm:"varchar(50) notnull"`
	CreateAt time.Time `xorm:"timestamp notnull created"`
}

func GetCluster(cluster string) (*Cluster, error) {
	cr := new(Cluster)
	if has, err := SEngine().Where("name = ?", cluster).Get(cr); err != nil {
		return nil, err
	} else if !has {
		return nil, NotFound
	}
	return cr, nil
}
