package iox

import (
	"bytes"
	"github.com/gotid/god/lib/fs"
	"github.com/gotid/god/lib/stringx"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

func TestReadText(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{
			input:  `a`,
			expect: `a`,
		}, {
			input: `a
`,
			expect: `a`,
		}, {
			input: `a
b`,
			expect: `a
b`,
		}, {
			input: `a
b
`,
			expect: `a
b`,
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			filename, err := fs.TempFilenameWithText(test.input)
			assert.Nil(t, err)
			defer os.Remove(filename)

			content, err := ReadText(filename)
			assert.Nil(t, err)
			assert.Equal(t, test.expect, content)
		})
	}
}

func TestReadTextLines(t *testing.T) {
	text := `1

    2

    #a
    3`

	filename, err := fs.TempFilenameWithText(text)
	assert.Nil(t, err)
	defer os.Remove(filename)

	tests := []struct {
		options     []TextReadOption
		expectLines int
	}{
		{
			nil,
			6,
		}, {
			[]TextReadOption{KeepSpace(), OmitWithPrefix("#")},
			6,
		}, {
			[]TextReadOption{WithoutBlank()},
			4,
		}, {
			[]TextReadOption{OmitWithPrefix("#")},
			5,
		}, {
			[]TextReadOption{WithoutBlank(), OmitWithPrefix("#")},
			3,
		},
	}

	for _, test := range tests {
		t.Run(stringx.Rand(), func(t *testing.T) {
			lines, err := ReadTextLines(filename, test.options...)
			assert.Nil(t, err)
			assert.Equal(t, test.expectLines, len(lines))
		})
	}
}

func TestDupReadCloser(t *testing.T) {
	input := "hello"
	reader := io.NopCloser(bytes.NewBufferString(input))
	r1, r2 := DupReadCloser(reader)
	verify := func(r io.Reader) {
		output, err := io.ReadAll(r)
		assert.Nil(t, err)
		assert.Equal(t, input, string(output))
	}

	verify(r1)
	verify(r2)
}
