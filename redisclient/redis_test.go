package redisclient_test

import (
	"os"
	"testing"
	"time"

	"github.com/hokonco/kitgo"
	"github.com/hokonco/kitgo/redisclient"
	. "github.com/onsi/gomega"
)

func TestMain(m *testing.M) { os.Exit(kitgo.ShouldCover(m.Run(), 1.00)) }

func Test_client_redis(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect

	redisclient.New(redisclient.Config{
		MinRetryBackoff:    "30s",
		MaxRetryBackoff:    "30s",
		DialTimeout:        "30s",
		ReadTimeout:        "30s",
		WriteTimeout:       "30s",
		MaxConnAge:         "30s",
		PoolTimeout:        "30s",
		IdleTimeout:        "30s",
		IdleCheckFrequency: "30s",
	})

	redisCli, mock := redisclient.Test()
	defer mock.Close()

	str, err := redisCli.Get("key").Result()
	Expect(err).NotTo(BeNil())
	Expect(err.Error()).To(Equal("redis: nil"))
	Expect(str).To(BeEmpty())

	err = mock.Set("key", "val")
	Expect(err).To(BeNil())
	str, err = redisCli.Get("key").Result()
	Expect(err).To(BeNil())
	Expect(str).To(Equal("val"))

	mock.SetTTL("key", time.Minute)
	mock.FastForward(time.Hour)
	str, err = redisCli.Get("key").Result()
	Expect(err).NotTo(BeNil())
	Expect(err.Error()).To(Equal("redis: nil"))
	Expect(str).To(BeEmpty())

	pipe := redisCli.Pipeline()
	str, err = pipe.Ping().Result()
	Expect(err).To(BeNil())
	Expect(str).To(Equal(""))
	str, err = pipe.Ping().Result()
	Expect(err).To(BeNil())
	Expect(str).To(Equal(""))
	str, err = pipe.Ping().Result()
	Expect(err).To(BeNil())
	Expect(str).To(Equal(""))
	cmds, err := pipe.Exec()
	Expect(err).To(BeNil())
	Expect(len(cmds)).To(Equal(3))
}
