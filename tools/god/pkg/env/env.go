package env

import (
	"github.com/gotid/god/tools/god/internal/version"
	sortedmap "github.com/gotid/god/tools/god/pkg/collection"
	"github.com/gotid/god/tools/god/pkg/protoc"
	"github.com/gotid/god/tools/god/pkg/protocgengo"
	"github.com/gotid/god/tools/god/pkg/protocgengogrpc"
	"github.com/gotid/god/tools/god/util/pathx"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	GodOS      = "GOD_OS"
	GodArch    = "GOD_ARCH"
	GodHome    = "GOD_HOME"
	GodDebug   = "GOD_DEBUG"
	GodCache   = "GOD_CACHE"
	GodVersion = "GOD_VERSION"

	ProtocVersion          = "PROTOC_VERSION"
	ProtocGenGoVersion     = "PROTOC_GEN_GO_VERSION"
	ProtocGenGoGrpcVersion = "PROTOC_GEN_GO_GRPC_VERSION"

	envFileDir = "env"
)

var godEnv *sortedmap.SortedMap

// 初始化 god 代码生成器的环境变量。
// * 变量是按顺序设置的，请勿改变。
func init() {
	godEnv = sortedmap.New()
	godEnv.SetKV(GodOS, runtime.GOOS)
	godEnv.SetKV(GodArch, runtime.GOARCH)

	defaultGodHome, err := pathx.GetDefaultGodHome()
	if err != nil {
		log.Fatalln(err)
	}

	existsEnv := readEnv(defaultGodHome)
	if existsEnv != nil {
		godHome, ok := existsEnv.GetString(GodHome)
		if ok && len(godHome) > 0 {
			godEnv.SetKV(GodHome, godHome)
		}
		if debug := existsEnv.GetOr(GodDebug, "").(string); debug != "" {
			if strings.EqualFold(debug, "true") || strings.EqualFold(debug, "false") {
				godEnv.SetKV(GodDebug, debug)
			}
		}
		if value := existsEnv.GetStringOr(GodCache, ""); value != "" {
			godEnv.SetKV(GodCache, value)
		}
	}

	if !godEnv.HasKey(GodHome) {
		godEnv.SetKV(GodHome, defaultGodHome)
	}
	if !godEnv.HasKey(GodDebug) {
		godEnv.SetKV(GodDebug, "false")
	}
	if !godEnv.HasKey(GodCache) {
		cacheDir, _ := pathx.GetCacheDir()
		godEnv.SetKV(GodCache, cacheDir)
	}
	godEnv.SetKV(GodVersion, version.BuildVersion)

	protocVer, _ := protoc.Version()
	godEnv.SetKV(ProtocVersion, protocVer)

	protocGenGoVer, _ := protocgengo.Version()
	godEnv.SetKV(ProtocGenGoVersion, protocGenGoVer)

	protocGenGoGrpcVer, _ := protocgengogrpc.Version()
	godEnv.SetKV(ProtocGenGoGrpcVersion, protocGenGoGrpcVer)

}

func readEnv(godHome string) *sortedmap.SortedMap {
	envFile := filepath.Join(godHome, envFileDir)
	data, err := os.ReadFile(envFile)
	if err != nil {
		return nil
	}
	dataStr := string(data)
	lines := strings.Split(dataStr, "\n")
	sm := sortedmap.New()
	for _, line := range lines {
		_, _, err = sm.SetExpression(line)
		if err != nil {
			continue
		}
	}
	return sm
}

func Print() string {
	return strings.Join(godEnv.Format(), "\n")
}

func Get(key string) string {
	return GetOr(key, "")
}

func GetOr(key, dft string) string {
	return godEnv.GetStringOr(key, dft)
}
