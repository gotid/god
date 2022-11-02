package codec

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"os"
)

var (
	// ErrPrivateKey 表示一个私钥无效的错误。
	ErrPrivateKey = errors.New("私钥错误")
	// ErrPublicKey 表示一个公钥无效的错误。
	ErrPublicKey = errors.New("公钥错误")
	// ErrNotRsaKey 表示无效的 RSA 密钥。
	ErrNotRsaKey = errors.New("密钥类型不是 RSA")
)

type (
	// RsaDecryptor 接口封装了一个基于 RSA 的解密器。
	RsaDecryptor interface {
		Decrypt(input []byte) ([]byte, error)
		DecryptBase64(input string) ([]byte, error)
	}

	// RsaEncryptor 接口封装了一个基于 RSA 的加密器。
	RsaEncryptor interface {
		Encrypt(input []byte) ([]byte, error)
	}

	rsaBase struct {
		bytesLimit int
	}

	rsaDecryptor struct {
		rsaBase
		privateKey *rsa.PrivateKey
	}

	rsaEncryptor struct {
		rsaBase
		publicKey *rsa.PublicKey
	}
)

// NewRsaDecryptor 返回一个给定私钥文件的 RsaDecryptor。
func NewRsaDecryptor(file string) (RsaDecryptor, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(content)
	if block == nil {
		return nil, ErrPrivateKey
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return &rsaDecryptor{
		rsaBase: rsaBase{
			bytesLimit: privateKey.N.BitLen() >> 3,
		},
		privateKey: privateKey,
	}, nil
}

func (d *rsaDecryptor) Decrypt(input []byte) ([]byte, error) {
	return d.crypt(input, func(block []byte) ([]byte, error) {
		return rsaDecryptBlock(d.privateKey, block)
	})
}

func (d *rsaDecryptor) DecryptBase64(input string) ([]byte, error) {
	if len(input) == 0 {
		return nil, nil
	}

	base64Decoded, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return nil, err
	}

	return d.Decrypt(base64Decoded)
}

// NewRsaEncryptor 返回一个给定公钥的 RsaEncryptor。
func NewRsaEncryptor(publicKey []byte) (RsaEncryptor, error) {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, ErrPublicKey
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch pubKey := pub.(type) {
	case *rsa.PublicKey:
		return &rsaEncryptor{
			rsaBase: rsaBase{
				// https://www.ietf.org/rfc/rfc2313.txt
				// 数据D的长度不得超过k-11个八位字节，这是正的，因为模的长度k是至少12个八位字节。
				bytesLimit: (pubKey.N.BitLen() >> 3) - 11,
			},
			publicKey: pubKey,
		}, nil
	default:
		return nil, ErrNotRsaKey
	}
}

func (r *rsaEncryptor) Encrypt(block []byte) ([]byte, error) {
	return r.crypt(block, func(bytes []byte) ([]byte, error) {
		return rsaEncryptBlock(r.publicKey, block)
	})
}

func (r *rsaBase) crypt(input []byte, cryptFn func([]byte) ([]byte, error)) ([]byte, error) {
	var result []byte
	inputLen := len(input)

	for i := 0; i*r.bytesLimit < inputLen; i++ {
		start := r.bytesLimit * i
		var stop int
		if r.bytesLimit*(i+1) > inputLen {
			stop = inputLen
		} else {
			stop = r.bytesLimit * (i + 1)
		}
		bs, err := cryptFn(input[start:stop])
		if err != nil {
			return nil, err
		}

		result = append(result, bs...)
	}

	return result, nil
}

func rsaDecryptBlock(privateKey *rsa.PrivateKey, block []byte) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, privateKey, block)
}

func rsaEncryptBlock(publicKey *rsa.PublicKey, block []byte) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, publicKey, block)
}
