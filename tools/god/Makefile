build:
	go build -ldflags="-s -w" god.go
	mv god ~/go/bin/

mac:
	GOOS=darwin go build -ldflags="-s -w" -o god-darwin god.go
	$(if $(shell command -v upx), upx god-darwin)

win:
	GOOS=windows go build -ldflags="-s -w" -o god.exe god.go
	$(if $(shell command -v upx), upx god.exe)

linux:
	GOOS=linux go build -ldflags="-s -w" -o god-linux god.go
	$(if $(shell command -v upx), upx god-linux)
