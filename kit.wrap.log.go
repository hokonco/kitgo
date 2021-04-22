package kitgo

import (
	"io"
	"log"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

var Log log_

type log_ struct{}

type LogConfig struct {
	Prefix string
	Flag   int

	multiWriter   []io.Writer
	consoleWriter *ConsoleWriter
}

type ConsoleWriter = zerolog.ConsoleWriter

func (c *LogConfig) ConsoleWriter(cw *ConsoleWriter) *LogConfig { c.consoleWriter = cw; return c }
func (c *LogConfig) MultiWriter(w ...io.Writer) *LogConfig      { c.multiWriter = w; return c }

func (log_) New(conf *LogConfig) *LogWrapper {
	if conf == nil {
		conf = &LogConfig{}
	}
	var ws []io.Writer
	if cw := conf.consoleWriter; cw != nil {
		if cw.Out == nil {
			cw.Out = io.Discard
		}
		conf.multiWriter = append(conf.multiWriter, *cw)
	}
	for i := 0; i < len(conf.multiWriter); i++ {
		if conf.multiWriter[i] == nil {
			continue
		}
		ws = append(ws, conf.multiWriter[i])
	}
	var z zerolog.Logger
	switch len(ws) {
	case 0:
		z = zerolog.Nop()
	case 1:
		z = zerolog.New(ws[0])
	default:
		z = zerolog.New(zerolog.MultiLevelWriter(ws...))
	}
	z = z.With().Timestamp().Stack().Logger()
	return &LogWrapper{log.New(&z, conf.Prefix, conf.Flag), &z}
}

// LogWrapper implement all the methods of *log.Logger
// also accessible from .Logger
//
// LogWrapper also have extension via .Z to access more
// robust logging experience using zerolog
type LogWrapper struct {
	// Logger is a default *log.Logger instance
	*log.Logger
	z *zerolog.Logger
}

func (x *LogWrapper) UseErrorStackMarshaler(use bool) *LogWrapper {
	if use && zerolog.ErrorStackMarshaler == nil {
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	} else if !use && zerolog.ErrorStackMarshaler != nil {
		zerolog.ErrorStackMarshaler = nil
	}
	return x
}
func (x *LogWrapper) Z(levelStr string) *zerolog.Event {
	level := zerolog.Disabled
	if lvl, err := zerolog.ParseLevel(strings.ToLower(levelStr)); err == nil {
		level = lvl
	}
	zLevel := x.z.Level(level)
	x.z = &(zLevel)
	return x.z.WithLevel(level)
}
