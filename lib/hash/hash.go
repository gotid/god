package hash

import (
	"crypto/md5"
	"fmt"

	"github.com/spaolacci/murmur3"
)

// Hash 返回 data 的 hash 值。
func Hash(data []byte) uint64 {
	return murmur3.Sum64(data)
}

// Md5 返回 data 的 hash 值。
func Md5(data []byte) []byte {
	digest := md5.New()
	digest.Write(data)
	return digest.Sum(nil)
}

// Md5Hex 返回 data 的 md5 hex 字符串。
func Md5Hex(data []byte) string {
	return fmt.Sprintf("%x", Md5(data))
}
