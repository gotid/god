package codec

import (
	"bytes"
	"compress/gzip"
	"io"
)

const unzipLimit = 100 * 1024 * 1024 // 100MB

// Gzip 压缩字节数组。
func Gzip(bs []byte) []byte {
	var b bytes.Buffer

	w := gzip.NewWriter(&b)
	w.Write(bs)
	w.Close()

	return b.Bytes()
}

// Gunzip 解压缩字节数组。
func Gunzip(bs []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewBuffer(bs))
	defer r.Close()
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	if _, err = io.Copy(&b, io.LimitReader(r, unzipLimit)); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
