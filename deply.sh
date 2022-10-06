linux:
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ./stat.go
  scp ./stat root@server4:/root/test