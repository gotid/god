package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/gotid/god/api/httpx"
	"github.com/gotid/god/examples/http/signature/internal"
	"github.com/gotid/god/lib/codec"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var crypt = flag.Bool("crypt", true, "是否加密正文")

func main() {
	flag.Parse()

	var err error
	body := "你好，世界！"

	// 正文 → 密文
	if *crypt {
		bodyBytes, err := codec.EcbEncrypt(internal.Key, []byte(body))
		if err != nil {
			log.Fatal(err)
		}
		body = base64.StdEncoding.EncodeToString(bodyBytes)
	}
	fmt.Println("请求体密文", body)

	// 构建给定密文的 POST 请求
	req, err := http.NewRequest(http.MethodPost, "http://localhost:3333/a/b?c=first&d=second", strings.NewReader(body))
	if err != nil {
		log.Fatal(err)
	}

	// 密文 → 签名
	sha := sha256.New()
	sha.Write([]byte(body))
	bodySign := fmt.Sprintf("%x", sha.Sum(nil))
	timestamp := time.Now().Unix()
	reqSign := strings.Join([]string{
		strconv.FormatInt(timestamp, 10),
		http.MethodPost,
		req.URL.Path,
		req.URL.RawQuery,
		bodySign,
	}, "\n")
	sign := codec.HmacBase64(internal.Key, reqSign)

	// 秘钥
	mode := "0"
	if *crypt {
		mode = "1"
	}
	content := strings.Join([]string{
		"version=v1",
		"type=" + mode,
		fmt.Sprintf("key=%s", base64.StdEncoding.EncodeToString(internal.Key)),
		"time=" + strconv.FormatInt(timestamp, 10),
	}, "; ")
	encryptor, err := codec.NewRsaEncryptor(internal.PubKey)
	if err != nil {
		log.Fatal(err)
	}
	output, err := encryptor.Encrypt([]byte(content))
	if err != nil {
		log.Fatal(err)
	}
	secret := base64.StdEncoding.EncodeToString(output)

	// 设置内容安全标头（fingerprint=?; secret=?; signature=?）
	contentSecurity := strings.Join([]string{
		fmt.Sprintf("fingerprint=%s", internal.Fingerprint),
		"secret=" + secret,
		"signature=" + sign,
	}, "; ")
	req.Header.Set(httpx.ContentSecurity, contentSecurity)
	req.Header.Set("Content-Type", "application/json")
	fmt.Println("请求体签名", httpx.ContentSecurity, ":")
	fmt.Println(fmt.Sprintf("fingerprint=%s", internal.Fingerprint))
	fmt.Println(fmt.Sprintf("secret=%s", secret))
	fmt.Println(fmt.Sprintf("signature=%s", sign))

	// 发起请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Println(resp.Status)

	// 解密响应
	var content2 []byte
	content2 = make([]byte, resp.ContentLength)
	_, err = io.ReadFull(resp.Body, content2)

	content2, err = base64.StdEncoding.DecodeString(string(content2))
	if err != nil {
		log.Fatal(err)
	}

	output, err = codec.EcbDecrypt(internal.Key, content2)
	if err != nil {
		log.Fatal(err)
	}

	var buf bytes.Buffer
	buf.Write(output)
	resp.Body = io.NopCloser(&buf)

	io.Copy(os.Stdout, resp.Body)
}
