package golang

import (
	"go/build"
	"os"
	"path/filepath"

	"github.com/gotid/god/tools/god/util/pathx"
)

// GoBin 返回 GOBIN 的路径。
func GoBin() string {
	defaultCtx := build.Default
	goPath := os.Getenv("GOPATH")
	goBin := filepath.Join(goPath, "bin")
	if !pathx.FileExists(goBin) {
		goRoot := os.Getenv("GOROOT")
		goBin = filepath.Join(goRoot, "bin")
	}
	if !pathx.FileExists(goBin) {
		goBin = os.Getenv("GOBIN")
	}
	if !pathx.FileExists(goBin) {
		goBin = filepath.Join(defaultCtx.GOPATH, "bin")
	}

	return goBin
}
