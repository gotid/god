package sysx

import (
	"github.com/gotid/god/lib/stringx"
	"os"
)

var hostname string

func init() {
	var err error
	hostname, err = os.Hostname()
	if err != nil {
		hostname = stringx.RandId()
	}
}

// Hostname 返回主机名，如无则返回随机字符串。
func Hostname() string {
	return hostname
}
