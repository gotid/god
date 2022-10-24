package env

import (
	"github.com/gotid/god/tools/god/vars"
	"os/exec"
	"runtime"
	"strings"
)

const (
	bin                = "bin"
	binGo              = "go"
	binProtoc          = "protoc"
	binProtocGenGo     = "protoc-gen-go"
	binProtocGenGoGrpc = "protoc-gen-go-grpc"
	cstOffset          = 60 * 60 * 8 // 中国标准时间偏差值为8
)

// LookupProtoc 在环境变量中查找 protoc 的可执行文件路径。
func LookupProtoc() (string, error) {
	suffix := getExeSuffix()
	xBin := binProtoc + suffix
	return LookPath(xBin)
}

// LookupProtocGenGo 在环境变量中查找 protoc-gen-go 的可执行文件路径。
func LookupProtocGenGo() (string, error) {
	suffix := getExeSuffix()
	xBin := binProtocGenGo + suffix
	return LookPath(xBin)
}

// LookupProtocGenGoGrpc 在环境变量中查找 protoc-gen-go-grpc 的可执行文件路径。
func LookupProtocGenGoGrpc() (string, error) {
	suffix := getExeSuffix()
	xBin := binProtocGenGoGrpc + suffix
	return LookPath(xBin)
}

// LookPath 在环境变量中查找给定的可执行文件路径。
func LookPath(xBin string) (string, error) {
	suffix := getExeSuffix()
	if len(suffix) > 0 && !strings.HasSuffix(xBin, suffix) {
		xBin = xBin + suffix
	}

	xPath, err := exec.LookPath(xBin)
	if err != nil {
		return "", err
	}

	return xPath, nil
}

// CanExec 判断当前系统能否使用 os.StartProcess 或（更常见）exec.Command 开启新进程。
func CanExec() bool {
	switch runtime.GOOS {
	case vars.OsJs, vars.OsIOS:
		return false
	default:
		return true
	}
}

func getExeSuffix() string {
	if runtime.GOOS == vars.OsWindows {
		return ".exe"
	}
	return ""
}
