package codec

import (
	"encoding/base64"
	"testing"

	"github.com/gotid/god/lib/logx"

	"github.com/stretchr/testify/assert"
)

func TestAesEcb(t *testing.T) {
	var (
		key     = []byte("!@#$%^&SDF$%^&*(")
		val     = []byte("hello")
		badKey1 = []byte("asdfasdf")
		// 超过 32 个字符
		badKey2 = []byte("hello hello hello hello hello hello hello")
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
		badKey1 = "asdfasdf"
		// 超过 32 个字符
		badKey2 = "hello hello hello hello hello hello hello"
	)

	key := []byte("!@#$%^&SDF$%^&*(")
	b64Key := base64.StdEncoding.EncodeToString(key)
	b64Val := base64.StdEncoding.EncodeToString([]byte(val))

	_, err := EcbEncryptBase64(badKey1, val)
	assert.NotNil(t, err)

	_, err = EcbEncryptBase64(badKey2, val)
	assert.NotNil(t, err)

	_, err = EcbEncryptBase64(b64Key, val)
	assert.NotNil(t, err)

	dst, err := EcbEncryptBase64(b64Key, b64Val)
	logx.Info("dst: ", dst)
	assert.Nil(t, err)
	_, err = EcbDecryptBase64(badKey1, dst)
	assert.NotNil(t, err)
	_, err = EcbDecryptBase64(badKey2, dst)
	assert.NotNil(t, err)
	_, err = EcbDecryptBase64(b64Key, val)
	assert.NotNil(t, err)

	src, err := EcbDecryptBase64(b64Key, dst)
	logx.Info("src: ", src)
	assert.Nil(t, err)
	b, err := base64.StdEncoding.DecodeString(src)
	assert.Nil(t, err)
	assert.Equal(t, val, string(b))
	logx.Info("val: ", val)
}
