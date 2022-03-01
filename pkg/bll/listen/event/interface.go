// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package event

const (
	Create = "create"
	Update = "update"
	Delete = "delete"
)

type capturer interface {
	valid() bool
	ready() bool
	parse() bool
	operate() bool
}

func handleEvent(c capturer) {
	if !c.valid() {
		return
	}
	if !c.ready() {
		return
	}
	if !c.parse() {
		return
	}
	if !c.operate() {
		return
	}
}
