package common

type Logger interface {
	Info(v ...any)
	Error(v ...any)
}
