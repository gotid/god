syntax = "proto3";

package {{.package}};
option go_package="./{{.package}}";

message Request {
    string ping = 1;
}

message Response {
    string pong = 2;
}

service {{.serviceName}} {
    rpc Ping(Request) returns(Response);
}