package iox

import (
	"bytes"
	"io"
	"os"
)

const bufSize = 32 * 1024

// CountLines 返回文件行数。
func CountLines(file string) (int, error) {
	f, err := os.Open(file)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	var noEOL bool
	buf := make([]byte, bufSize)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := f.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			if noEOL {
				count++
			}
			return count, nil
		case err != nil:
			return count, err
		}

		noEOL = c > 0 && buf[c-1] != '\n'
	}
}
