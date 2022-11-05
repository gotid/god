package httpx

import "github.com/gotid/god/api/internal/header"

const (
	// ContentEncoding 意为 Content-Encoding。
	ContentEncoding = "Content-Encoding"
	// ContentSecurity 意为 X-Content-Security。
	ContentSecurity = "X-Content-Security"
	// ContentType 意为 Content-Type。
	ContentType = header.ContentType
	// JsonContentType 意为 application/json。
	JsonContentType = header.JsonContentType
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
