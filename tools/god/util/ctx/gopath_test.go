package ctx

import (
	"go/build"
	"os"
	"path/filepath"
	"testing"

	"github.com/gotid/god/lib/fs"

	"github.com/gotid/god/lib/stringx"
	"github.com/stretchr/testify/assert"
)

func TestProjectFromGoPath(t *testing.T) {
	dft := build.Default
	gp := dft.GOPATH
	if len(gp) == 0 {
		return
	}
	projectName := stringx.Rand()
	dir := filepath.Join(gp, "src", projectName)
	err := fs.MkdirIfNotExist(dir)
	if err != nil {
		return
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	ctx, err := projectFromGoPath(dir)
	assert.Nil(t, err)
	assert.Equal(t, dir, ctx.Dir)
	assert.Equal(t, projectName, ctx.Path)
}

func TestProjectFromGoPathNotInGoSrc(t *testing.T) {
	dft := build.Default
	gp := dft.GOPATH
	if len(gp) == 0 {
		return
	}
	projectName := stringx.Rand()
	dir := filepath.Join(gp, "src", projectName)
	err := fs.MkdirIfNotExist(dir)
	if err != nil {
		return
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	_, err = projectFromGoPath("testPath")
	assert.NotNil(t, err)
}
