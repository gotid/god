package trace

import "net/http"

// TraceIdKey 是跟踪标识头。
// https://www.w3.org/TR/trace-context/#trace-id
// 以后可能更改为 trace-id。
var TraceIdKey = http.CanonicalHeaderKey("x-trace-id")
