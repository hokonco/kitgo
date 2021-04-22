package kitgo_test

import (
	"context"
	"testing"

	"github.com/hokonco/kitgo"
	. "github.com/onsi/gomega"
)

func Test_client_redis(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect

	kitgo.Redis.New(&kitgo.RedisConfig{
		Addrs: []string{},
	})

	ctx := context.Background()
	wrap, mock := kitgo.Redis.Test()
	defer Expect(mock.ExpectationsWereMet()).To(BeNil())

	const (
		key = "key"
		val = "val"
		ok  = "ok"
	)

	mock.ExpectGet(key).RedisNil()
	str, err := wrap.Get(ctx, key).Result()
	Expect(err).NotTo(BeNil())
	Expect(err.Error()).To(Equal("redis: nil"))
	Expect(str).To(BeEmpty())

	mock.ExpectGet(key).SetVal(val)
	str, err = wrap.Get(ctx, key).Result()
	Expect(err).To(BeNil())
	Expect(str).To(Equal(val))

	mock.ExpectGet(key).RedisNil()
	str, err = wrap.Get(ctx, key).Result()
	Expect(err).NotTo(BeNil())
	Expect(err.Error()).To(Equal("redis: nil"))
	Expect(str).To(BeEmpty())

	pipe := wrap.Pipeline()
	mock.ExpectPing().SetVal(ok)
	str, err = pipe.Ping(ctx).Result()
	Expect(err).To(BeNil())
	Expect(str).To(Equal(""))
	mock.ExpectPing().SetVal(ok)
	str, err = pipe.Ping(ctx).Result()
	Expect(err).To(BeNil())
	Expect(str).To(Equal(""))
	mock.ExpectPing().SetVal(ok)
	str, err = pipe.Ping(ctx).Result()
	Expect(err).To(BeNil())
	Expect(str).To(Equal(""))
	cmds, err := pipe.Exec(ctx)
	Expect(err).To(BeNil())
	Expect(len(cmds)).To(Equal(3))
}
