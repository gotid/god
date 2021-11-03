package security

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"git.zc0901.com/go/god/api/httpx"
	"git.zc0901.com/go/god/lib/codec"
	"git.zc0901.com/go/god/lib/iox"
	"git.zc0901.com/go/god/lib/logx"
)

const (
	requestUriHeader = "X-Request-Uri"
	signatureField   = "signature"
	timeField        = "time"
)

var (
	ErrInvalidHeader      = errors.New("无效的 X-Content-Security 头")
	ErrInvalidPublicKey   = errors.New("无效的公钥")
	ErrInvalidSecret      = errors.New("无效的秘钥")
	ErrInvalidKey         = errors.New("无效的键")
	ErrInvalidContentType = errors.New("无效的 Content-Type")
)

// ContentSecurityHeader 是一个内容安全标头。
type ContentSecurityHeader struct {
	Key         []byte
	Timestamp   string
	ContentType int
	Signature   string
}

// Encrypted 判断是否为加密请求。
func (h *ContentSecurityHeader) Encrypted() bool {
	return h.ContentType == httpx.CryptionType
}

// ParseContentSecurity 解析指定请求的内容安全设置。
func ParseContentSecurity(decrypters map[string]codec.RsaDecryptor, r *http.Request) (
	*ContentSecurityHeader, error) {
	contentSecurity := r.Header.Get(httpx.ContentSecurity)
	attrs := httpx.ParseHeader(contentSecurity)
	fingerprint := attrs[httpx.KeyField]
	secret := attrs[httpx.SecretField]
	signature := attrs[signatureField]

	if len(fingerprint) == 0 || len(secret) == 0 || len(signature) == 0 {
		return nil, ErrInvalidHeader
	}

	decrypter, ok := decrypters[fingerprint]
	if !ok {
		return nil, ErrInvalidPublicKey
	}

	decryptedSecret, err := decrypter.DecryptBase64(secret)
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

// VerifySignature 校验指定请求的签名。
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

	passed := securityHeader.Signature == actualSignature
	if !passed {
		logx.Infof("签名被篡改, 期望：%s，实际：%s",
			securityHeader.Signature, actualSignature)
	}

	if passed {
		return httpx.CodeSignaturePass
	}

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
