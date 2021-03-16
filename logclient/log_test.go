package logclient_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/hokonco/kitgo"
	"github.com/hokonco/kitgo/logclient"
	. "github.com/onsi/gomega"
)

func TestMain(m *testing.M) { os.Exit(kitgo.ShouldCover(m.Run(), 1.00)) }

func Test_client_log(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect

	dir := t.TempDir()
	fs, err := ioutil.TempFile(dir, "the.log")
	Expect(err).To(BeNil())
	buf, msg, f := &bytes.Buffer{}, "MESSAGE", fmt.Sprintf
	now := func() string { return time.Now().Format(time.RFC3339) }

	logclient.New(logclient.Config{}).Z("")
	logclient.New(logclient.Config{FilePath: fs.Name()}).Z("trace")
	logclient.New(logclient.Config{ConsoleUse: true}).Z("trace")
	logclient.New(logclient.Config{}).Z("info")
	logclient.New(logclient.Config{}).Z("warn")
	logclient.New(logclient.Config{}).Z("error")
	logclient.New(logclient.Config{}).Z("fatal")
	logclient.New(logclient.Config{}).Z("panic")
	logclient.New(logclient.Config{}).Z("no_level")
	logclient.New(logclient.Config{}).Z("disable")
	logclient.New(logclient.Config{}).Z("debug")
	logclient.New(logclient.Config{}).Z("trace")

	logCli := logclient.New(logclient.Config{}, buf)
	buf.Truncate(0)
	logCli.Z("debug").Msg(msg)
	Expect(buf.String()).To(Equal(f(`{"level":"debug","time":"%s","message":"%s"}%s`, now(), msg, "\n")))

	buf.Truncate(0)
	logCli.Print(msg)
	Expect(buf.String()).To(Equal(f(`{"time":"%s","message":"%s"}%s`, now(), msg, "\n")))

	logCli = logclient.New(logclient.Config{}, buf)
	buf.Truncate(0)
	logCli.Z("warn").Msg(msg)
	Expect(buf.String()).To(Equal(f(`{"level":"warn","time":"%s","message":"%s"}%s`, now(), msg, "\n")))

	buf.Truncate(0)
	logCli.Print(msg)
	Expect(buf.String()).To(Equal(f(`{"time":"%s","message":"%s"}%s`, now(), msg, "\n")))
}
