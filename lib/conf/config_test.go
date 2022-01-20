package conf

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/gotid/god/lib/hash"
	"github.com/stretchr/testify/assert"
)

func TestLoadJsonConfig(t *testing.T) {
	exts := []string{
		".json",
		".yaml",
		".yml",
	}
	text := `{
		"a": "hi",
		"b": 1
	}`

	for _, ext := range exts {
		ext := ext
		t.Run(ext, func(t *testing.T) {
			t.Parallel()

			tempFile, err := createTempFile(ext, text)
			assert.Nil(t, err)
			defer os.Remove(tempFile)

			var val struct {
				A string `json:"a"`
				B int    `json:"b"`
			}
			MustLoad(tempFile, &val)
			assert.Equal(t, "hi", val.A)
			assert.Equal(t, 1, val.B)
		})
	}
}

func createTempFile(ext, text string) (string, error) {
	tempFile, err := ioutil.TempFile(os.TempDir(), hash.Md5Hex([]byte(text))+"*"+ext)
	if err != nil {
		return "", err
	}

	if err = ioutil.WriteFile(tempFile.Name(), []byte(text), os.ModeTemporary); err != nil {
		return "", err
	}

	filename := tempFile.Name()
	if err = tempFile.Close(); err != nil {
		return "", err
	}

	return filename, nil
}
