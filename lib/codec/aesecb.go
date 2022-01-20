package codec

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"

	"github.com/gotid/god/lib/logx"
)

var ErrPaddingSize = errors.New("填充尺寸错误")

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

// NewECBEncryptor 返回一个 ECB 模式的 AES 加密器。
func NewECBEncryptor(b cipher.Block) cipher.BlockMode {
	return (*ecbEncryptor)(newECB(b))
}

func (x *ecbEncryptor) BlockSize() int { return x.blockSize }

// CryptBlocks 未返回错误是因为 cipher.BlockMode 不允许这样。
func (x *ecbEncryptor) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		logx.Error("crypto/cipher: 输入块未满")
		return
	}
	if len(dst) < len(src) {
		logx.Error("crypto/cipher: 输出不能小于输出")
		return
	}

	for len(src) > 0 {
		x.b.Encrypt(dst, src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}

type ecbDecryptor ecb

// NewECBDecryptor 返回一个 ECB 解密器。
func NewECBDecryptor(b cipher.Block) cipher.BlockMode {
	return (*ecbDecryptor)(newECB(b))
}

func (x *ecbDecryptor) BlockSize() int {
	return x.blockSize
}

// CryptBlocks 未返回错误是因为 cipher.BlockMode 不允许这样。
func (x *ecbDecryptor) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		logx.Error("crypto/cipher: 输入块未满")
		return
	}
	if len(dst) < len(src) {
		logx.Error("crypto/cipher: 输出不能小于输出")
		return
	}

	for len(src) > 0 {
		x.b.Decrypt(dst, src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}

// EcbDecrypt 使用指定的键加密原文本。
func EcbDecrypt(key, src []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		logx.Errorf("解密秘钥错误：%s", key)
		return nil, err
	}

	decrypter := NewECBDecryptor(block)
	decrypted := make([]byte, len(src))
	decrypter.CryptBlocks(decrypted, src)

	return pkcs5UnPadding(decrypted, decrypter.BlockSize())
}

// EcbDecryptBase64 使用指定的 base64 编码的键，解密 base64 编码的原文本。
// 返回 base64 编码后的字符串。
func EcbDecryptBase64(key, src string) (string, error) {
	keyBytes, err := getKeyBytes(key)
	if err != nil {
		return "", err
	}

	encryptedBytes, err := base64.StdEncoding.DecodeString(src)
	if err != nil {
		return "", err
	}

	decryptedBytes, err := EcbDecrypt(keyBytes, encryptedBytes)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(decryptedBytes), nil
}

// EcbEncrypt 使用指定的键加密原文本。
func EcbEncrypt(key, src []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		logx.Errorf("加密秘钥错误：%s", key)
		return nil, err
	}

	padded := pkcs5Padding(src, block.BlockSize())
	encrypted := make([]byte, len(padded))
	encryptor := NewECBEncryptor(block)
	encryptor.CryptBlocks(encrypted, padded)

	return encrypted, nil
}

// EcbEncryptBase64 使用指定的 base64 编码的键，加密 base64 编码的原文本。
//// 返回 base64 编码后的字符串。
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

func getKeyBytes(key string) ([]byte, error) {
	if len(key) > 32 {
		if keyBytes, err := base64.StdEncoding.DecodeString(key); err != nil {
			return nil, err
		} else {
			return keyBytes, nil
		}
	}

	return []byte(key), nil
}

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padText...)
}

func pkcs5UnPadding(src []byte, blockSize int) ([]byte, error) {
	length := len(src)
	unPadding := int(src[length-1])
	if unPadding >= length || unPadding > blockSize {
		return nil, ErrPaddingSize
	}

	return src[:length-unPadding], nil
}
