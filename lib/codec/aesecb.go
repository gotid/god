package codec

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"

	"github.com/gotid/god/lib/logx"
)

// ErrPaddingSize 表示错误的填充大小。
var ErrPaddingSize = errors.New("填充大小错误")

type ecb struct {
	b         cipher.Block
	blockSize int
}

func newECB(b cipher.Block) *ecb {
	return &ecb{
		b:         b,
		blockSize: b.BlockSize(),
	}
}

type ecbEncryptor ecb

func (x *ecbEncryptor) BlockSize() int {
	return x.blockSize
}

func (x *ecbEncryptor) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		logx.Error("crypto/cipher: 输入块不完整")
		return
	}
	if len(dst) < len(src) {
		logx.Error("crypto/cipher: 输出长度不能小于输入")
		return
	}

	for len(src) > 0 {
		x.b.Encrypt(dst, src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}

// NewECBEncryptor 返回一个 ECB 加密器。
func NewECBEncryptor(b cipher.Block) cipher.BlockMode {
	return (*ecbEncryptor)(newECB(b))
}

type ecbDecryptor ecb

func (x *ecbDecryptor) BlockSize() int {
	return x.blockSize
}

func (x *ecbDecryptor) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		logx.Error("crypto/cipher: 输入块不完整")
		return
	}
	if len(dst) < len(src) {
		logx.Error("crypto/cipher: 输出长度不能小于输入")
		return
	}

	for len(src) > 0 {
		x.b.Decrypt(dst, src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}

// NewECBDecryptor 返回一个 ECB 解密器。
func NewECBDecryptor(b cipher.Block) cipher.BlockMode {
	return (*ecbDecryptor)(newECB(b))
}

// EcbEncrypt 使用给定的密钥加密 src。
func EcbEncrypt(key, src []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		logx.Errorf("加密密钥错误：% x", key)
		return nil, err
	}

	padded := pkcs5Padding(src, block.BlockSize())
	encrypted := make([]byte, len(padded))
	encryptor := NewECBEncryptor(block)
	encryptor.CryptBlocks(encrypted, padded)

	return encrypted, nil
}

// EcbEncryptBase64 使用给定的 base64 编码后的密钥加密 base64 编码后的 src。
// 返回的字符串也是 base64 编码的。
func EcbEncryptBase64(key, src string) (string, error) {
	keyBytes, err := getKeyBytes(key)
	if err != nil {
		return "", err
	}

	srcBytes, err := base64.StdEncoding.DecodeString(src)
	if err != nil {
		return "", err
	}

	encryptedBytes, err := EcbEncrypt(keyBytes, srcBytes)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encryptedBytes), nil
}

// EcbDecrypt 使用给定的密钥解密 src。
func EcbDecrypt(key, src []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		logx.Errorf("解密密钥错误：% x", key)
		return nil, err
	}

	decryptor := NewECBDecryptor(block)
	decrypted := make([]byte, len(src))
	decryptor.CryptBlocks(decrypted, src)

	return pkcs5UnPadding(decrypted, decryptor.BlockSize())
}

// EcbDecryptBase64 使用给定的 base64 编码后的密钥解密 base64 编码后的 src。
// 返回的字符串也是 base64 编码的。
func EcbDecryptBase64(key, src string) (string, error) {
	keyBytes, err := getKeyBytes(key)
	if err != nil {
		return "", err
	}

	srcBytes, err := base64.StdEncoding.DecodeString(src)
	if err != nil {
		return "", err
	}

	decryptedBytes, err := EcbDecrypt(keyBytes, srcBytes)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(decryptedBytes), nil
}

func getKeyBytes(key string) ([]byte, error) {
	if len(key) <= 32 {
		return []byte(key), nil
	}

	keyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, err
	}

	return keyBytes, nil
}

func pkcs5Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padText...)
}

func pkcs5UnPadding(src []byte, blockSize int) ([]byte, error) {
	length := len(src)
	unPadding := int(src[length-1])
	if unPadding >= length || unPadding > blockSize {
		return nil, ErrPaddingSize
	}

	return src[:length-unPadding], nil
}
