package common

import (
	"time"
)

type Logger interface {
	Info(v ...any)
	Error(v ...any)
	RequestInfo(method, path, clientIP string, statusCode int, latency time.Duration)
}
