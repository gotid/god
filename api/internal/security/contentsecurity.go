package security

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gotid/god/api/httpx"
	"github.com/gotid/god/lib/codec"
	"github.com/gotid/god/lib/iox"
	"github.com/gotid/god/lib/logx"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	signatureField   = "signature"
	timeField        = "time"
	requestUriHeader = "X-Request-Uri"
)

var (
	// ErrInvalidHeader 表示 X-Content-Security 无效的错误。
	ErrInvalidHeader = errors.New("无效的 X-Content-Security 请求头")
	// ErrInvalidPublicKey 表示公钥无效的错误。
	ErrInvalidPublicKey = errors.New("无效的公钥")
	// ErrInvalidSecret 表示秘钥无效。
	ErrInvalidSecret = errors.New("无效的秘钥")
	// ErrInvalidKey 表示解析后的键无效。
	ErrInvalidKey = errors.New("无效的键")
	// ErrInvalidContentType 表示内容类型无效。
	ErrInvalidContentType = errors.New("无效的内容类型")
)

// ContentSecurityHeader 是一个内容安全的标头。
type ContentSecurityHeader struct {
	Key         []byte
	Timestamp   string
	ContentType int
	Signature   string
}

// Encrypted 判断是否为加密请求。
func (h *ContentSecurityHeader) Encrypted() bool {
	return h.ContentType == httpx.EncryptionType
}

// ParseContentSecurity 解析给定 http 请求的内容安全设置，并返回 ContentSecurityHeader。
func ParseContentSecurity(decryptors map[string]codec.RsaDecryptor, r *http.Request) (*ContentSecurityHeader, error) {
	contentSecurity := r.Header.Get(httpx.ContentSecurity)
	attrs := httpx.ParseHeader(contentSecurity)
	fingerprint := attrs[httpx.KeyField]
	secret := attrs[httpx.SecretField]
	signature := attrs[signatureField]

	if len(fingerprint) == 0 || len(secret) == 0 || len(signature) == 0 {
		return nil, ErrInvalidHeader
	}

	decryptor, ok := decryptors[fingerprint]
	if !ok {
		return nil, ErrInvalidPublicKey
	}

	decryptedSecret, err := decryptor.DecryptBase64(secret)
	if err != nil {
		return nil, ErrInvalidSecret
	}

	attrs = httpx.ParseHeader(string(decryptedSecret))
	base64Key := attrs[httpx.KeyField]
	timestamp := attrs[timeField]
	contentType := attrs[httpx.TypeField]

	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, ErrInvalidKey
	}

	cType, err := strconv.Atoi(contentType)
	if err != nil {
		return nil, ErrInvalidContentType
	}

	return &ContentSecurityHeader{
		Key:         key,
		Timestamp:   timestamp,
		ContentType: cType,
		Signature:   signature,
	}, nil
}

// VerifySignature 检验给定请求的签名是否正确。
func VerifySignature(r *http.Request, securityHeader *ContentSecurityHeader, tolerance time.Duration) int {
	seconds, err := strconv.ParseInt(securityHeader.Timestamp, 10, 64)
	if err != nil {
		return httpx.CodeSignatureInvalidHeader
	}

	now := time.Now().Unix()
	toleranceSeconds := int64(tolerance.Seconds())
	if seconds+toleranceSeconds < now || now+toleranceSeconds < seconds {
		return httpx.CodeSignatureWrongTime
	}

	reqPath, reqQuery := getPathQuery(r)
	signContent := strings.Join([]string{
		securityHeader.Timestamp,
		r.Method,
		reqPath,
		reqQuery,
		computeBodySignature(r),
	}, "\n")
	actualSignature := codec.HmacBase64(securityHeader.Key, signContent)

	if securityHeader.Signature == actualSignature {
		return httpx.CodeSignaturePass
	}

	logx.Infof("签名不同，期望：%s，实际：%s", securityHeader.Signature, actualSignature)

	return httpx.CodeSignatureInvalidToken
}

func computeBodySignature(r *http.Request) string {
	var dup io.ReadCloser
	r.Body, dup = iox.DupReadCloser(r.Body)
	sha := sha256.New()
	io.Copy(sha, r.Body)
	r.Body = dup
	return fmt.Sprintf("%x", sha.Sum(nil))
}

func getPathQuery(r *http.Request) (string, string) {
	requestUri := r.Header.Get(requestUriHeader)
	if len(requestUri) == 0 {
		return r.URL.Path, r.URL.RawQuery
	}

	uri, err := url.Parse(requestUri)
	if err != nil {
		return r.URL.Path, r.URL.RawQuery
	}

	return uri.Path, uri.RawQuery
}
