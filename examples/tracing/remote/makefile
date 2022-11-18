new-proto:
	god rpc -o user.proto
	god rpc -o portal.proto

protoc:
	protoc -I=. ./user.proto --go_out=./ --go-grpc_out=./
	protoc -I=. ./portal.proto --go_out=./ --go-grpc_out=./