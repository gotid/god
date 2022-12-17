package utils

import "github.com/google/uuid"

// NewUUID 返回一个 uuid 字符串。
func NewUUID() string {
	return uuid.New().String()
}
