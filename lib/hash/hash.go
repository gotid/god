package hash

import (
	"crypto/md5"
	"fmt"

	"github.com/spaolacci/murmur3"
)

// Hash 返回指定数据的哈希值。
//
// 字节大于10，性能远高于MD5，且碰撞率更低
func Hash(data []byte) uint64 {
	return murmur3.Sum64(data)
}

// Md5 返回指定数据的 md5 字节切片。
func Md5(data []byte) []byte {
	digest := md5.New()
	digest.Write(data)
	return digest.Sum(nil)
}

// Md5Hex 返回指定数据的 md5 十六进制字符串。
func Md5Hex(data []byte) string {
	return fmt.Sprintf("%x", Md5(data))
}
