package trace

import "net/http"

// TraceIdKey 表示一个跟踪ID标头。
// https://www.w3.org/TR/trace-context/#trace-id
var TraceIdKey = http.CanonicalHeaderKey("trace-id")
