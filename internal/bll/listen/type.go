// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package listen

const (
	Create = "create"
	Update = "update"
	Delete = "delete"
)

var (
	depResourceVersionMap = make(map[string]string)
	endResourceVersionMap = make(map[string]string)
)
