// Copyright @ 2021 OPS Inc.
//
// Author: Jinlong Yang
//

package model

import (
	"fmt"

	"github.com/yyjinlong/golib/db"
	"xorm.io/xorm"
)

var (
	NotFound = fmt.Errorf("query data not found")
)

func MEngine() *xorm.Engine {
	return db.MEngine
}

func SEngine() *xorm.Engine {
	return db.SEngine
}
