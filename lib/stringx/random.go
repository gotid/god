package stringx

import (
	crand "crypto/rand"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	defaultRandLen = 8
	idLen          = 8
	letterIdxBits  = 6                    // 6比特表示一个字母索引
	letterIdxMax   = 63 / letterIdxBits   // of letter indices fitting in 63 bits
	letterIdxMask  = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterBytes    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var src = newLockedSource(time.Now().UnixNano())

// Rand 返回一个 8 位数的随机字符串。
func Rand() string {
	return Randn(defaultRandLen)
}

// RandId 返回一个随机 id 字符串。
func RandId() string {
	b := make([]byte, idLen)
	_, err := crand.Read(b)
	if err != nil {
		return Randn(idLen)
	}

	return fmt.Sprintf("%x%x%x%x", b[0:2], b[2:4], b[4:6], b[6:8])
}

// Randn 返回一个长度为 n 的随机字符串。
func Randn(n int) string {
	b := make([]byte, n)

	// src.Int63() 生成 63 个随机位，足以容纳 letterIdxMax 个字符！
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

// Seed 设置随机种子数。
func Seed(seed int64) {
	src.Seed(seed)
}

type lockedSource struct {
	source rand.Source
	lock   sync.Mutex
}

func newLockedSource(seed int64) *lockedSource {
	return &lockedSource{
		source: rand.NewSource(seed),
	}
}

func (ls *lockedSource) Int63() int64 {
	ls.lock.Lock()
	defer ls.lock.Unlock()
	return ls.source.Int63()
}

func (ls *lockedSource) Seed(seed int64) {
	ls.lock.Lock()
	defer ls.lock.Unlock()
	ls.source.Seed(seed)
}
