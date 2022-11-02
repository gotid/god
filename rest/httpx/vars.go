package httpx

import "github.com/gotid/god/rest/internal/header"

const (
	// ContentEncoding 意为 Content-Encoding。
	ContentEncoding = "Content-Encoding"
	// ContentSecurity 意为 X-Content-Security。
	ContentSecurity = "X-Content-Security"
	// ContentType 意为 Content-Type。
	ContentType = header.ContentType
	// JsonContentType 意为 application/json。
	JsonContentType = header.JsonContentType
	// KeyField 意为键字段。
	KeyField = "key"
	// SecretField 意为内容密钥字段。
	SecretField = "secret"
	// TypeField 意为内容类型字段，一般为整数。
	TypeField = "type"
	// EncryptionType 意为加密类型。
	EncryptionType = 1
)

const (
	// CodeSignaturePass 意为通过签名验证。
	CodeSignaturePass = iota
	// CodeSignatureInvalidHeader 意为签名标头无效。
	CodeSignatureInvalidHeader
	// CodeSignatureWrongTime 意为签名时间错误。
	CodeSignatureWrongTime
	// CodeSignatureInvalidToken 意为签名令牌无效。
	CodeSignatureInvalidToken
)
