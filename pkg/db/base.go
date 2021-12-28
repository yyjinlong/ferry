// Copyright @ 2021 OPS Inc.
//
// Author: Jinlong Yang
//

package db

import (
	"sync"

	_ "github.com/lib/pq"
	"xorm.io/xorm"
	"xorm.io/xorm/names"

	"ferry/pkg/g"
	"ferry/pkg/log"
)

var (
	once    sync.Once
	MEngine *xorm.Engine
	SEngine *xorm.Engine
)

func Connect() {
	master, err := xorm.NewEngine("postgres", g.Config().Postgres.Master)
	if err != nil {
		log.Panicf("Connect master database error: %s", err)
	}

	slave1, err := xorm.NewEngine("postgres", g.Config().Postgres.Slave1)
	if err != nil {
		log.Panicf("Connect slave1 database error: %s", err)
	}

	slave2, err := xorm.NewEngine("postgres", g.Config().Postgres.Slave2)
	if err != nil {
		log.Panicf("Connect slave2 database error: %s", err)
	}

	slaves := []*xorm.Engine{slave1, slave2}
	eg, err := xorm.NewEngineGroup(master, slaves)
	if err != nil {
		log.Panicf("Create engine group failed: %s", err)
	}

	if err := eg.Ping(); err != nil {
		log.Panicf("Ping connect error: %s", err)
	}

	eg.ShowSQL(false)
	eg.SetMapper(names.GonicMapper{})

	// 单列模式
	once.Do(func() {
		MEngine = eg.Master()
		SEngine = eg.Slave()
	})
}
