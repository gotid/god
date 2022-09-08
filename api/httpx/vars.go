package httpx

import "github.com/gotid/god/api/internal/header"

const (
	ContentEncoding = "Content-Encoding"
	ContentSecurity = "X-Content-Security"
	ContentType     = header.ContentType
	JsonContentType = header.JsonContentType
	KeyField        = "key"
	SecretField     = "secret"
	TypeField       = "type"
	CryptionType    = 1
)

const (
	CodeSignaturePass          = iota // 签名通过
	CodeSignatureInvalidHeader        // 无效的签名头
	CodeSignatureWrongTime            // 错误的签名时间
	CodeSignatureInvalidToken         // 无效的签名令牌
)
