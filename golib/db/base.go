// Copyright @ 2021 OPS Inc.
//
// Author: Jinlong Yang
//

package db

import (
	"sync"

	_ "github.com/lib/pq"
	"github.com/yyjinlong/golib/log"
	"xorm.io/xorm"
	"xorm.io/xorm/names"
)

var (
	once    sync.Once
	MEngine *xorm.Engine
	SEngine *xorm.Engine
)

// Connect 连接到数据库 TODO: 实时感知主从变化
func Connect(driver, master string, slaves ...string) {
	m, err := xorm.NewEngine(driver, master)
	if err != nil {
		log.Panicf("Connect master: %s database error: %s", master, err)
	}

	sList := make([]*xorm.Engine, 0)
	for _, slave := range slaves {
		s, err := xorm.NewEngine(driver, slave)
		if err != nil {
			log.Panicf("Connect slave: %s database error: %s", slave, err)
		}
		sList = append(sList, s)
	}

	eg, err := xorm.NewEngineGroup(m, sList)
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
