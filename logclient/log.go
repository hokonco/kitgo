package logclient

import (
	"io"
	"log"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

type Config struct {
	ConsoleUse        bool   `yaml:"console_use" json:"console_use"`
	ConsoleNoColor    bool   `yaml:"console_no_color" json:"console_no_color"`
	ConsoleTimeFormat string `yaml:"console_time_format" json:"console_time_format"`
	FilePath          string `yaml:"file_path" json:"file_path"`
	Prefix            string `yaml:"prefix" json:"prefix"`
	Flag              int    `yaml:"flag" json:"flag"`
}

func New(conf Config, writers ...io.Writer) *Client {
	flag, perm := os.O_RDWR|os.O_CREATE|os.O_APPEND, os.FileMode(0666)
	if f, err := os.OpenFile(conf.FilePath, flag, perm); err == nil && f != nil {
		writers = append(writers, f)
	}
	if conf.ConsoleUse {
		writers = append(writers, zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
			w.NoColor = conf.ConsoleNoColor
			w.TimeFormat = conf.ConsoleTimeFormat
		}))
	}
	if len(writers) < 1 {
		writers = append(writers, zerolog.Nop())
	}
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	z := zerolog.New(zerolog.MultiLevelWriter(writers...)).With().Timestamp().Stack().Logger()
	return &Client{log.New(&z, conf.Prefix, conf.Flag), &z}
}

// Client implement all the methods of *log.Logger
// also accessible from .Logger
//
// Client also have extension via .Z to access more
// robust logging experience using zerolog
type Client struct {
	// Logger is a default *log.Logger instance
	*log.Logger

	z *zerolog.Logger
}

func (x *Client) Z(levelStr string) *zerolog.Event {
	levelStr, level := strings.ToLower(levelStr), zerolog.Disabled
	if lvl, err := zerolog.ParseLevel(levelStr); err == nil {
		level = lvl
	}
	newZ := x.z.Level(level)
	x.z = &newZ
	return x.z.WithLevel(level)
}
