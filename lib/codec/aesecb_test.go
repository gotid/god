package codec

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAesEcb(t *testing.T) {
	var (
		// 16的倍数
		key = []byte("aaaaaaaaaaaaaaaa")
		// 明文
		val = []byte("hello")
		// 不足16位
		badKey1 = []byte("aaaaaa")
		// 超过32位字符
		badKey2 = []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaabbbbb")
	)

	_, err := EcbEncrypt(badKey1, val)
	assert.NotNil(t, err)

	_, err = EcbEncrypt(badKey2, val)
	assert.NotNil(t, err)

	dst, err := EcbEncrypt(key, val)
	assert.Nil(t, err)

	_, err = EcbDecrypt(badKey1, dst)
	assert.NotNil(t, err)

	_, err = EcbDecrypt(badKey2, dst)
	assert.NotNil(t, err)

	_, err = EcbDecrypt(key, val)
	assert.Nil(t, err)

	src, err := EcbDecrypt(key, dst)
	assert.Nil(t, err)
	assert.Equal(t, val, src)
}

func TestAesEcbBase64(t *testing.T) {
	const (
		val     = "hello"
		badKey1 = "aaaaaaaaa"
		// more than 32 chars
		badKey2 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	)
	key := []byte("q4t7w!z%C*F-JaNdRgUjXn2r5u8x/A?D")
	b64Key := base64.StdEncoding.EncodeToString(key)
	b64Val := base64.StdEncoding.EncodeToString([]byte(val))
	_, err := EcbEncryptBase64(badKey1, val)
	assert.NotNil(t, err)
	_, err = EcbEncryptBase64(badKey2, val)
	assert.NotNil(t, err)
	_, err = EcbEncryptBase64(b64Key, val)
	assert.NotNil(t, err)
	dst, err := EcbEncryptBase64(b64Key, b64Val)
	assert.Nil(t, err)
	_, err = EcbDecryptBase64(badKey1, dst)
	assert.NotNil(t, err)
	_, err = EcbDecryptBase64(badKey2, dst)
	assert.NotNil(t, err)
	_, err = EcbDecryptBase64(b64Key, val)
	assert.NotNil(t, err)
	src, err := EcbDecryptBase64(b64Key, dst)
	assert.Nil(t, err)
	b, err := base64.StdEncoding.DecodeString(src)
	assert.Nil(t, err)
	assert.Equal(t, val, string(b))
}
