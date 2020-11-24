package httpx

const (
	ApplicationJson   = "application/json"
	ContentEncoding   = "Content-Encoding"
	ContentSecurity   = "X-Content-Security"
	ContentType       = "Content-Type"
	MultipartFormData = "multipart/form-data"
	KeyField          = "key"
	SecretField       = "secret"
	TypeField         = "type"
	EncryptedType     = 1
)

const (
	CodeSignaturePass          = iota // 签名通过
	CodeSignatureInvalidHeader        // 无效的签名头
	CodeSignatureWrongTime            // 错误的签名时间
	CodeSignatureInvalidToken         // 无效的签名令牌
)
