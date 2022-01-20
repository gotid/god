package conf

import (
	"fmt"
	"os"
	"testing"

	"github.com/gotid/god/lib/fs"
	"github.com/stretchr/testify/assert"
)

func TestLoadProperties(t *testing.T) {
	text := `app.name = test
	app.program = app
	
	# 这是注释
	app.threads = 5`
	filename, err := fs.TempFilenameWithText(text)
	assert.Nil(t, err)
	defer os.Remove(filename)

	props, err := LoadProperties(filename)
	assert.Nil(t, err)
	assert.Equal(t, "test", props.GetString("app.name"))
	assert.Equal(t, "app", props.GetString("app.program"))
	assert.Equal(t, 5, props.GetInt("app.threads"))
	fmt.Println(props)
}

func TestMapBasedProperties_SetString(t *testing.T) {
	props := NewProperties()
	props.SetString("a", "the value of a")
	assert.Equal(t, "the value of a", props.GetString("a"))
}

func TestMapBasedProperties_SetInt(t *testing.T) {
	props := NewProperties()
	props.SetInt("a", 101)
	assert.Equal(t, 101, props.GetInt("a"))
}

func TestLoadBadFile(t *testing.T) {
	_, err := LoadProperties("no-such-file")
	assert.NotNil(t, err)
}
