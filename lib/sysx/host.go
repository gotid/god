package sysx

import (
	"os"

	"github.com/gotid/god/lib/stringx"
)

var hostname string

func init() {
	var err error
	hostname, err = os.Hostname()
	if err != nil {
		hostname = stringx.RandId()
	}
}

// Hostname 返回主机的名称，如无则返回一个随机id。
func Hostname() string {
	return hostname
}
