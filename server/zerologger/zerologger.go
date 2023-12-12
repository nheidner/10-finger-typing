package zerologger

import (
	"10-typing/common"
	"fmt"
	"io"
	"time"

	"github.com/rs/zerolog"
)

type Zerologger struct {
	logger zerolog.Logger
}

func New(writer io.Writer) common.Logger {
	logger := zerolog.New(writer).With().Timestamp().Logger()

	return &Zerologger{logger}
}

func (zl *Zerologger) Info(v ...any) {
	zl.logger.Info().Msg(fmt.Sprint(v...))
}

func (zl *Zerologger) Error(v ...any) {
	zl.logger.Error().Msg(fmt.Sprint(v...))
}

func (zl *Zerologger) RequestInfo(method, path, clientIP string, statusCode int, latency time.Duration) {
	zl.logger.Info().
		Str("method", method).
		Str("path", path).
		Int("status", statusCode).
		Str("ip", clientIP).
		Dur("latency", latency).
		Msg("")
}
