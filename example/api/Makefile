version := $(shell /bin/date "+%Y-%m-%d %H:%M")

linux:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -ldflags="-X 'main.BuildTime=$(version)'" -o tuji-api main.go
	command -v upx &> /dev/null && upx tuji-api
	scp tuji-api root@tj:/src/tuji/
	#scp config.yaml root@tj:/etc/nginx/go/
	rm -rf tuji-api