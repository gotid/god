version := $(shell /bin/date "+%Y-%m-%d %H:%M")

mac:
	GOOS=darwin go build -ldflags="-s -w" -ldflags="-X 'main.BuildTime=$(version)'" -o trans main.go
	#$(if $(shell command -v upx), upx trans)
	mv trans ~/go/bin/


linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -ldflags="-X 'main.BuildTime=$(version)'" -o trans_linux main.go
#	$(if $(shell command -v upx), upx trans)
	scp trans_linux root@gpu1:/root/test