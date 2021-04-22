package kitgo_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/hokonco/kitgo"
	. "github.com/onsi/gomega"
)

func Test_client_log(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect

	buf, buf2, msg := new(bytes.Buffer), new(bytes.Buffer), "MESSAGE"
	now := func() string { return time.Now().Format(time.RFC3339) }

	kitgo.Log.New(&kitgo.LogConfig{}).Z("")
	kitgo.Log.New(&kitgo.LogConfig{}).Z("info")
	kitgo.Log.New(&kitgo.LogConfig{}).Z("warn")
	kitgo.Log.New(&kitgo.LogConfig{}).Z("error")
	kitgo.Log.New(&kitgo.LogConfig{}).Z("fatal")
	kitgo.Log.New(&kitgo.LogConfig{}).Z("panic")
	kitgo.Log.New(&kitgo.LogConfig{}).Z("no_level")
	kitgo.Log.New(&kitgo.LogConfig{}).Z("disable")
	kitgo.Log.New(&kitgo.LogConfig{}).Z("debug")
	kitgo.Log.New(&kitgo.LogConfig{}).Z("trace")

	_ = kitgo.Log.New(nil).
		UseErrorStackMarshaler(true).
		UseErrorStackMarshaler(false)
	wrap := kitgo.Log.New(new(kitgo.LogConfig).MultiWriter(buf, buf2, nil).ConsoleWriter(&kitgo.ConsoleWriter{}))
	buf.Truncate(0)
	wrap.Z("debug").Msg(msg)
	Expect(buf.String()).To(Equal(fmt.Sprintf(`{"level":"debug","time":"%s","message":"%s"}%s`, now(), msg, "\n")))

	buf.Truncate(0)
	wrap.Print(msg)
	Expect(buf.String()).To(Equal(fmt.Sprintf(`{"time":"%s","message":"%s"}%s`, now(), msg, "\n")))

	wrap = kitgo.Log.New(new(kitgo.LogConfig).MultiWriter(buf))
	buf.Truncate(0)
	wrap.Z("warn").Msg(msg)
	Expect(buf.String()).To(Equal(fmt.Sprintf(`{"level":"warn","time":"%s","message":"%s"}%s`, now(), msg, "\n")))

	buf.Truncate(0)
	wrap.Print(msg)
	Expect(buf.String()).To(Equal(fmt.Sprintf(`{"time":"%s","message":"%s"}%s`, now(), msg, "\n")))
}
