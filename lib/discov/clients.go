package discov

import (
	"fmt"
	"github.com/gotid/god/lib/discov/internal"
	"strings"
)

const (
	_ = iota
	indexOfId
)

const timeToLive int64 = 10

// TimeToLive 是在 etcd 中存活的时间。
var TimeToLive = timeToLive

func extract(etcdKey string, index int) (string, bool) {
	if index < 0 {
		return "", false
	}

	fields := strings.FieldsFunc(etcdKey, func(r rune) bool {
		return r == internal.Delimiter
	})
	if index >= len(fields) {
		return "", false
	}

	return fields[index], true
}

func extractId(etcdKey string) (string, bool) {
	return extract(etcdKey, indexOfId)
}

func makeEtcdKey(key string, id int64) string {
	return fmt.Sprintf("%s%c%d", key, internal.Delimiter, id)
}
