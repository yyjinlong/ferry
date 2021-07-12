// copyright @ 2020 ops inc.
//
// author: jinlong yang
//

package g

func In(data string, dataList []string) bool {
	for _, item := range dataList {
		if data == item {
			return true
		}
	}
	return false
}
